import http.client
import json
import threading
import time
from urllib.parse import urlparse

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
        return self._local.conn

    def request(self, method, path, body=None):
        retries = 3
        backoff = 1.0
        for attempt in range(retries + 1):
            try:
                conn = self._get_connection()
                headers = {
                    'Content-Type': 'application/json',
                    'Authorization': f'Bearer {self.token}'
                }
                json_body = json.dumps(body, cls=NexusJSONEncoder) if body is not None else None
                conn.request(method, path, body=json_body, headers=headers)
                resp = conn.getresponse()
                resp_body = resp.read().decode('utf-8')
                
                if resp.status >= 500 or resp.status == 429:
                    if attempt < retries:
                        time.sleep(backoff)
                        backoff *= 2
                        if hasattr(self._local, 'conn'): del self._local.conn
                        continue
                
                if resp.status >= 400:
                    # Return error for caller to handle
                    return None, f"API Error {resp.status}: {resp_body}"

                try:
                    return json.loads(resp_body) if resp_body else {}, None
                except json.JSONDecodeError:
                    return resp_body, None

            except Exception as e:
                if attempt < retries:
                    time.sleep(backoff)
                    backoff *= 2
                    if hasattr(self._local, 'conn'): del self._local.conn
                    continue
                return None, str(e)
        return None, "Max retries exceeded"

import decimal
import datetime

class NexusJSONEncoder(json.JSONEncoder):
    def default(self, obj):
        if isinstance(obj, (datetime.date, datetime.datetime)):
            return obj.isoformat()
        if isinstance(obj, decimal.Decimal):
            return float(obj)
        return super().default(obj)
