from functools import total_ordering
from typing import List

'''
TODO:
- node label enumeration
- edge label enumeration
'''

@total_ordering
class DcfgId:
    def __init__(self, call_id: List[int], idx: int, call_cnt: int):
        self.call_id = call_id
        self.idx = idx
        self.call_cnt = call_cnt
    
    @classmethod
    def from_str(cls, dcfg_id_str: str):
        '''
        dcfg_id: {call_id}-{idx}-{call_cnt}
        call_id: x:x:x:x
        '''
        call_id, idx, call_cnt = dcfg_id_str.split('-')
        call_id = [int(x) for x in call_id.split(':')]
        idx = int(idx)
        call_cnt = int(call_cnt)
        return cls(call_id, idx, call_cnt)
    
    def __lt__(self, other):
        for i1, i2 in zip(self.call_id, other.call_id):
            if i1 < i2:
                return True
            elif i1 > i2:
                return False
        
        if len(self.call_id) == len(other.call_id):
            return self.idx < other.idx
        
        if len(self.call_id) > len(other.call_id):
            return other.call_cnt > self.call_id[len(other.call_id)]
        elif len(self.call_id) < len(other.call_id):
            return self.call_cnt <= other.call_id[len(self.call_id)]
        
        raise RuntimeError('Never Reached')
    
    def __str__(self):
        return f"{':'.join(str(x) for x in self.call_id)}-{self.idx}-{self.call_cnt}"