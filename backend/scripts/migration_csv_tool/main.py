import argparse
import signal
import sys
from .client import NexusClient
from .importer import process_file
from .utils import stats, logger

shutdown_requested = False

def signal_handler(signum, frame):
    global shutdown_requested
    logger.warning("⚠️  Shutdown requested...")
    shutdown_requested = True

signal.signal(signal.SIGINT, signal_handler)
signal.signal(signal.SIGTERM, signal_handler)

def check_shutdown():
    return shutdown_requested

def main():
    parser = argparse.ArgumentParser(description="Generic CSV Importer")
    parser.add_argument("--file", required=True, help="CSV file path")
    parser.add_argument("--obj", required=True, help="Target Object Name")
    parser.add_argument("--url", required=True, help="API URL")
    parser.add_argument("--token", required=True, help="Auth Token")
    parser.add_argument("--keep-id", action="store_true", help="Keep 'id' column as is")
    parser.add_argument("--sample-size", type=int, default=5000, help="Number of rows to sample. 0 = Full Scan.")
    parser.add_argument("--concurrency", type=int, default=20, help="Max concurrent requests")
    parser.add_argument("--batch-size", type=int, default=1000, help="Rows per batch")
    args = parser.parse_args()
    
    client = NexusClient(args.url, args.token)
    
    print(f"\n>>> Importing {args.obj} from {args.file}...")
    process_file(args.file, args.obj, client, args, shutdown_signal=check_shutdown)
    stats.print_summary()

if __name__ == "__main__":
    main()
