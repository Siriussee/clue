from pathlib import Path
from loguru import logger
import sys
import subprocess
from typing import Tuple
from itertools import chain
import time
import argparse
from pprint import pformat
import pandas as pd

from gremlin_python.process.graph_traversal import GraphTraversal
from gremlin_python.process.graph_traversal import __ as T__
from gremlin_python.process.traversal import P, T
from gremlin_python.driver.protocol import GremlinServerError

from evaluation import Evaluation
from epg_graph import DcfgId
import dataset

# CONFIG
# TODO: command line options / yaml config file
ROOT = Path(__file__).parents[1].absolute() # /graph

GREMLIN_URL = 'ws://127.0.0.1:8182/gremlin'

def reentrancy_control_dependency_traverse(g: GraphTraversal, n_pairs_limit=500):
    return (
        g.V().hasLabel('contractCall').as_('attacker')
            .repeat(T__.out('call')).emit().as_('victim')
            .inE('call').has('callTrace:type', P.neq('DELEGATECALL'))
            .where('attacker', P.neq('victim')).by('address')
            .select('victim')
            .limit(n_pairs_limit)
            .repeat(T__.out('call')).emit().as_('re_attacker')
            .inE('call').has('callTrace:type', P.neq('DELEGATECALL'))
            .where('attacker', P.eq('re_attacker')).by('address')
            .select('re_attacker')
            .repeat(T__.out('call')).emit().as_('re_victim')
            .inE('call').has('callTrace:type', P.neq('DELEGATECALL'))
            .where('victim', P.eq('re_victim')).by('address')
            .select('re_victim')
            .limit(n_pairs_limit)
        # g.V().hasLabel('contractCall').as_('victim')
        #     .repeat(T__.out('call')).emit().as_('attacker')
        #     .inE('call').has('callTrace:type', P.neq('DELEGATECALL'))
        #     .where('victim', P.neq('attacker')).by('address')
        #     .select('attacker')
        #     .repeat(T__.out('call')).emit().as_('re_victim')
        #     .inE('call').has('callTrace:type', P.neq('DELEGATECALL'))
        #     .where('victim', P.eq('re_victim')).by('address')
        #     .select('re_victim')
            .local(
                T__
                .emit().repeat(T__.out('call'))
                .out('transfer').as_('victim_flow')
                .in_('dcfg_to_asset_flow').dedup().as_('victim_flow_dcfg')
                .emit().repeat(T__.in_('jump'))
                .in_('dataflow:control')
                .emit().repeat(T__.in_('dataflow:dependency')).dedup()
                .has('sourceType', 'Storage')
                .repeat(T__.out('dataflow:transition')).emit().as_('state_change')
                .in_('dataflow:write').as_('state_change_dcfg').dedup()
                .where(
                    T__
                    .repeat(T__.in_('jump')).emit()
                    .hasLabel('contractCall')
                    .emit().repeat(T__.in_('call'))
                    .where(P.eq('victim')).count().is_(P.gt(0))
                )
                .select('attacker', 're_attacker', 'victim', 're_victim', 'state_change', 'state_change_dcfg', 'victim_flow', 'victim_flow_dcfg').by(T__.elementMap()).dedup()
                # .select('attacker', 'victim', 're_victim', 'state_change', 'state_change_dcfg', 'victim_flow', 'victim_flow_dcfg').by(T__.elementMap()).dedup()
                # .select('victim_flow_dcfg', 'state_change_dcfg').by('dcfgId').dedup()
            )
    )


def reentrancy_amount_dependency_traverse(g: GraphTraversal, n_pairs_limit=500):
    return (
        g.V().hasLabel('contractCall').as_('attacker')
            .repeat(T__.out('call')).emit().as_('victim')
            .inE('call').has('callTrace:type', P.neq('DELEGATECALL'))
            .where('attacker', P.neq('victim')).by('address')
            .select('victim')
            .limit(n_pairs_limit)
            .repeat(T__.out('call')).emit().as_('re_attacker')
            .inE('call').has('callTrace:type', P.neq('DELEGATECALL'))
            .where('attacker', P.eq('re_attacker')).by('address')
            .select('re_attacker')
            .repeat(T__.out('call')).emit().as_('re_victim')
            .inE('call').has('callTrace:type', P.neq('DELEGATECALL'))
            .where('victim', P.eq('re_victim')).by('address')
            .select('re_victim')
            .limit(n_pairs_limit)
        # g.V().hasLabel('contractCall').as_('victim')
        #     .repeat(T__.out('call')).emit().as_('attacker')
        #     .inE('call').has('callTrace:type', P.neq('DELEGATECALL'))
        #     .where('victim', P.neq('attacker')).by('address')
        #     .select('attacker')
        #     .repeat(T__.out('call')).emit().as_('re_victim')
        #     .inE('call').has('callTrace:type', P.neq('DELEGATECALL'))
        #     .where('victim', P.eq('re_victim')).by('address')
        #     .select('re_victim')
            .local(
                T__
                .emit().repeat(T__.out('call'))
                .out('transfer').as_('victim_flow')
                .out('dataflow:read').dedup()
                .emit().repeat(T__.in_('dataflow:dependency')).dedup()
                .has('sourceType', 'Storage')
                .repeat(T__.out('dataflow:transition')).emit().as_('state_change')
                .in_('dataflow:write').as_('state_change_dcfg').dedup()
                .where(
                    T__
                    .repeat(T__.in_('jump')).emit()
                    .has_label('contractCall')
                    .emit().repeat(T__.in_('call'))
                    .where(P.eq('victim')).count().is_(P.gt(0))
                )
                .select('victim_flow').in_('dcfg_to_asset_flow').as_('victim_flow_dcfg')
                .select('attacker', 're_attacker', 'victim', 're_victim', 'state_change', 'state_change_dcfg', 'victim_flow', 'victim_flow_dcfg').by(T__.element_map()).dedup()
                # .select('attacker', 'victim', 're_victim', 'state_change', 'state_change_dcfg', 'victim_flow', 'victim_flow_dcfg').by(T__.element_map()).dedup()
                # .select('victim_flow_dcfg', 'state_change_dcfg').by('dcfgId').dedup()
            )
    )


def reentrancy_external_control_dependency_traverse(g: GraphTraversal):
    return (
        g.V().hasLabel('contractCall').as_('attacker')
            .repeat(T__.out('call')).emit().as_('victim')
            .inE('call').has('callTrace:type', P.neq('DELEGATECALL'))
            .where('attacker', P.neq('victim')).by('address')
            .select('victim')
            .repeat(T__.out('call')).emit().as_('re_attacker')
            .inE('call').has('callTrace:type', P.neq('DELEGATECALL'))
            .where('attacker', P.eq('re_attacker')).by('address')
            .select('re_attacker')
            .repeat(T__.out('call')).emit().as_('re_victim')
            .inE('call').has('callTrace:type', P.neq('DELEGATECALL'))
            .where('victim', P.eq('re_victim')).by('address')
            .V().hasLabel('contractCall').where(P.eq('attacker')).by('address')
            .emit().repeat(T__.out('call')).dedup()
            .out('transfer').as_('victim_flow').dedup()
            .local(
                T__
                # .emit().repeat(T__.out('call'))
                # .out('transfer').as_('victim_flow')
                .in_('dcfg_to_asset_flow').as_('victim_flow_dcfg')
                .emit().repeat(
                    T__.union(
                        T__.out('dcfg_call').in_('dcfg_ret'),
                        T__.in_('jump')
                    )
                ).dedup()
                .in_('dataflow:control')
                .emit().repeat(T__.in_('dataflow:dependency')).dedup()
                .has('sourceType', 'Storage')
                .repeat(T__.out('dataflow:transition')).emit().as_('state_change')
                .in_('dataflow:write').as_('state_change_dcfg').dedup()
                .where(
                    T__
                    .repeat(T__.in_('jump')).emit()
                    .hasLabel('contractCall')
                    .emit().repeat(T__.in_('call'))
                    .where(P.eq('re_attacker'))
                )
                .select('attacker', 're_attacker', 'victim', 're_victim', 'state_change', 'state_change_dcfg', 'victim_flow', 'victim_flow_dcfg').by(T__.element_map()).dedup()
            )
    )

def reentrancy_external_amount_dependency_traverse(g: GraphTraversal):
    return (
        g.V().hasLabel('contractCall').as_('attacker')
            .repeat(T__.out('call')).emit().as_('victim')
            .inE('call').has('callTrace:type', P.neq('DELEGATECALL'))
            .where('attacker', P.neq('victim')).by('address')
            .select('victim')
            .repeat(T__.out('call')).emit().as_('re_attacker')
            .inE('call').has('callTrace:type', P.neq('DELEGATECALL'))
            .where('attacker', P.eq('re_attacker')).by('address')
            .select('re_attacker')
            .repeat(T__.out('call')).emit().as_('re_victim')
            .inE('call').has('callTrace:type', P.neq('DELEGATECALL'))
            .where('victim', P.eq('re_victim')).by('address')
            .V().hasLabel('contractCall').where(P.eq('attacker')).by('address')
            .emit().repeat(T__.out('call'))
            .out('transfer').as_('victim_flow').dedup()
            .local(
                T__
                .out('dataflow:read').dedup()
                .emit().repeat(T__.in_('dataflow:dependency')).dedup()
                .has('sourceType', 'Storage')
                .repeat(T__.out('dataflow:transition')).emit().as_('state_change')
                .in_('dataflow:write').as_('state_change_dcfg').dedup()
                .where(
                    T__
                    .repeat(T__.in_('jump')).emit()
                    .has_label('contractCall')
                    .emit().repeat(T__.in_('call'))
                    .where(P.eq('re_attacker'))
                )
                .select('victim_flow').in_('dcfg_to_asset_flow').as_('victim_flow_dcfg')
                .select('attacker', 're_attacker', 'victim', 're_victim', 'state_change', 'state_change_dcfg', 'victim_flow', 'victim_flow_dcfg').by(T__.element_map()).dedup()
            )
    )


class ReentrancyEvaluation(Evaluation):
    # def detect_attack(self, attack_type, logfile=None) -> Tuple[float, bool]:
    #     time_cost, res = super().traverse_go(attack_type, logfile)
    #     return time_cost, b'Victime' in res.stderr

    def detect_attack(self, curr_epg_runner, attack_type, logfile=None) -> Tuple[float, bool]:
        if attack_type != 'reentrancy':
            return 'Not supported type', False

        tic = time.perf_counter()
        detected = False
        attack_cands = []
        g = curr_epg_runner.g
        # logger.info(g)
        try:
            for x in chain(reentrancy_amount_dependency_traverse(g.with_('evaluationTimeout', 300000)),
                        reentrancy_control_dependency_traverse(g.with_('evaluationTimeout', 300000))):
                victim_flow_dcfg_id = DcfgId.from_str(x['victim_flow_dcfg']['dcfgId'])
                # victim_flow_dcfg_id = DcfgId.from_str(x['victim_flow_dcfg'])
                state_change_dcfg_id = DcfgId.from_str(x['state_change_dcfg']['dcfgId'])
                # state_change_dcfg_id = DcfgId.from_str(x['state_change_dcfg'])
                if victim_flow_dcfg_id < state_change_dcfg_id:
                    detected = True
                    if logfile is not None:
                        attack_cands.append(x)
                    else:
                        break
            # if not detected or logfile is not None:
            #     # for x in chain(reentrancy_external_control_dependency_traverse(g.with_('evaluationTimeout', 300000)),
            #     #                reentrancy_external_amount_dependency_traverse(g.with_('evaluationTimeout', 300000))):
            #     for x in reentrancy_external_control_dependency_traverse(g.with_('evaluationTimeout', 300000)):
            #         victim_flow_dcfg_id = DcfgId.from_str(x['victim_flow_dcfg']['dcfgId'])
            #         state_change_dcfg_id = DcfgId.from_str(x['state_change_dcfg']['dcfgId'])
            #         if victim_flow_dcfg_id > state_change_dcfg_id:
            #             detected = True
            #             if logfile is not None:
            #                 attack_cands.append(x)
            #             else:
            #                 break
            time_cost = time.perf_counter() - tic
        except GremlinServerError as e:
            return 'Timeout', False

        if detected and logfile is not None:
            with open(logfile, 'w') as f:
                for x in attack_cands:
                    print(pformat(x), file=f)
        return time_cost, detected


def cursor2tx(mongo_cursor):
    for itm in mongo_cursor:
        yield '0x' + itm['txhash'].hex()


if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('-d', '--dataset',
                        choices=['gas', 'random', 'attack', 'random_large', 'file', 'query'],
                        help='dataset to use, dataset `file` requires `-f` option to specify the file')
    parser.add_argument('-f', '--dataset-file',
                        help='dataset file to use, only valid when `-d file` is specified')
    parser.add_argument('-o', '--output',
                        default='/tmp/res_reentrancy.csv',
                        help='output result file')
    parser.add_argument('-l', '--logfile-dir',
                        help='log file directory')
    parser.add_argument('-c', '--cache-dir',
                        help='cache directory')
    parser.add_argument('-n', '--repeat-times',
                        type=int, default=1,
                        help='repeat times')
    parser.add_argument('-q', '--no-tqdm',
                        action='store_true',
                        help='do not use tqdm to show progress bar')
    parser.add_argument('-p', '--parallel',
                        action='store_true',
                        help='use parallel based on the number of gremlin servers')
    parser.add_argument('-a', '--is-attack',
                        action='store_true',
                        help='mark the evaluated dataset as attack dataset')

    args = parser.parse_args()
    args.use_tqdm = not args.no_tqdm

    import resource
    resource.setrlimit(
        resource.RLIMIT_NOFILE,
        (4096, 4096)
    )
    
    attack_type = 'reentrancy'

    ATTACK_TX_LIST = dataset.REENTRANCY_TX_LIST

    if args.dataset == 'gas':
        DATASET_TX_LIST = sorted(set(dataset.DATASET_GAS_TX_LIST) - set(ATTACK_TX_LIST))
    elif args.dataset == 'random':
        DATASET_TX_LIST = sorted(set(dataset.DATASET_RANDOM_TX_LIST) - set(ATTACK_TX_LIST))
    elif args.dataset == 'attack':
        DATASET_TX_LIST = ATTACK_TX_LIST
    elif args.dataset == 'file':
        DATASET_TX_LIST = dataset.load_tx_file(args.dataset_file)

    if args.is_attack:
        ATTACK_TX_LIST = DATASET_TX_LIST

    ins = ReentrancyEvaluation(
        gremlin_url=GREMLIN_URL,
        dataset_tx=DATASET_TX_LIST,
        attack_tx=ATTACK_TX_LIST,
        epg_path=ROOT / 'build/bin/epg',
    )

    logger.remove()
    logger.add(sys.stderr, level='INFO')
    all_res = pd.DataFrame()
    for i in range(args.repeat_times):
        logger.info('exp: %d' % i)
        if args.parallel:
            res = ins.run_parallel(attack_type, use_tqdm=args.use_tqdm, logfile_dir=args.logfile_dir, cache_dir=args.cache_dir)
        else:
            res = ins.run(attack_type, use_tqdm=args.use_tqdm, logfile_dir=args.logfile_dir, cache_dir=args.cache_dir)
        res['exp_id'] = [i] * len(res)
        all_res = pd.concat([all_res, res], ignore_index=True)
    logger.info('save to %s' % args.output)
    all_res.to_csv(args.output, index=False)