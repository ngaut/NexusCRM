import csv
import os
import sys
import pandas as pd
import pyarrow.dataset as ds
from .utils import logger, stats, normalize_date, normalize_number, normalize_boolean
from .schema import SchemaManager

# Increase CSV field size limit to handle large formula fields (default is 128KB, set to 32MB)
csv.field_size_limit(32 * 1024 * 1024)

def process_file(file_path, obj_name, client, args, shutdown_signal=None):
    stats.objects_processed += 1
    
    # Check if Parquet (Directory or File)
    # If directory, assume Parquet dataset. If file ends with .parquet, assume Parquet.
    is_parquet = os.path.isdir(file_path) or file_path.endswith('.parquet')
    
    if is_parquet:
        _process_parquet(file_path, obj_name, client, args, shutdown_signal)
    else:
        _process_csv(file_path, obj_name, client, args, shutdown_signal)

def _process_parquet(file_path, obj_name, client, args, shutdown_signal):
    logger.info(f"📂 Processing Parquet dataset: {file_path}")
    
    try:
        # Use PyArrow Dataset for efficient batching of partitions
        dataset = ds.dataset(file_path, format="parquet")
        
        # 1. Ensure Schema
        mgr = SchemaManager(client)
        if not mgr.ensure_object(obj_name):
            return
            
        mgr.ensure_fields_parquet(obj_name, dataset.schema, concurrency=args.concurrency)
        
        # 2. Iterate Batches
        batch_count = 0
        total_rows = 0
        
        # To avoid OOM, process in batches. 
        # dataset.to_batches() yields RecordBatches
        # args.batch_size default 1000
        
        batches = dataset.to_batches(batch_size=args.batch_size)
        
        import concurrent.futures
        with concurrent.futures.ThreadPoolExecutor(max_workers=args.concurrency) as executor:
            futures = []
            
            for batch in batches:
                if shutdown_signal and shutdown_signal(): break
                
                df = batch.to_pandas()
                # Clean NaNs (JSON standard doesn't support NaN)
                # Must cast to object to allow None in float columns, otherwise None becomes NaN again
                df = df.astype(object).where(pd.notnull(df), None)
                
                # Convert to records: [{'col': val}, ...]
                records = df.to_dict(orient='records')
                
                # Filter/Clean records & Generate Names
                clean_records = []
                for rec in records:
                    # Auto-name if needed
                    # Note: Parquet data might use mixed case keys if specified in schema, 
                    # but Nexus schema normalizes to lowercase.
                    # We should probably normalize keys to lowercase to match SchemaManager creation?
                    # dataset.schema names match dataframe columns.
                    # SchemaManager used `field.name.lower().strip()`.
                    # So we should lowercase keys here too or rely on `to_dict` handling it?
                    # Better to normalize keys to ensure match.
                    
                    norm_rec = {}
                    for k, v in rec.items():
                        norm_rec[k.lower()] = v
                    rec = norm_rec

                    if 'name' not in rec:
                        # heuristic
                        name_val = rec.get('title') or rec.get('subject') or rec.get('email') or rec.get('id')
                        if name_val:
                            rec['name'] = str(name_val)[:255]
                        else:
                            # Use a global counter offset? Or just batch index?
                            rec['name'] = f"Record" # Simple fallback
                    
                    clean_records.append(rec)

                if clean_records:
                    futures.append(executor.submit(_send_batch, client, obj_name, clean_records))
                    total_rows += len(clean_records)
            
            concurrent.futures.wait(futures)
            logger.info(f"✅ Finished Parquet import for {obj_name}. Total rows: {total_rows}")
            
    except Exception as e:
        logger.error(f"❌ Parquet processing error for {obj_name}: {e}")
        # Don't crash entire script, just log error
        import traceback
        traceback.print_exc()

def _process_csv(file_path, obj_name, client, args, shutdown_signal):
    with open(file_path, 'r', encoding='utf-8-sig', errors='replace') as f:
        reader = csv.reader(f)
        try:
            headers = next(reader)
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
                    pct = i / NUM_CHUNKS 
                    if i == 0: pass
                    elif total_bytes > 50000:
                         target_pos = int(total_bytes * pct)
                         f.seek(target_pos)
                         f.readline() 
                    
                    chunk_reader = csv.reader(f)
                    rows_read_in_chunk = 0
                    try:
                        while rows_read_in_chunk < chunk_size:
                            row = next(chunk_reader)
                            sample_rows.append(row)
                            rows_read_in_chunk += 1
                            if len(sample_rows) >= limit: break
                    except StopIteration: pass
                    if len(sample_rows) >= limit: break

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
        
        # Calculate Safe Batch Size
        num_columns = len(headers)
        if num_columns > 0:
            safe_limit = int(50000 / num_columns)
            batch_size = min(args.batch_size, safe_limit)
        else:
            batch_size = args.batch_size
            
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
                        
                if 'name' not in record and 'name' not in norm_headers:
                    name_candidates = ['id', 'title', 'subject', 'email']
                    for cand in name_candidates:
                        if record.get(cand):
                            record['name'] = str(record[cand])[:255]
                            break
                    if 'name' not in record:
                        record['name'] = f"Record {len(batch) + 1}"
                
                meaningful_keys = set(record.keys()) - {'id', 'name'}
                if not meaningful_keys:
                    continue
                
                batch.append(record)
                
                if len(batch) >= batch_size:
                    current_batch = list(batch)
                    futures.append(executor.submit(_send_batch, client, obj_name, current_batch))
                    batch = []
                    
            if batch:
                futures.append(executor.submit(_send_batch, client, obj_name, batch))
            
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
