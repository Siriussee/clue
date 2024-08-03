import subprocess
from loguru import logger
from tqdm import tqdm
from pathlib import Path
from typing import List, Dict, Set, Callable, Iterable, Tuple
import shlex
import time
import pandas as pd
import os
import multiprocessing as mp

from gremlin_python.process.anonymous_traversal import traversal
from gremlin_python.driver.driver_remote_connection import DriverRemoteConnection



ROOT = Path(__file__).parents[1].absolute() # /graph


class EPGRunner:
    '''
    drop graph:
    >>> build/bin/epg drop --remote $gremlin_url

    build graph:
    >>> build/bin/epg build-remote --remote $gremlin_url --tx $tx_hash

    export graph:
    >>> build/bin/epg export --format [json|xml] --remote $gremlin_url --name $filename

    import graph:
    >>> build/bin/epg import --format [json|xml] --remote $gremlin_url --name $filename
    '''
    def __init__(self, gremlin_url: str, epg_path: str=None, exported_graph_dir: str=None):
        self.gremlin_url = gremlin_url

        self.epg_path = str(ROOT / 'build/bin/epg') if epg_path is None \
                        else str(epg_path)

        self.exported_graph_dir = str(ROOT / 'exported-graphs') if exported_graph_dir is None \
                                else exported_graph_dir


        self.g = traversal().withRemote(DriverRemoteConnection(self.gremlin_url,'g'))
        self.build_result = None

        return


    def drop_graph(self):
        logger.debug('drop graph: start')
        self.g.V().drop().iterate()
        assert self.g.V().count().toList()[0] == 0
        logger.debug('drop graph: end')


    def export_graph(self, filepath):
        logger.debug('export to %s/%s' % (self.exported_graph_dir, filepath))
        self.g.io('/exported-graphs/%s' % (filepath,)).write().iterate()


    def import_graph(self, filepath):
        logger.debug('import from %s/%s' % (self.exported_graph_dir, filepath))
        self.g.io('/exported-graphs/%s' % (filepath,)).read().iterate()


    def build_graph(self, tx_hash, use_cache=True, timeout=180, cache_dir=None):
        '''
        >>> build/bin/epg build --remote $gremlin_url --tx $tx_hash

        Returns
        -------
        float: time (seconds) of graph building
        str: error, 'Not Found' | 'Timeout'
        '''
        if cache_dir is None:
            exported_graph_path = f'{tx_hash}.xml'
        else:
            exported_graph_path = f'{cache_dir}/{tx_hash}.xml'
        if use_cache and os.path.isfile(f'{self.exported_graph_dir}/{exported_graph_path}'):
            tic = time.perf_counter()
            self.import_graph(exported_graph_path)
            time_cost = time.perf_counter() - tic
            logger.debug('load graph of tx <%s> from cache, cost time: %.06fs' % (tx_hash, time_cost))
            return time_cost

        cmd = [
                self.epg_path, 'build',
                '--remote', self.gremlin_url,
                '--time',
                '--tx', tx_hash,
            ]
        print(self.epg_path)

        logger.debug('build graph of tx <%s>\n%s' % (tx_hash, shlex.join(cmd)))
        try:
            tic = time.perf_counter()
            self.build_result = subprocess.run(cmd, capture_output=True, timeout=timeout)
            time_cost = time.perf_counter() - tic
        except subprocess.TimeoutExpired:
            logger.error('build graph of tx <%s> timeout <%ds>' % (tx_hash, timeout))
            return 'Timeout'

        if self.build_result.returncode != 0:
            logger.error('build graph of tx <%s> failed, output:\n%s' % (tx_hash, self.build_result.stderr.decode()))
            return 'Failed'

        # !!! save all graphs
        if use_cache:
            logger.debug('save graph to %s/%s' % (self.exported_graph_dir, exported_graph_path))
            self.export_graph(exported_graph_path)

        logger.debug('build group time: %.06fs, return: %d' % (time_cost, self.build_result.returncode))

        return time_cost

class Evaluation:

    def __init__(self, gremlin_url: str=None, dataset_tx: Iterable[str]=None, attack_tx: Iterable[str]=None, epg_path: str=None, exported_graph_dir: str=None):
        self.dataset_tx = [] if dataset_tx is None else sorted(dataset_tx)
        self.attack_tx = set() if attack_tx is None else set(attack_tx)
        
        self.gremlin_urls = [gremlin_url]
        
        self.epg_runners = [
            EPGRunner(gremlin_url, epg_path, exported_graph_dir)
            for gremlin_url in self.gremlin_urls
        ]


    def detect_attack(self, curr_epg_runner, attack_type, logfile=None) -> Tuple[float, bool]:
        raise NotImplemented
    
    @property
    def epg_runner(self):
        return self.epg_runners[0]

    @staticmethod
    def summary(res: pd.DataFrame):
        '''
        show TP, TN, FP, FN
        '''
        # https://stackoverflow.com/questions/45338209/filtering-string-float-integer-values-in-pandas-dataframe-columns
        traverse_time = pd.to_numeric(res['traverse_time'], errors='coerce')
        mask = traverse_time.notnull()
        TP = (mask & res['is_attack'] & res['detect_attack']).sum()
        TN = (mask & ~res['is_attack'] & ~res['detect_attack']).sum()
        FP = (mask & ~res['is_attack'] & res['detect_attack']).sum()
        FN = (mask & res['is_attack'] & ~res['detect_attack']).sum()
        total = mask.sum()
        accuracy = (TN + TP) / total
        precision = TP / (TP + FP)
        recall = TP / (TP + FN)
        F1_score = 2 * precision * recall / (precision + recall)
        logger.info(f'total: {total}, TP: {TP}, TN: {TN}, FP: {FP}, FN: {FN}')
        logger.info(f'Accuracy: {accuracy:.6f}, Precision: {precision:.6f}, Recall: {recall:.6f}, F1 score: {F1_score:.6f}')
        
        logger.info(f'traverse time: {traverse_time.mean():.6f} ± {traverse_time.std():.6f}s, max: {traverse_time.max():.6f}s, min: {traverse_time.min():.6f}s')


    def run_one(self, tx_hash, attack_type, logfile=None, use_cache=True):
        logger.info(f'detect attack<{attack_type}> in tx<{tx_hash}>, output to {logfile}')

        self.epg_runner.drop_graph()
        # time.sleep(0.3)
        build_time = self.epg_runner.build_graph(tx_hash, use_cache=use_cache)
        if isinstance(build_time, str):
            return
        traverse_time, detect_attack = self.epg_runner.traverse_go(attack_type, logfile=logfile)
        logger.info('traverse time: %.06f, detect attack: %s' % (traverse_time, detect_attack))


    def run(self, attack_type, logfile_dir=None, use_tqdm=False, show_summary=True, use_cache=True, build_timeout=180, cache_dir=None):
        # iter_dataset = iter(self.dataset_tx)
        # if use_tqdm:
        #     iter_dataset = tqdm(self.dataset_tx)

        results = pd.DataFrame(columns=['tx_hash', 'traverse_time', 'is_attack', 'detect_attack', 'logfile'])

        for tx_hash in tqdm(self.dataset_tx, disable=not use_tqdm):
            is_attack = (tx_hash in self.attack_tx)
            self.epg_runner.drop_graph()
            # time.sleep(0.5)
            build_time = self.epg_runner.build_graph(tx_hash, use_cache=use_cache, timeout=build_timeout, cache_dir=cache_dir)
            if isinstance(build_time, str):
                results.loc[len(results)] = [tx_hash, None, is_attack, False, None]
                continue

            logfile = None
            if logfile_dir is not None:
                logfile=f'{logfile_dir}/{attack_type}-{tx_hash}.log'
            traverse_time, detected = self.detect_attack(self.epg_runner, attack_type, logfile=logfile)

            results.loc[len(results)] = [tx_hash, traverse_time, is_attack, detected, logfile]

        if show_summary:
            self.summary(results)

        return results
    
    
    def run_parallel(self, attack_type, logfile_dir=None, use_tqdm=False, show_summary=True, use_cache=True, build_timeout=180, cache_dir=None):
        global parallel_config
        parallel_config.attack_type = attack_type
        parallel_config.logfile_dir = logfile_dir
        parallel_config.use_tqdm = use_tqdm
        parallel_config.use_cache = use_cache
        parallel_config.build_timeout = build_timeout
        parallel_config.cache_dir = cache_dir
        parallel_config.epg_runners = self.epg_runners
        parallel_config.attack_tx = self.attack_tx
        parallel_config.eval_instance = self

        results = run_parallel(self.dataset_tx)
        
        if show_summary:
            self.summary(results)
        
        return results


# !!!WARNING!!!: ugly implementation with many side effects
from argparse import Namespace

parallel_config = Namespace(
    attack_type=None,
    logfile_dir=None,
    use_tqdm=False,
    use_cache=True,
    build_timeout=180,
    cache_dir=None,
    attack_tx=None,
    epg_runners=None,
    eval_instance=None,
    ins_queue=None,
)

def task(tx_hash):
    global parallel_config
    
    def run_one(epg_runner: EPGRunner, tx_hash: str):
        '''
        return ('tx_hash', 'traverse_time', 'is_attack', 'detect_attack', 'logfile')
        '''
        # logger.info(tx_hash)
        eval_ins = parallel_config.eval_instance
        is_attack = (tx_hash in eval_ins.attack_tx)
        epg_runner.drop_graph()
        build_time = epg_runner.build_graph(tx_hash, use_cache=parallel_config.use_cache, cache_dir=parallel_config.cache_dir, timeout=parallel_config.build_timeout)
        
        if isinstance(build_time, str):
            return tx_hash, None, is_attack, False, None
        
        logfile = None
        if parallel_config.logfile_dir is not None:
            logfile = f'{parallel_config.logfile_dir}/{parallel_config.attack_type}-{tx_hash}.log'
        traverse_time, detected = eval_ins.detect_attack(epg_runner, parallel_config.attack_type, logfile=logfile)
        
        return tx_hash, traverse_time, is_attack, detected, logfile
    
    # logger.info(parallel_config)
    ins_id = parallel_config.ins_queue.get()
    # logger.info(ins_id)
    res = run_one(parallel_config.epg_runners[ins_id], tx_hash)
    parallel_config.ins_queue.put(ins_id)
    return res

mp.set_start_method('fork')
def run_parallel(tx_list):
    global parallel_config
    results = []
    
    with mp.Manager() as mgr:
        ins_queue = mgr.Queue()
        parallel_config.ins_queue = ins_queue
        for ins_id in range(len(parallel_config.epg_runners)):
            ins_queue.put(ins_id)
        # logger.info(parallel_config)
        
        with mp.Pool(len(parallel_config.epg_runners)) as pool:
            
            for res in tqdm(pool.imap_unordered(task, tx_list, chunksize=5), total=len(tx_list), disable=not parallel_config.use_tqdm):
                results.append(res)
    
    results_df = pd.DataFrame(results, columns=['tx_hash', 'traverse_time', 'is_attack', 'detect_attack', 'logfile'])
    
    return results_df