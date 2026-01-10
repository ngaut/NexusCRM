#!/usr/bin/env python3
"""
Salesforce Data Import Tool for NexusCRM
=========================================
Production-grade import script with:
- Stratified sampling for type inference
- Self-healing schema correction
- Graceful shutdown handling
- Detailed progress and statistics
"""
import argparse
import csv
import http.client
import json
import logging
import os
import re
import signal
import sys
import time
import concurrent.futures
import threading
from dataclasses import dataclass, field
from typing import Dict, List, Optional, Set, Tuple, Any
from urllib.parse import urlparse

# ============================================================================
# Configuration Constants
# ============================================================================
TYPE_INFERENCE_SAMPLE_SIZE = int(os.environ.get('IMPORT_SAMPLE_SIZE', '5000'))
DEFAULT_CONCURRENCY = int(os.environ.get('IMPORT_CONCURRENCY', '10'))
DEFAULT_BATCH_SIZE = int(os.environ.get('IMPORT_BATCH_SIZE', '50'))
CHECKPOINT_INTERVAL = int(os.environ.get('IMPORT_CHECKPOINT_INTERVAL', '5000'))
REQUEST_TIMEOUT = int(os.environ.get('IMPORT_REQUEST_TIMEOUT', '30'))
MAX_RETRIES = int(os.environ.get('IMPORT_MAX_RETRIES', '3'))

# Text field heuristics - field names containing these are forced to LongTextArea
TEXT_FIELD_INDICATORS = frozenset([
    'name', 'title', 'desc', 'subject', 'note', 'comment', 'street', 'city',
    'state', 'zip', 'country', 'phone', 'email', 'url', 'link', 'status',
    'type', 'code', 'address', 'message', 'body', 'content', 'text', 'label'
])

# ============================================================================
# Logging Setup
# ============================================================================
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s | %(levelname)-5s | %(message)s',
    datefmt='%H:%M:%S'
)
logger = logging.getLogger('import_salesforce')

# Force unbuffered stdout for progress bars
sys.stdout.reconfigure(line_buffering=True)

# ============================================================================
# Statistics Tracking
# ============================================================================
@dataclass
class ImportStatistics:
    """Tracks import statistics for summary reporting."""
    start_time: float = field(default_factory=time.time)
    total_rows: int = 0
    success_count: int = 0
    error_count: int = 0
    auto_corrections: int = 0
    objects_processed: int = 0
    corrected_fields: Set[str] = field(default_factory=set)
    errors_by_type: Dict[str, int] = field(default_factory=dict)
    
    @property
    def elapsed_seconds(self) -> float:
        return time.time() - self.start_time
    
    @property
    def rows_per_second(self) -> float:
        elapsed = self.elapsed_seconds
        return self.total_rows / elapsed if elapsed > 0 else 0
    
    @property
    def success_rate(self) -> float:
        total = self.success_count + self.error_count
        return (self.success_count / total * 100) if total > 0 else 0
    
    def record_error(self, error_type: str):
        self.errors_by_type[error_type] = self.errors_by_type.get(error_type, 0) + 1
    
    def print_summary(self):
        """Print a detailed summary report."""
        print("\n" + "=" * 60)
        print("📊 IMPORT SUMMARY")
        print("=" * 60)
        print(f"  Duration:        {self.elapsed_seconds:.1f}s")
        print(f"  Objects:         {self.objects_processed}")
        print(f"  Total Rows:      {self.total_rows:,}")
        print(f"  Success:         {self.success_count:,} ({self.success_rate:.1f}%)")
        print(f"  Errors:          {self.error_count:,}")
        print(f"  Auto-Corrections:{self.auto_corrections}")
        print(f"  Throughput:      {self.rows_per_second:,.0f} rows/sec")
        if self.corrected_fields:
            print(f"\n  🔧 Corrected Fields:")
            for f in sorted(self.corrected_fields):
                print(f"     - {f}")
        if self.errors_by_type:
            print(f"\n  ❌ Errors by Type:")
            for t, c in sorted(self.errors_by_type.items(), key=lambda x: -x[1])[:5]:
                print(f"     - {t}: {c}")
        print("=" * 60)

# Global statistics and shutdown flag
stats = ImportStatistics()
shutdown_requested = False

def signal_handler(signum, frame):
    """Handle SIGINT/SIGTERM for graceful shutdown."""
    global shutdown_requested
    logger.warning("⚠️  Shutdown requested. Finishing current batch...")
    shutdown_requested = True

signal.signal(signal.SIGINT, signal_handler)
signal.signal(signal.SIGTERM, signal_handler)

def estimate_total_rows(file_path):
    try:
        f_size = os.path.getsize(file_path)
        if f_size == 0: return 0
        
        with open(file_path, 'r', encoding='utf-8-sig', errors='ignore') as f:
            # Read header first (skipping it for average calc)
            header = f.readline()
            header_bytes = len(header.encode('utf-8'))
            
            # Read sample lines to calculate average row length
            sample_bytes = 0
            sample_count = 0
            
            for _ in range(50):
                line = f.readline()
                if not line: break
                sample_bytes += len(line.encode('utf-8'))
                sample_count += 1
            
            if sample_count == 0:
                return 0
                
            avg_row_len = sample_bytes / sample_count
            if avg_row_len == 0: return 1
            
            # Estimate remaining rows
            remaining_bytes = f_size - header_bytes
            return int(remaining_bytes / avg_row_len)
    except Exception:
        return 0

class NexusClient:
    def __init__(self, base_url, token, debug=False):
        self.base_url = base_url
        self.token = token
        self.debug = debug
        parsed = urlparse(base_url)
        self.host = parsed.netloc
        self.scheme = parsed.scheme
        self._local = threading.local()

    def _get_connection(self):
        if not hasattr(self._local, 'conn'):
            if self.scheme == 'https':
                self._local.conn = http.client.HTTPSConnection(self.host, timeout=30)
            else:
                self._local.conn = http.client.HTTPConnection(self.host, timeout=30)
            if self.debug:
                 print(f"[DEBUG] Opened new connection for thread {threading.get_ident()}")
        return self._local.conn

    def _close_connection(self):
        if hasattr(self._local, 'conn'):
            try:
                self._local.conn.close()
            except Exception:
                pass
            del self._local.conn

    def request(self, method, path, body=None):
        retries = 3
        backoff = 1.0
        
        for attempt in range(retries + 1):
            try:
                conn = self._get_connection()
    
                headers = {
                    'Content-Type': 'application/json',
                    'Authorization': f'Bearer {self.token}',
                    'X-Source': 'import_salesforce_py',
                    'Connection': 'keep-alive'
                }
    
                json_body = json.dumps(body) if body is not None else None
                
                if self.debug:
                    print(f"[DEBUG] {method} {path} Body: {json_body}")
    
                conn.request(method, path, body=json_body, headers=headers)
                resp = conn.getresponse()
                resp_body = resp.read().decode('utf-8')
    
                if self.debug:
                    print(f"[DEBUG] Response {resp.status}: {resp_body}")
    
                # Retry on 500s or 503s?
                if resp.status >= 500:
                    if attempt < retries:
                        time.sleep(backoff)
                        backoff *= 2
                        self._close_connection() # Force reconnect on error
                        continue
                    return None, f"API Error {resp.status}: {resp_body}"

                if resp.status >= 400:
                    # 400 errors are usually logic errors, don't retry connection unless we think it's protocol related?
                    # Generally 400 is application error, so keep connection alive.
                    # But we'll read body and return.
                    return None, f"API Error {resp.status}: {resp_body}"
    
                try:
                    if resp_body:
                        data = json.loads(resp_body)
                        return data, None
                    return {}, None
                except json.JSONDecodeError:
                    return resp_body, None
    
            except Exception as e:
                # Retry connection errors
                self._close_connection()
                if attempt < retries:
                    time.sleep(backoff)
                    backoff *= 2
                    continue
                return None, str(e)
            # Do NOT close locally, let connection persist
        return None, "Max retries exceeded"

    def update_field_type(self, obj_name, field_name, new_type):
        """Update a field's type. Used for auto-schema-correction."""
        payload = {"type": new_type}
        _, err = self.request('PATCH', f'/api/metadata/objects/{obj_name}/fields/{field_name}', payload)
        return err is None

class SchemaManager:
    def __init__(self, client):
        self.client = client
        self.system_fields = {'id', 'created_at', 'updated_at', 'created_by', 'updated_by'}

    def infer_type(self, value, field_name=None):
        """Infer field type from value with name-based heuristics.
        
        ROBUST DESIGN: Hybrid Logic.
        1. Heuristics: If name implies Text (e.g. 'name', 'desc'), force LongTextArea.
        2. Data: If name is ambiguous, check sample values for patterns.
        
        Supported types: LongTextArea, Boolean, Number, DateTime, Email, Url
        """
        import re
        
        if not field_name:
             # Fallback if no name provided (shouldn't happen with updated caller)
             return 'LongTextArea'

        fn = field_name.lower()
        
        # 1. Name-based Heuristics (Strong Signal)
        # Force text for known text field patterns
        if any(indicator in fn for indicator in TEXT_FIELD_INDICATORS):
             return 'LongTextArea'
        
        # Empty value - default to text
        if not value or value.strip() == '':
            return 'LongTextArea'
        
        val = value.strip()
        
        # 2. Data-based Inference (Ordered by specificity)
        
        # Boolean detection
        if val.lower() in ('true', 'false', '1', '0', 'yes', 'no'):
            return 'Boolean'
        
        # DateTime detection (ISO format, Salesforce format)
        # Patterns: 2024-01-15T10:30:00Z, 2024-01-15 10:30:00, 2024-01-15
        datetime_patterns = [
            r'^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}',  # ISO with T
            r'^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}',  # With space
            r'^\d{4}-\d{2}-\d{2}$',                    # Date only
        ]
        for pattern in datetime_patterns:
            if re.match(pattern, val):
                return 'DateTime'
        
        # Number detection (int or float)
        # Be careful: don't match phone numbers, zip codes, IDs
        # Only match if it's purely numeric (with optional decimal and sign)
        if re.match(r'^-?\d+\.?\d*$', val) and not fn.endswith('_id') and 'phone' not in fn and 'zip' not in fn:
            # Check if it's a reasonable number (not a huge ID-like string)
            if len(val.replace('.', '').replace('-', '')) <= 15:
                return 'Number'
        
        # Email detection
        if re.match(r'^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$', val):
            return 'Email'
        
        # URL detection
        if re.match(r'^https?://', val):
            return 'Url'
        
        # Default: LongTextArea (safest fallback)
        return 'LongTextArea'

    def ensure_object(self, obj_name):
        print(f"Checking object: {obj_name}...")
        # Check if exists
        data, err = self.client.request('GET', f'/api/metadata/objects/{obj_name}')
        if not err:
            print(f"Object {obj_name} already exists.")
            return True

        # Create
        print(f"Creating object {obj_name}...")
        payload = {
            "api_name": obj_name,
            "label": obj_name.replace('_', ' ').title(),
            "description": f"Imported from Salesforce {obj_name}"
        }
        data, err = self.client.request('POST', '/api/metadata/objects', payload)
        if err:
            print(f"Failed to create object {obj_name}: {err}")
            return False
        return True

    def ensure_fields(self, obj_name, headers, sample_row):
        print(f"Ensuring fields for {obj_name}...")
        
        # Ensure 'original_id' field specifically for mapping
        self._ensure_field(obj_name, 'original_id', 'LongTextArea', 'Original Salesforce ID')
        
        # We need to list existing fields to avoid attempting re-creation
        # For now, we'll just try-create and ignore 400/409 errors (or check strictly if API supports listing fields)
        # Optimize: Get object definition first to see fields?
        
        # Prepare field definitions
        fields_to_create = []
        for header, value in zip(headers, sample_row):
            field_name = header.lower().strip()
            if field_name in self.system_fields or not field_name:
                continue
            if field_name == 'id':
                continue

            field_type = self.infer_type(value, header)
            fields_to_create.append((field_name, field_type, header))

        print(f"Creating {len(fields_to_create)} fields in parallel...")
        
        with concurrent.futures.ThreadPoolExecutor(max_workers=1 if self.client.debug else 20) as executor:
            futures = [
                executor.submit(self._ensure_field, obj_name, api, type_, label)
                for api, type_, label in fields_to_create
            ]
            concurrent.futures.wait(futures)

    def _ensure_field(self, obj_name, api_name, type_name, label):
        payload = {
            "api_name": api_name,
            "label": label,
            "type": type_name
        }
        # Fire and forget
        self.client.request('POST', f'/api/metadata/objects/{obj_name}/fields', payload)

    def write_stats(self, obj_name, total, success, failures):
        file_exists = os.path.exists("import_stats.csv")
        with open("import_stats.csv", "a") as f:
            if not file_exists:
                f.write("Object,Total,Success,Failures,Date\n")
            f.write(f"{obj_name},{total},{success},{failures},{time.strftime('%Y-%m-%d %H:%M:%S')}\n")

class CheckpointManager:
    def __init__(self, obj_name):
        self.filename = f".{obj_name}.checkpoint"
    
    def read(self):
        if os.path.exists(self.filename):
            try:
                with open(self.filename, 'r') as f:
                    return int(f.read().strip())
            except (ValueError, OSError):
                return 0
        return 0

    def write(self, count):
        with open(self.filename, 'w') as f:
            f.write(str(count))

    def clear(self):
        if os.path.exists(self.filename):
            os.remove(self.filename)

class Importer:
    def __init__(self, client, concurrency=10):
        self.client = client
        self.concurrency = concurrency
        self.corrected_fields = set()  # Track auto-corrected fields to avoid repeated API calls

    def process_batch(self, obj_name: str, headers: List[str], rows: List[List[str]], field_mapping: Optional[Dict[str, str]] = None) -> Tuple[int, int]:
        """Process a batch of rows using the Bulk Insert API.
        
        Returns:
            Tuple of (success_count, error_count)
        """
        success = 0
        error = 0
        if field_mapping is None:
            field_mapping = {}

        bulk_payload = []
        
        # 1. Prepare Data
        for row in rows:
            # Check for graceful shutdown (check per batch, not per row is fine)
            if shutdown_requested:
                logger.warning("Shutdown requested, stopping batch processing")
                break
                
            mapped_data = {}
            for h, v in zip(headers, row):
                raw_key = h.lower().strip()
                if not raw_key:
                    continue
                
                # Apply mapping
                key = field_mapping.get(raw_key, raw_key)
                mapped_data[key] = v
                    
            # Basic cleaning - remove empty values
            clean_data = {k: v for k, v in mapped_data.items() if v != ""}
            
            # Fallback: Generate name if missing (required field)
            if 'name' not in clean_data or not clean_data.get('name'):
                # Try to use original_id as fallback
                fallback = clean_data.get('original_id', '') or clean_data.get('subject', '') or f"Record-{len(self.corrected_fields)}"
                if fallback:
                    clean_data['name'] = str(fallback)[:255]
            
            bulk_payload.append(clean_data)

        if not bulk_payload or shutdown_requested:
            return 0, 0
            
        # 2. Bulk API Call
        # Endpoint: POST /api/data/:obj_name/bulk
        # Payload: { "records": [ ... ], "batch_size": 100 }
        payload_wrapper = {
            "records": bulk_payload,
            "batch_size": 100, # Default batch size for backend processing
            "skip_flows": False,
            "skip_auto_numbers": False
        }
        data, err = self.client.request('POST', f'/api/data/{obj_name}/bulk', payload_wrapper)
        
        if err:
            # Entire batch failed (network error or crash)
            logger.error(f"Bulk Batch Error: {err}")
            stats.record_error(self._extract_error_type(str(err)))
            return 0, len(bulk_payload)
        
        # 3. Process Result
        # Response might be wrapped in "data" envelope: { "data": { "success_count": ... } }
        result_data = data
        if "data" in data and isinstance(data["data"], dict):
            result_data = data["data"]

        s_count = result_data.get('success_count', 0)
        f_count = result_data.get('failed_count', 0)
        errors = result_data.get('errors', [])
        
        if f_count > 0:
            # First 10 errors
            for i, e_msg in enumerate(errors[:10]):
                logger.warning(f"Bulk Row Error: {e_msg}")
                self._handle_bulk_error(obj_name, e_msg)
            
            if f_count > 10:
                 logger.warning(f"... and {f_count - 10} more errors")

            # Update stats for errors
            for _ in range(f_count):
                 stats.record_error("bulk_validation_error")

        return s_count, f_count

    def _handle_bulk_error(self, obj_name, error_msg):
        """Attempt to handle specific bulk errors (like Schema Correction)."""
        # "record 5: validation error on field 'amount': expected number"
        type_error_match = re.search(r"validation error on field '([^']+)': expected (\w+)", error_msg)
        if type_error_match:
            bad_field = type_error_match.group(1)
            expected_type = type_error_match.group(2)
            if bad_field not in self.corrected_fields:
                logger.info(f"🔧 Auto-correcting field '{bad_field}' from {expected_type} to LongTextArea...")
                if self.client.update_field_type(obj_name, bad_field, 'LongTextArea'):
                    self.corrected_fields.add(bad_field)
                    stats.corrected_fields.add(f"{obj_name}.{bad_field}")
                    stats.auto_corrections += 1
                    # We can't easily retry just this row in bulk mode without re-architecting.
                    # For now, accept failure and rely on re-run or next batch.
                else:
                    logger.error(f"Failed to auto-correct field '{bad_field}'")

    
    def _extract_error_type(self, error_str: str) -> str:
        """Extract a short error type from an error message for categorization."""
        if 'expected boolean' in error_str:
            return 'type:boolean'
        elif 'expected number' in error_str:
            return 'type:number'
        elif 'Data too long' in error_str:
            return 'data_too_long'
        elif 'Unauthorized' in error_str:
            return 'auth_error'
        elif 'timeout' in error_str.lower():
            return 'timeout'
        else:
            return 'other'

    def _infer_type_from_samples(self, field_name: str, sample_values: List[str]) -> str:
        """
        Infer field type from sample values.
        
        Priority: Boolean > Number > DateTime > LongTextArea
        
        SAFE DESIGN: Only infer non-text types if ALL samples match the pattern.
        Any ambiguity -> default to LongTextArea (safest).
        """
        if not sample_values:
            return 'LongTextArea'
        
        fn = field_name.lower()
        
        # Name-based hints for Number fields (override data analysis)
        number_hints = ('amount', 'count', 'quantity', 'total', 'price', 'rate', 
                        'percent', 'probability', 'duration', 'age', 'size', 'score',
                        'latitude', 'longitude', 'distance', 'weight', 'height')
        if any(hint in fn for hint in number_hints):
            # Verify samples are actually numeric
            if self._all_numeric(sample_values):
                return 'Number'
        
        # Quick Boolean check (only true/false, NOT 1/0)
        if self._all_boolean(sample_values):
            return 'Boolean'
        
        # Number detection (for fields without name hints)
        if self._all_numeric(sample_values) and not fn.endswith('_id') and 'phone' not in fn:
            return 'Number'
        
        # DateTime detection
        if self._all_datetime(sample_values):
            return 'DateTime'
        
        # Default to safe text type
        return 'LongTextArea'
    
    def _all_boolean(self, values: List[str]) -> bool:
        """Check if all values are boolean strings."""
        for v in values:
            if v.lower() not in ('true', 'false'):
                return False
        return True
    
    def _all_numeric(self, values: List[str]) -> bool:
        """Check if all values are numeric (int or float, with optional sign)."""
        import re
        for v in values:
            # Allow empty strings (sparse data)
            if not v:
                continue
            # Match: -123, 123.45, .45, 123., but NOT phone-like patterns
            if not re.match(r'^-?\d*\.?\d+$', v):
                return False
            # Reject overly long numbers (likely IDs)
            if len(v.replace('.', '').replace('-', '')) > 15:
                return False
        return True
    
    def _all_datetime(self, values: List[str]) -> bool:
        """Check if all values match common datetime patterns."""
        import re
        patterns = [
            r'^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}',  # ISO with T
            r'^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}',  # With space
            r'^\d{4}-\d{2}-\d{2}$',                    # Date only
        ]
        for v in values:
            if not v:
                continue
            matched = False
            for pattern in patterns:
                if re.match(pattern, v):
                    matched = True
                    break
            if not matched:
                return False
        return True

    def run(self, file_path, obj_name):
        _, ext = os.path.splitext(file_path)
        if ext.lower() != '.csv':
            print("Only CSV files are supported.")
            return

        print(f"Reading {file_path}...")
        
        with open(file_path, 'r', encoding='utf-8-sig', errors='replace') as f:
            reader = csv.reader(f)
            try:
                headers = next(reader)
            except StopIteration:
                print("Empty file.")
                return

            # Read one row for type inference
            try:
                first_row = next(reader)
            except StopIteration:
                print("Header only, no data.")
                return
            
            # Inference & Schema components
            mgr = SchemaManager(self.client)
            if not mgr.ensure_object(obj_name):
                return

            # Stratified Sampling: Sample rows evenly across the file for better type inference
            print(f"Stratified sampling for type inference (~{TYPE_INFERENCE_SAMPLE_SIZE} samples)...")
            
            # Count lines cheaply
            with open(file_path, 'rb') as fb:
                total_lines = sum(1 for _ in fb)
            total_lines -= 1  # Exclude header
            
            if total_lines <= TYPE_INFERENCE_SAMPLE_SIZE:
                # Small file: sample everything
                sample_positions = set(range(total_lines))
            else:
                # Large file: sample at regular intervals
                interval = total_lines // TYPE_INFERENCE_SAMPLE_SIZE
                sample_positions = set(i * interval for i in range(TYPE_INFERENCE_SAMPLE_SIZE))
            
            sample_rows = [first_row]  # Always include first row
            current_pos = 1  # We already read first_row
            
            for row in reader:
                if current_pos in sample_positions:
                    sample_rows.append(row)
                current_pos += 1
                # Early exit once we have enough samples
                if len(sample_rows) >= TYPE_INFERENCE_SAMPLE_SIZE + 1:
                    break
            
            # Reset file position for actual data import
            f.seek(0)
            reader = csv.reader(f)
            next(reader)  # Skip header again
            
            # Analyze headers
            fields_to_create = []

            # Prepare Field Mapping
            header_map = {h.lower().strip(): h for h in headers}
            field_mapping = {
                "id": "original_id",
                "ownerid": "original_owner_id",
                "createdbyid": "original_created_by_id",
                "lastmodifiedbyid": "original_last_modified_by_id",
                "accountid": "original_account_id",
                "contactid": "original_contact_id",
            }
             # Smart Name Mapping
            if "name" not in header_map:
                candidates = ["name", "currency", "subject", "title", "casenumber", "ordernumber", "solutionname", "developername"]
                found_name = False
                for c in candidates:
                    if c in header_map:
                        field_mapping[c] = "name"
                        found_name = True
                        break
                
                # If still no name, use the first non-id column that isn't system
                if not found_name:
                    for h in headers:
                        h_lower = h.lower().strip()
                        if h_lower not in ("id", "ownerid", "systemmodstamp", "createddate", "createdbyid"):
                             field_mapping[h_lower] = "name"
                             break
            
            print(f"Field Mapping: {field_mapping}")

            for idx, header in enumerate(headers):
                field_name = header.lower().strip()
                if field_name in mgr.system_fields or not field_name:
                    continue
                if field_name == 'id':
                    continue
                
                # ==========================================
                # Type Inference: Name Heuristics > Data
                # ==========================================
                
                # 1. Name-based heuristics (highest priority)
                if any(indicator in field_name for indicator in TEXT_FIELD_INDICATORS):
                    final_type = 'LongTextArea'
                else:
                    # 2. Data-based inference (only for non-text-like names)
                    # Collect sample values for this field
                    sample_values = []
                    for row in sample_rows:
                        if idx < len(row) and row[idx]:
                            sample_values.append(row[idx].strip())
                    
                    if not sample_values:
                        final_type = 'LongTextArea'
                    else:
                        # Detect type from samples
                        final_type = self._infer_type_from_samples(field_name, sample_values)
                
                fields_to_create.append((field_name, final_type, header))


            print(f"Creating {len(fields_to_create)} fields in parallel...")
            with concurrent.futures.ThreadPoolExecutor(max_workers=1 if self.client.debug else 20) as executor:
                futures = [
                    executor.submit(mgr._ensure_field, obj_name, api, type_, label)
                    for api, type_, label in fields_to_create
                ]
                
                # Show progress for fields
                completed_fields = 0
                total_fields = len(futures)
                for f in concurrent.futures.as_completed(futures):
                    completed_fields += 1
                    if completed_fields % 10 == 0 or completed_fields == total_fields:
                         print(f"Creating fields... {completed_fields}/{total_fields}")

            # Process Data
            
            # Restartability
            ckpt = CheckpointManager(obj_name)
            start_row = ckpt.read()
            if start_row > 0:
                print(f"Resuming {obj_name} from row {start_row}...")

            # Progress estimation
            estimated_total = estimate_total_rows(file_path)
            
            # Start Processing
            total_success = 0
            total_errors = 0
            
            # Function to process a chunk
            def process_chunk(chunk):
                return self.process_batch(obj_name, headers, chunk, field_mapping)

            executor = concurrent.futures.ThreadPoolExecutor(max_workers=self.concurrency)
            futures = set()
            
            def submit_batch(batch):
                f = executor.submit(process_chunk, batch)
                futures.add(f)
            
            try:
                # We treat the sample_rows and reader as a unified stream of data rows
                # Row index 0 = first data row (header excluded)
                
                current_row_index = 0
                batch = []

                # 1. Process reader (which covers all rows since we seek(0))
                # sample_rows was only for inference and should not be re-processed separately
                
                # 2. Process reader
                for row in reader:
                    if current_row_index >= start_row:
                        batch.append(row)
                        if len(batch) >= 50:
                            submit_batch(batch)
                            batch = []
                    current_row_index += 1
                    
                    # Periodic checkpoint update (not perfect with threads, but approximation)
                    if current_row_index % 5000 == 0:
                         # Note: This is a "submitted" checkpoint, not "completed".
                         # On crash, some submitted batches may not have finished.
                         # Worst case: a few hundred rows may be re-imported as duplicates.
                         ckpt.write(current_row_index)

                if batch:
                    submit_batch(batch)
                    
                # Wait for completion
                total_processed_count = 0
                
                print("Importing data...")
                for future in concurrent.futures.as_completed(futures):
                    s, e = future.result()
                    total_success += s
                    total_errors += e
                    total_processed_count += (s + e)
                    
                    # Update global stats for summary
                    stats.success_count += s
                    stats.error_count += e
                    stats.total_rows += (s + e)
                    
                    if total_processed_count % 1000 == 0:
                         # Calculate percentage
                         current_total = total_processed_count + start_row
                         percent = 0
                         if estimated_total > 0:
                             percent = int((current_total / estimated_total) * 100)
                             if percent > 100: percent = 99
                         
                         # Visual Bar with ETA
                         bar_len = 30
                         filled_len = int(bar_len * percent / 100)
                         bar = '=' * filled_len + ' ' * (bar_len - filled_len)
                         
                         # Calculate ETA
                         elapsed = time.time() - stats.start_time
                         if current_total > 0 and percent > 0 and percent < 100:
                             eta_seconds = elapsed * (100 - percent) / percent
                             eta_str = f"{int(eta_seconds // 60)}m{int(eta_seconds % 60)}s"
                         else:
                             eta_str = "--"
                         
                         print(f"[{bar}] {percent:>3}% | {current_total:>8,} / ~{estimated_total:,} rows | ✅ {total_success:,} | ❌ {total_errors:,} | ETA: {eta_str}")

                print(f"Done. Success: {total_success}. Failures: {total_errors}.")
                ckpt.clear() # Done successfully
                mgr.write_stats(obj_name, total_processed_count + start_row, total_success, total_errors)
            finally:
                executor.shutdown(wait=False)

def main():
    parser = argparse.ArgumentParser(
        description='Salesforce Data Import Tool for NexusCRM',
        formatter_class=argparse.ArgumentDefaultsHelpFormatter
    )
    parser.add_argument('--file', required=True, help='Path to CSV file')
    parser.add_argument('--obj', required=True, help='Target Object Name (e.g. lead)')
    parser.add_argument('--url', default='http://localhost:3001', help='NexusCRM Base URL')
    parser.add_argument('--token', required=True, help='Authentication Token')
    parser.add_argument('--concurrency', type=int, default=DEFAULT_CONCURRENCY, help='Thread count')
    parser.add_argument('--debug', action='store_true', help='Enable debug logging')
    parser.add_argument('--dry-run', action='store_true', help='Analyze schema without importing data')

    args = parser.parse_args()
    
    if args.debug:
        logger.setLevel(logging.DEBUG)
    
    logger.info(f"Starting import: {args.obj} from {args.file}")
    logger.info(f"Config: concurrency={args.concurrency}, sample_size={TYPE_INFERENCE_SAMPLE_SIZE}")

    try:
        client = NexusClient(args.url, args.token, args.debug)
        importer = Importer(client, args.concurrency)
        importer.run(args.file, args.obj)
        
        # Update global stats
        stats.objects_processed += 1
        # total_rows is already updated inside importer.run() via the completion loop
        
    except Exception as e:
        logger.error(f"Import failed: {e}")
        raise
    finally:
        # Always print summary
        if not shutdown_requested:
            stats.print_summary()
        else:
            logger.warning("Import interrupted. Partial results:")
            stats.print_summary()

if __name__ == '__main__':
    main()
