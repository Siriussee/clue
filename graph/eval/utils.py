import requests
from loguru import logger
from pathlib import Path
import shelve
import time

SCRIPT_DIR = Path(__file__).parent.absolute()
CACHE_DIR = SCRIPT_DIR / '__mycache__'

CACHE_TRACE_FILE = str(CACHE_DIR / 'trace.shelve')
def get_trace_api(tx_hash, cache_file=CACHE_TRACE_FILE):
    API_URL = 'https://tx.eth.samczsun.com/api/v1/trace/ethereum/{tx_hash}'
    
    with shelve.open(cache_file) as trace_db:
        if tx_hash in trace_db:
            return trace_db[tx_hash]
        target_url = API_URL.format(tx_hash=tx_hash)
        r = requests.get(target_url)
        if not r.ok:
            logger.warning('sleep 5 sec then retry one time')
            time.sleep(5)
            r = requests.get(target_url)
            if not r.ok:
                logger.error('failed to get trace info of <%s>\nurl: %s' % (tx_hash, target_url))
                return ''
        trace = r.text
        trace_db[tx_hash] = trace
        return  trace