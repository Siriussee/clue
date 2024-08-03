from pathlib import Path

def load_tx_file(tx_file_path):
    tx_list = []
    with open(str(tx_file_path), 'r') as f:
        for line in f:
            line = line.strip()
            if line:
                tx_list.append(line)
    return sorted(tx_list)


ROOT = Path(__file__).parents[1].absolute() # /graph

REENTRANCY_TX_FILE = ROOT / 'trace/reentrancy.txt'
REENTRANCY_READ_TX_FILE = ROOT / 'trace/reentrancy_read.txt'
ORACLE_TX_FILE = ROOT / 'trace/oracle.txt'
DATASET_RANDOM_TX_FILE = ROOT / 'trace/random.txt'

REENTRANCY_TX_LIST = load_tx_file(REENTRANCY_TX_FILE)
ORACLE_TX_LIST = load_tx_file(ORACLE_TX_FILE)
REENTRANCY_READ_TX_LIST = load_tx_file(REENTRANCY_READ_TX_FILE)
DATASET_RANDOM_TX_LIST = load_tx_file(DATASET_RANDOM_TX_FILE)
