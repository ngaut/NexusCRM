#!/bin/bash

# Change to script directory
cd "$(dirname "$0")"
SCRIPT_DIR="$(pwd)"

# Configuration
DATA_DIR="../../salesforce_data/data"
SCRIPT="./import_salesforce.py"
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyIjp7ImlkIjoiMTAwMDAwMDAtMDAwMC0wMDAwLTAwMDAtMDAwMDAwMDAwMDAxIiwibmFtZSI6IlN5c3RlbSBBZG1pbmlzdHJhdG9yIiwiZW1haWwiOiJhZG1pbkB0ZXN0LmNvbSIsInByb2ZpbGVfaWQiOiJzeXN0ZW1fYWRtaW4ifSwiZXhwIjoxNzY4MTA2NDE5LCJpYXQiOjE3NjgwMjAwMTksImp0aSI6IjFmM2ZiYzY2LWZmMTctNDZhNi1hOTg1LTRlZTg4MjRlNzVjOSJ9.YiZj626Fd8BoUQXKLeekHEe5ALWG09bWmn73fd0G6co"
CONCURRENCY="${1:-20}"
LOG_FILE="import_all.log"

echo "Starting Bulk Import at $(date)" > "$LOG_FILE"
echo "Concurrency: $CONCURRENCY threads" | tee -a "$LOG_FILE"

# Priority Objects (Order matters)
CORE_OBJECTS=("account" "lead" "contact" "opportunity" "task" "campaign")

processed_files=("user")
if [ -f "completed_objects.txt" ]; then
    processed_files=($(cat completed_objects.txt))
fi

run_import() {
    obj_name=$1
    file_path=$2

    echo "---------------------------------------------------" | tee -a "$LOG_FILE"
    echo "Importing $obj_name from $file_path..." | tee -a "$LOG_FILE"
    
    python3 "$SCRIPT" --file "$DATA_DIR/$obj_name.csv" --obj "$obj_name" --token "$TOKEN" --concurrency "$CONCURRENCY" 2>&1 | tee -a "$LOG_FILE"
    
    status=$?
    if [ $status -eq 0 ]; then
        echo "SUCCESS: $obj_name" | tee -a "$LOG_FILE"
        echo "$obj_name" >> completed_objects.txt
    else
        echo "FAILURE: $obj_name (Exit Code: $status)" | tee -a "$LOG_FILE"
    fi
}

# 1. Import Core Objects first
for obj in "${CORE_OBJECTS[@]}"; do
    file="$DATA_DIR/$obj.csv"
    if [ -f "$file" ]; then
        run_import "$obj" "$file"
        processed_files+=("$obj")
    else
        echo "Warning: Priority file $file not found." | tee -a "$LOG_FILE"
    fi
done

# 2. Import Everything Else
for file in "$DATA_DIR"/*.csv; do
    filename=$(basename -- "$file")
    obj_name="${filename%.*}"
    
    # Skip if processed
    if [[ " ${processed_files[*]} " == *" $obj_name "* ]]; then
        continue
    fi

    # Skip specific problematic or system files if needed
    if [[ "$obj_name" == *"fivetran"* ]]; then
        # Optional: Uncomment to skip fivetran audit tables if not needed
        # continue
        :
    fi

    run_import "$obj_name" "$file"
done

echo "Bulk Import Completed at $(date)" >> "$LOG_FILE"
