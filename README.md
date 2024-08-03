# Clue

## Prerequisites

- Go 1.19
- Python
    ```bash
    pip install loguru pandas gremlinpython tqdm
    ```
- Docker

## Build

```bash
cd graph
make
```

## Analyis using trace file

### Set up tinkerpop server

```bash
cd graph
docker compose up -d # This will start the tinkerpop server in the backgorund
```

To stop the server:
```bash
cd graph
docker compose down
```

### Unzip sample dataset

```bash
cd graph/trace
unzip txs.zip
```

### Run reentrancy analysis

#### Attack dataset (sampled)

```bash
cd graph
python3 eval/eval_reentrancy.py -d attack -o output/attack_reentrancy.csv # add -p for parallel evaluation
```

It should correctly identify 1 reentrancy vulnerability in the attack dataset.

#### Control dataset (sampled)

```bash
cd graph
python3 eval/eval_reentrancy.py -d random -o output/control_reentrancy.csv # add -p for parallel evaluation
```

It should not identify any reentrancy vulnerability in the control dataset.

### Run read-only reentrancy analysis

#### Attack dataset (sampled)

```bash
cd graph
python3 eval/eval_reentrancy_read.py -d attack -o output/attack_reentrancy_read.csv # add -p for parallel evaluation
```

It should correctly identify 1 read-only reentrancy vulnerability in the attack dataset.

#### Control dataset (sampled)

```bash
cd graph
python3 eval/eval_reentrancy_read.py -d random -o output/control_reentrancy_read.csv # add -p for parallel evaluation
```

It should not identify any read-only reentrancy vulnerability in the control dataset.

### Run price manipulation analysis

#### Attack dataset (sampled)

```bash
cd graph
python3 eval/eval_oracle.py -d attack -o output/attack_price_manipulation.csv # add -p for parallel evaluation
```

It should correctly identify 1 price manipulation vulnerability in the attack dataset.

#### Control dataset (sampled)

```bash
cd graph
python3 eval/eval_oracle.py -d random -o output/control_price_manipulation.csv # add -p for parallel evaluation
```

It should not identify any price manipulation vulnerability in the control dataset.

### Run analysis on custom dataset

To run analysis on a custom dataset, fill in transaction hashes in `custom.txt` (one per line).

Before running the analysis, make sure to generate the trace file for each transaction (See next section).

```bash
cd graph
python3 eval/eval_<analysis name>.py -d file -f custom.txt -o output/custom.csv # add -p for parallel evaluation; add -a to mark the dataset as attack
```

## Generate trace file from archive node (optional)

To generate trace file for a transaction, you additionally need to have an Ethereum archive node running (with `debug` API enabled).

Do not use Erigon archive node as it does not support standard tracing format.
In evaluation we used `geth` and `reth` archive node.

```bash
cd graph
./build/bin/epg trace --tx <tx hash> --eth-archive-remote <eth archive node url>
```

## Full dataset

Full dataset is available at `graph/full-dataset` in the form of transaction hashes.

## Citation

If you use Clue in your research, please cite our paper:

```
Kaihua Qin, Zhe Ye, Zhun Wang, Weilin Li, Liyi Zhou, Chao Zhang, Dawn Song, and Arthur Gervais. 2025.
Enhancing Smart Contract Security Analysis with Execution Property Graphs. Proc. ACM Softw. Eng. 2, ISSTA,
Article ISSTA049 (July 2025), 22 pages. https://doi.org/10.1145/3728924
```
