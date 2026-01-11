import re
import concurrent.futures
from typing import Dict, List, Optional, Tuple
from .utils import logger

class SchemaManager:
    def __init__(self, client):
        self.client = client

    def infer_type(self, sample_values: List[str], field_name: str) -> Tuple[str, Optional[Dict]]:
        """
        Infer type strictly from data.
        Returns: (TypeName, ExtraParams)
        """
        clean_values = [v for v in sample_values if v and v.strip()]
        
        if not clean_values:
            return 'LongTextArea', None

        # 0. Heuristics based on field name (Safety First)
        lower_name = field_name.lower()
        if 'phone' in lower_name or 'mobile' in lower_name or 'fax' in lower_name:
             return 'LongTextArea', None # Store phones as text
        if '_id' in lower_name and not lower_name.endswith('_id'): 
             # e.g. "external_id_c" -> Text, but "owner_id" might be Lookup (std) or "custom_lookup_id_c"??
             # Actually, simpler: if it looks like an ID but isn't a lookup ref, keep it text.
             # "zoom_info_company_id_c" contained "_id" but validation failed as Number.
             return 'LongTextArea', None

        # 1. Boolean
        # Strict logic: values must belong to a consistent boolean set
        unique_lower = {v.lower().strip() for v in clean_values}
        bool_sets = [
            {'true', 'false'},
            {'1', '0'},
            {'yes', 'no'},
            {'y', 'n'},
            {'t', 'f'}
        ]
        is_bool = False
        for s in bool_sets:
            if unique_lower.issubset(s):
                is_bool = True
                break
        if is_bool:
            return 'Boolean', None

        # 2. DateTime
        is_date = True
        # Regexes for common date start patterns (ISO, US/EU, Dot, YYYY/MM/DD)
        date_patterns = [
            re.compile(r'^\d{4}-\d{2}-\d{2}'),      # ISO YYYY-MM-DD
            re.compile(r'^\d{4}/\d{2}/\d{2}'),      # YYYY/MM/DD
            re.compile(r'^\d{4}\.\d{2}\.\d{2}'),    # Dot YYYY.MM.DD
            re.compile(r'^\d{1,2}/\d{1,2}/\d{4}'),  # US/EU MM/DD/YYYY
        ]
        
        for v in clean_values:
            matched_any = False
            for pat in date_patterns:
                if pat.match(v):
                    matched_any = True
                    break
            if not matched_any:
                is_date = False
                break
        if is_date:
            return 'DateTime', None

        # 3. Number
        # Allow integers, floats, negatives, and currency/commas ($1,000.50)
        is_number = True
        num_pattern = re.compile(r'^-?\d*\.?\d+([eE][-+]?\d+)?$')
        for v in clean_values:
            # Clean common formatting chars
            v_clean = v.replace(',', '').replace('$', '').replace('£', '').replace('€', '').strip()
            
            # Reject massive ID-like numbers > 15 digits (strip e/E/+/- for digit count)
            clean_digits = re.sub(r'[.\-eE+]', '', v_clean)
            if len(clean_digits) > 15 or not num_pattern.match(v_clean):
                is_number = False
                break
        if is_number:
            # Heuristic: if name ends in _id, it's likely NOT a math Number (it's a FK key)
            if not field_name.lower().endswith('_id'): 
                return 'Number', None

        # 4. Email
        is_email = True
        # Simple robust email regex
        email_pattern = re.compile(r'^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$')
        for v in clean_values:
            if not email_pattern.match(v):
                is_email = False
                break
        if is_email:
            return 'Email', None

        # 5. URL
        is_url = True
        url_pattern = re.compile(r'^https?://')
        for v in clean_values:
            if not url_pattern.match(v):
                is_url = False
                break
        if is_url:
            return 'Url', None

        # 6. Lookup Heuristic
        if field_name.lower().endswith('_id'):
            target = field_name[:-3]
            # Try to infer target object name from field name
            return 'Lookup', {"logicalType": "Lookup", "referenceTo": [target]}

        return 'LongTextArea', None

    def ensure_object(self, obj_name):
        print(f"Checking object: {obj_name}...")
        _, err = self.client.request('GET', f'/api/metadata/objects/{obj_name}')
        if not err:
            return True

        print(f"Creating object {obj_name}...")
        payload = {
            "api_name": obj_name,
            "label": obj_name.replace('_', ' ').title(),
            "description": f"Imported Generic Object {obj_name}"
        }
        _, err = self.client.request('POST', '/api/metadata/objects', payload)
        if err:
            print(f"Failed to create object {obj_name}: {err}")
            return False
        return True

    def ensure_fields(self, obj_name, headers, sample_rows, concurrency=20) -> Dict[str, str]:
        print(f"Ensuring fields for {obj_name}...")
        
        fields_to_create = []
        field_types = {} # local map: col -> type
        
        # Transpose samples for column-wise analysis
        columns = {}
        for h in headers:
            columns[h] = []
        
        for row in sample_rows:
            for i, h in enumerate(headers):
                if i < len(row):
                    columns[h].append(row[i])
                    
        for header in headers:
            col_name = header.lower().strip()
            # SKIP: empty headers only
            if not col_name:
                continue

            # Infer type
            samples = columns[header]
            type_name, extra = self.infer_type(samples, col_name)
            field_types[col_name] = type_name
            
            fields_to_create.append((col_name, type_name, header, extra))

        # Create in parallel
        with concurrent.futures.ThreadPoolExecutor(max_workers=concurrency) as executor:
            futures = []
            for api, type_, label, extra in fields_to_create:
                futures.append(executor.submit(self._create_field_safe, obj_name, api, type_, label, extra))
            concurrent.futures.wait(futures)
            
        return field_types

    def _create_field_safe(self, obj_name, api_name, type_name, label, extra=None):
        # wrapper to handle lookup fallback
        err = self._create_field(obj_name, api_name, type_name, label, extra)
        if err:
            if extra and "Lookup" in str(extra):
                 # Fallback to Text if Lookup creation failed (likely target table missing)
                print(f"⚠️ Lookup creation failed for {api_name} ({err}). Falling back to LongTextArea.")
                self._create_field(obj_name, api_name, "LongTextArea", label)
            else:
                print(f"❌ Field creation failed for {api_name}: {err}")

    def _create_field(self, obj_name, api_name, type_name, label, extra_params=None):
        payload = {
            "api_name": api_name,
            "label": label,
            "type": type_name
        }
        if extra_params:
            payload.update(extra_params)
            # Remove physical type if logical is present? No, API expects 'type' (physical) AND 'logicalType' (optional).
            if extra_params.get("logicalType") == "Lookup":
                payload["type"] = "VARCHAR(255)" # Enforce physical type for Lookup
        
        _, err = self.client.request('POST', f'/api/metadata/objects/{obj_name}/fields', payload)
        return err
