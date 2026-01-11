import csv
import os
import sys
from .utils import logger, stats, normalize_date, normalize_number, normalize_boolean
from .schema import SchemaManager

# Increase CSV field size limit to handle large formula fields (default is 128KB, set to 32MB)
csv.field_size_limit(32 * 1024 * 1024)

def process_file(file_path, obj_name, client, args, shutdown_signal=None):
    stats.objects_processed += 1
    
    with open(file_path, 'r', encoding='utf-8-sig', errors='replace') as f:
        reader = csv.reader(f)
        try:
            headers = next(reader)
            # Handle potential BOM leakage in first header
            if headers and headers[0].startswith('\ufeff'):
                headers[0] = headers[0][1:]
        except StopIteration:
            return

        # Check for duplicate normalized headers
        norm_headers = {}
        for h in headers:
            lower_h = h.lower().strip()
            if not lower_h: continue
            if lower_h in norm_headers:
                logger.warning(f"⚠️  Duplicate case-insensitive header found: '{h}' conflicts with '{norm_headers[lower_h]}'. Data may be overwritten.")
            norm_headers[lower_h] = h

        # 1. Distributed Sampling for Inference
        sample_rows = []
        try:
            limit = args.sample_size
            
            if limit <= 0:
                logger.info("Performing FULL SCAN for type inference...")
                for row in reader:
                    sample_rows.append(row)
            else:
                f.seek(0, 2)
                total_bytes = f.tell()
                f.seek(0)
                next(csv.reader(f)) # Skip header reset
                f.seek(0)
                next(f) 
                
                # Strengthened Distributed Sampling
                NUM_CHUNKS = 20
                logger.info(f"Performing DISTRIBUTED SAMPLING ({limit} rows) - {NUM_CHUNKS} Chunks @ 5% intervals...")
                
                chunk_size = max(1, int(limit / NUM_CHUNKS))
                
                for i in range(NUM_CHUNKS):
                    # Calculate target offset
                    pct = i / NUM_CHUNKS 
                    
                    if i == 0:
                         pass
                    elif total_bytes > 50000: # Increased threshold to avoid seeking on tiny files
                         target_pos = int(total_bytes * pct)
                         f.seek(target_pos)
                         f.readline() 
                         # logger.info(f"  Chunk {i}: Seek to {target_pos}/{total_bytes} ({int(pct*100)}%)")
                    else:
                         # For small files (<50KB), just read linearly?
                         # If we don't seek, we just continue from previous position.
                         # This effectively means "Linear Read" for small files.
                         pass
                         
                    chunk_reader = csv.reader(f)
                    
                    rows_read_in_chunk = 0
                    try:
                        while rows_read_in_chunk < chunk_size:
                            row = next(chunk_reader)
                            sample_rows.append(row)
                            rows_read_in_chunk += 1
                            if len(sample_rows) >= limit:
                                break
                    except StopIteration:
                         # End of file reached during this chunk
                         pass
                    
                    if len(sample_rows) >= limit:
                        break

                # Fill remaining if any (e.g. if EOF was hit early in chunks)
                # If we are strictly sampling, maybe we don't need to fill exactly to limit 
                # if the file is smaller than limit.

        except Exception as e:
            logger.warning(f"Sampling error: {e}. Falling back to whatever we have.")
        
        # 2. Ensure Schema
        mgr = SchemaManager(client)
        if not mgr.ensure_object(obj_name):
            logger.error(f"Skipping {obj_name} due to object creation failure.")
            return

        typed_map = mgr.ensure_fields(obj_name, headers, sample_rows, concurrency=args.concurrency)
        
    # 3. Import Data with Concurrency
        f.seek(0)
        reader = csv.reader(f)
        next(reader) # Skip header
        
        batch = []
        
        # Calculate Safe Batch Size (MySQL Limit: 65,535 placeholders)
        # Formula: num_rows * num_columns <= 50000 (with safety buffer for system columns)
        num_columns = len(headers)
        if num_columns > 0:
            safe_limit = int(50000 / num_columns) # 50k buffer for safety
            batch_size = min(args.batch_size, safe_limit)
        else:
            batch_size = args.batch_size
            
        if batch_size < args.batch_size:
            logger.info(f"📉 Reduced batch size for {obj_name} from {args.batch_size} to {batch_size} due to column count ({num_columns})")
        else:
            logger.info(f"Using batch size: {batch_size} (Columns: {num_columns})")
        
        import concurrent.futures
        with concurrent.futures.ThreadPoolExecutor(max_workers=args.concurrency) as executor:
            futures = []
            for row in reader:
                if shutdown_signal and shutdown_signal(): break
                
                record = {}
                for i, val in enumerate(row):
                    if i >= len(headers): continue
                    col = headers[i].lower().strip()
                    if not col: continue
                    
                    # Filter empty
                    if not val: continue
                    
                    # Normalize values
                    if col in typed_map:
                        if typed_map[col] == 'DateTime':
                            val = normalize_date(val)
                        elif typed_map[col] == 'Number':
                            val = normalize_number(val)
                        elif typed_map[col] == 'Boolean':
                            val = normalize_boolean(val)
                            
                    record[col] = val
                        
                # Auto-name if missing (only if 'name' was not already in the headers)
                if 'name' not in record and 'name' not in norm_headers:
                    name_candidates = ['id', 'title', 'subject', 'email']
                    for cand in name_candidates:
                        if record.get(cand):
                            record['name'] = str(record[cand])[:255]  # Truncate for safety
                            break
                    if 'name' not in record:
                        record['name'] = f"Record {len(batch) + 1}"
                
                # Skip records with only metadata (id, name) and no actual data
                meaningful_keys = set(record.keys()) - {'id', 'name'}
                if not meaningful_keys:
                    continue
                
                batch.append(record)
                
                if len(batch) >= batch_size:
                    # Submit batch
                    current_batch = list(batch) # copy
                    futures.append(executor.submit(_send_batch, client, obj_name, current_batch))
                    batch = []
                    
            if batch:
                futures.append(executor.submit(_send_batch, client, obj_name, batch))
            
            # Wait for all batches
            concurrent.futures.wait(futures)

def _send_batch(client, obj_name, records):
    payload = {
        "records": records,
        "batch_size": len(records)
    }
    data, err = client.request('POST', f'/api/data/{obj_name}/bulk', payload)
    
    if err:
        logger.error(f"Batch failed: {err}")
        stats.update(total=len(records), error=len(records))
    else:
        # Check inner data
        if "data" in data: data = data["data"]
        s = data.get('success_count', 0)
        f = data.get('failed_count', 0)
        stats.update(total=len(records), success=s, error=f)
        
        if f > 0:
             logger.warning(f"Batch errors: {data.get('errors')[:5]}")
