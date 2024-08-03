from pathlib import Path
from loguru import logger
import sys
import subprocess
from typing import Tuple
import time
import pandas as pd
import argparse
from pprint import pformat

from gremlin_python.process.graph_traversal import GraphTraversal
from gremlin_python.process.graph_traversal import __ as T__
from gremlin_python.process.traversal import P, T, Column
from gremlin_python.driver.protocol import GremlinServerError

from evaluation import Evaluation
import dataset


# CONFIG
# TODO: command line options / yaml config file
ROOT = Path(__file__).parents[1].absolute() # /graph

GREMLIN_URL = 'ws://127.0.0.1:8182/gremlin'


def oracle_manipulation_traverse(g: GraphTraversal):
    return (
        g.V().hasLabel('assetFlow')
        .where(
            T__
            .out('dataflow:read').dedup()
            .emit().repeat(T__.in_('dataflow:dependency')).dedup()
            .has("sourceType", "Storage").in_('dataflow:write').dedup()
            # .emit().repeat(
            #     T__.or_(
            #         T__.out('dcfg_call').in_('dcfg_ret'),
            #         T__.in_('jump')
            #     )
            # ).dedup()
            .emit().repeat(T__.in_('jump'))
            .in_('dataflow:control').dedup()
            .emit().repeat(T__.in_('dataflow:dependency'))
            .has("sourceType", "Caller").limit(1).count().is_(P.gt(0)),
	    ).elementMap()
    )


def oracle_manipulation_price_change_times_traverse(g: GraphTraversal):
    return (
        g.V().hasLabel('assetFlow')
        .local(
            T__
            .out('dataflow:read').dedup()
            .emit().repeat(T__.in_('dataflow:dependency')).dedup()
            .has("sourceType", "Storage").as_('price')
            .local(
                T__
                .in_('dataflow:write').dedup()
                .emit().repeat(T__.in_('jump'))
                .in_('dataflow:control').dedup()
                .emit().repeat(T__.in_('dataflow:dependency'))
                .has("sourceType", "Caller")
                .where(T__.limit(1).count().is_(P.gt(0)))
            )
            .select('price')
        ).dedup()
        .group().by('sourceId').by(
            T__
            .local(
                T__
                .repeat(T__.out('dataflow:transition')).emit()
                .in_('dataflow:write').count()
            ).max_()
        ).select(Column.values).unfold().max_()
    )


def oracle_manipulation_price_traverse(g: GraphTraversal):
    return (
        g.V().hasLabel('assetFlow')
        .local(
            T__
            .out('dataflow:read').dedup()
            .emit().repeat(T__.in_('dataflow:dependency')).dedup()
            .has("sourceType", "Storage").as_('price')
            .local(
                T__
                .in_('dataflow:write').dedup()
                .emit().repeat(T__.in_('jump'))
                .in_('dataflow:control').dedup()
                .emit().repeat(T__.in_('dataflow:dependency'))
                .has("sourceType", "Caller")
                .where(T__.limit(1).count().is_(P.gt(0)))
            )
            .select('price')
        ).values('sourceId').dedup()
    )


def source_change_times_traverse(g: GraphTraversal, source_id):
    return (
        g.V().hasLabel('dataSource').has('sourceId', source_id)
        .local(
            T__
            .emit().repeat(T__.out('dataflow:transition'))
            .in_('dataflow:write').count()
        ).max_()
    )


def oracle_manipulation_state_change_times_traverse(g: GraphTraversal):
    return (
        g.V().hasLabel('dataSource').has('sourceType', 'Storage')
            .group().by('sourceId').by(
                T__
                .local(
                    T__
                    .repeat(T__.out('dataflow:transition')).emit()
                    .in_('dataflow:write').count()
                ).max_()
            ).select(Column.values).unfold().max_()
    )


def oracle_manipulation_has_swap_traverse(g: GraphTraversal):
    return (
        g
        .V().hasLabel('assetFlow').as_('token1').has('to', P.neq('0x0000000000000000000000000000000000000000')).has('from', P.neq('0x0000000000000000000000000000000000000000'))
        .V().hasLabel('assetFlow').as_('token2').has('to', P.neq('0x0000000000000000000000000000000000000000')).has('from', P.neq('0x0000000000000000000000000000000000000000'))
            # .where('token1', P.neq('token2'))
            .where('token1', P.neq('token2')).by('asset')
            .select('token1')
            .has('from', T__.select('token2').values('to'))
            .has('to', T__.select('token2').values('from'))
            .limit(1).count()
    )


def oracle_manipulation_has_borrow_traverse(g: GraphTraversal):
    return (
        g
        .V().hasLabel('assetFlow').as_('collateral').has('to', P.neq('0x0000000000000000000000000000000000000000')).has('from', P.neq('0x0000000000000000000000000000000000000000'))
        .V().hasLabel('assetFlow').as_('token1').has('to', P.neq('0x0000000000000000000000000000000000000000'))
        .V().hasLabel('assetFlow').as_('token2').has('to', P.neq('0x0000000000000000000000000000000000000000'))
            .where('collateral', P.neq('token1')).by('asset')
            .where('collateral', P.neq('token2')).by('asset')
            .where('token1', P.neq('token2')).by('asset')
            .select('collateral')
                .has('to', T__.select('token1').values('from'))
                .has('to', T__.select('token2').values('from'))
            .limit(1).count()
            # .select('has_swap')
    )



class OracleManipulationEvaluation(Evaluation):
    # def detect_attack(self, attack_type, logfile=None) -> Tuple[float, bool]:
    #     time_cost, res = super().traverse_go(attack_type, logfile)
    #     return time_cost, b'label:assetFlow' in res.stderr
    def detect_attack(self, curr_epg_runner, attack_type, logfile=None) -> Tuple[float, bool]:
        if attack_type != 'oracle':
            return 'Not supported attack', False
        g = curr_epg_runner.g.with_('evaluationTimeout', 600000)
        tic = time.perf_counter()
        detected = False
        try:
            res = oracle_manipulation_traverse(g).limit(1).count().toList()[0]
            price_change_times = oracle_manipulation_price_change_times_traverse(g).toList()
            has_swap = oracle_manipulation_has_swap_traverse(g).toList()[0]
            has_borrow = oracle_manipulation_has_borrow_traverse(g).toList()[0]
            if res > 0 and has_swap and has_borrow:
                detected = True
            if price_change_times and price_change_times[0] >= 10:
                detected = True

            # 0. price slots
            # cnt = oracle_manipulation_traverse(g).count().toList()[0]
            # detected = cnt > 0
            

            # 1. price change times
            # price_source_ids = oracle_manipulation_price_traverse(g).toList()
            # if price_source_ids:
            #     price_change_times = [
            #         source_change_times_traverse(g, source_id).toList()[0]
            #         for source_id in price_source_ids
            #     ]
            #     max_price_change_times = max(price_change_times)
            #     if max_price_change_times >= 10:
            #         detected = True
            #     if logfile is not None:
            #         with open(logfile, 'w') as f:
            #             for source_id, change_times in zip(price_source_ids, price_change_times):
            #                 print(f'{source_id},{change_times}', file=f)
            
            # 2. price slot
            # price_slots = oracle_manipulation_price_traverse(g).toList()
            # if price_slots:
            #     detected = True
            #     if logfile is not None:
            #         with open(logfile, 'w') as f:
            #             for slot in price_slots:
            #                 print(f'{slot}', file=f)
            
            time_cost = time.perf_counter() - tic

            # logger.info('write times: %s, %s, %s, %s' % (res, price_change_times, state_write_times, [has_swap, has_borrow]))
        except GremlinServerError as e:
            return 'Timeout', False

        return time_cost, detected


if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('-d', '--dataset',
                        choices=['random', 'attack', 'file'],
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

    attack_type = 'oracle'

    ATTACK_TX_LIST = dataset.ORACLE_TX_LIST
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


    ins = OracleManipulationEvaluation(
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
            res = ins.run_parallel(attack_type, use_tqdm=args.use_tqdm, logfile_dir=args.logfile_dir, cache_dir=args.cache_dir, build_timeout=300)
        else:
            res = ins.run(attack_type, use_tqdm=args.use_tqdm, logfile_dir=args.logfile_dir, cache_dir=args.cache_dir, build_timeout=300)
        res['exp_id'] = [i] * len(res)
        all_res = pd.concat([all_res, res], ignore_index=True)
    logger.info('save to %s' % args.output)
    all_res.to_csv(args.output, index=False)