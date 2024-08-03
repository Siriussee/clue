package dataflow

// memory for every call in a transaction
type memory struct {
	m map[uint64]*DataSource
}

func newMemory() *memory {
	return &memory{
		m: make(map[uint64]*DataSource),
	}
}

func (m *memory) mload(offset uint64) word {
	var w word
	for i := uint64(0); i < 32; i++ {
		w[i] = m.m[offset+i]
	}
	return w
}

func (m *memory) mloadChunk(start uint64, size uint64) []*DataSource {
	var chunk []*DataSource
	for i := start; i < start+size; i++ {
		chunk = append(chunk, m.m[i])
	}

	return chunk
}

func (m *memory) mstore(offset uint64, value word) {
	for i := uint64(0); i < 32; i++ {
		if value[i] != nil {
			m.m[i+offset] = value[i]
		}
	}
}

func (m *memory) mstore8(offset uint64, value *DataSource) {
	if value != nil {
		m.m[offset] = value
	}
}

func (m *memory) mstoreChunk(start uint64, size uint64, value []*DataSource) {
	if len(value) == 0 {
		return
	}
	if len(value) < int(size) {
		size = uint64(len(value))
	}
	for i := uint64(0); i < size; i++ {
		if value[i] != nil {
			m.m[start+i] = value[i]
		}
	}
}
