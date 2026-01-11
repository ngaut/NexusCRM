import re
import logging
import threading
import time
from dataclasses import dataclass, field

# Logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s | %(levelname)-5s | %(message)s',
    datefmt='%H:%M:%S'
)
logger = logging.getLogger('migration_tool')

@dataclass
class ImportStatistics:
    start_time: float = field(default_factory=time.time)
    total_rows: int = 0
    success_count: int = 0
    error_count: int = 0
    objects_processed: int = 0
    _lock: threading.Lock = field(default_factory=threading.Lock)
    
    def update(self, total=0, success=0, error=0):
        with self._lock:
            self.total_rows += total
            self.success_count += success
            self.error_count += error
    
    @property
    def elapsed_seconds(self) -> float:
        return time.time() - self.start_time
    
    def print_summary(self):
        print("\n" + "=" * 60)
        print("📊 IMPORT SUMMARY")
        print("=" * 60)
        print(f"  Duration:        {self.elapsed_seconds:.1f}s")
        print(f"  Objects:         {self.objects_processed}")
        print(f"  Total Rows:      {self.total_rows:,}")
        print(f"  Success:         {self.success_count:,}")
        print(f"  Errors:          {self.error_count:,}")
        print("=" * 60)

stats = ImportStatistics()

# Normalizers
def normalize_date(val):
    val = val.strip()
    # 1. YYYY.MM.DD or YYYY/MM/DD or YYYY-MM-DD
    # Allows Dot, Slash, Dash
    m = re.match(r'^(\d{4})[\./-](\d{2})[\./-](\d{2})', val)
    if m:
        return f"{m.group(1)}-{m.group(2)}-{m.group(3)}"
        
    # 2. MM/DD/YYYY or DD/MM/YYYY (Slash only for this ambiguous case usually)
    m = re.match(r'^(\d{1,2})/(\d{1,2})/(\d{4})', val)
    if m:
        p1, p2, p3 = int(m.group(1)), int(m.group(2)), m.group(3)
        # Heuristic: if p1 > 12, it MUST be DD/MM
        if p1 > 12:
             return f"{p3}-{p2:02d}-{p1:02d}"
        # simplistic US assumption: MM/DD
        return f"{p3}-{p1:02d}-{p2:02d}"

    return val

def normalize_number(val):
    # Strip currency and commas
    # Keep digits, dot, minus
    if not val: return val
    # Remove common non-numeric chars but be careful not to strip minus sign, dot, or e/E
    clean = val.replace(',', '').replace('$', '').replace('£', '').replace('€', '').strip()
    return clean

def normalize_boolean(val):
    if not val: return None
    v = val.lower().strip()
    if v in ('true', '1', 'yes', 'y', 't'):
        return True
    if v in ('false', '0', 'no', 'n', 'f'):
        return False
    return val # Fallback
