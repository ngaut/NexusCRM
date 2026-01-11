import re
import concurrent.futures
from typing import Dict, List, Optional, Tuple
from .utils import logger

class SchemaManager:
    def __init__(self, client):
        self.client = client

    def infer_type(self, sample_values: List[str], field_name: str) -> Tuple[str, Optional[Dict]]:
        """
        Infer type strictly from data using threshold-based scoring.
        Allows up to 5% dirty data.
        Returns: (TypeName, ExtraParams)
        """
        clean_values = [v for v in sample_values if v and v.strip()]
        total_count = len(clean_values)
        
        if total_count == 0:
            return 'LongTextArea', None
            
        THRESHOLD = 0.99

        # 0. Heuristics based on field name (Safety First)
        lower_name = field_name.lower()
        if 'phone' in lower_name or 'mobile' in lower_name or 'fax' in lower_name:
             return 'LongTextArea', None # Store phones as text
        if '_id' in lower_name and not lower_name.endswith('_id'): 
             # e.g. "external_id_c" -> Text, but "owner_id" might be Lookup (std) or custom lookup
             pass # Continue to data check. If it's numeric, it will be Number. If text, LongTextArea.

        # Helper to calculate match rate
        def get_match_rate(validator_func):
            matches = 0
            for v in clean_values:
                if validator_func(v):
                    matches += 1
            return matches / total_count

        # 1. Boolean
        # Check against comprehensive boolean set
        bool_terms = {'true', 'false', '1', '0', 'yes', 'no', 'y', 'n', 't', 'f'}
        def is_bool(v):
            return v.lower().strip() in bool_terms
            
        if get_match_rate(is_bool) >= THRESHOLD:
            # Special check to avoid confusion with Numbers if data is ONLY 1s and 0s
            # If it's 1/0, prioritizing Boolean is usually fine for CRM.
            return 'Boolean', None

        # 2. DateTime
        # Regexes for common date start patterns (ISO, US/EU, Dot, YYYY/MM/DD)
        date_patterns = [
            re.compile(r'^\d{4}-\d{2}-\d{2}'),      # ISO YYYY-MM-DD
            re.compile(r'^\d{4}/\d{2}/\d{2}'),      # YYYY/MM/DD
            re.compile(r'^\d{4}\.\d{2}\.\d{2}'),    # Dot YYYY.MM.DD
            re.compile(r'^\d{1,2}/\d{1,2}/\d{4}'),  # US/EU MM/DD/YYYY
        ]
        def is_date(v):
            for pat in date_patterns:
                if pat.match(v):
                    return True
            return False
            
        if get_match_rate(is_date) >= THRESHOLD:
            return 'DateTime', None

        # 3. Number
        num_pattern = re.compile(r'^-?\d*\.?\d+([eE][-+]?\d+)?$')
        def is_number(v):
            # Clean common formatting chars
            v_clean = v.replace(',', '').replace('$', '').replace('£', '').replace('€', '').strip()
            # Reject massive ID-like numbers > 15 digits (strip e/E/+/- for digit count)
            clean_digits = re.sub(r'[.\-eE+]', '', v_clean)
            if len(clean_digits) > 15: # ID-like
                return False
            return bool(num_pattern.match(v_clean))

        if get_match_rate(is_number) >= THRESHOLD:
            # Heuristic: if name ends in _id, contains _id_, or ends with _id_c, it's likely NOT a math Number (it's a FK key or external ID)
            lower_name = field_name.lower()
            if not lower_name.endswith('_id') and '_id_' not in lower_name and not lower_name.endswith('_id_c'):
                return 'Number', None

        # 4. Email
        email_pattern = re.compile(r'^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$')
        if get_match_rate(lambda v: bool(email_pattern.match(v))) >= THRESHOLD:
            return 'Email', None

        # 5. URL
        url_pattern = re.compile(r'^https?://')
        if get_match_rate(lambda v: bool(url_pattern.match(v))) >= THRESHOLD:
            return 'Url', None

        # 6. Lookup Heuristic
        if field_name.lower().endswith('_id'):
            # Check length constraint for Lookup (VARCHAR 36 usually)
            # If any value is massive, it's not a standard UUID/Key lookup we can handle with simple FKs.
            # (Unless we change Lookup to VARCHAR 255, but usually it's strict).
            max_len = 0
            for v in clean_values:
                if len(v) > max_len:
                    max_len = len(v)
            
            if max_len <= 36:
                target = field_name[:-3]
                return 'Lookup', {"logicalType": "Lookup", "referenceTo": [target]}
            else:
                # If IDs are huge, treat as Text
                return 'LongTextArea', None

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

            # Truncate to 63 chars (MySQL/TiDB limit)
            if len(col_name) > 63:
                original_col = col_name
                col_name = col_name[:63]
                # Ensure no trailing underscore if strictly truncating? 
                # Actually simpler just to cut.
                print(f"⚠️ Truncated field '{original_col}' -> '{col_name}'")

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
