package types

import (
	"strconv"
	"strings"
)

type CallId []int

func NewCallId(callFrameIds []int) CallId {
	id := make([]int, len(callFrameIds))
	copy(id, callFrameIds)
	return CallId(id)
}

func StringToCallId(id string) (CallId, error) {
	// "0:1:2:3" -> [1, 2, 3]
	// Leading 0 is for the root call frame.
	if id == "0" {
		return CallId{}, nil
	}
	parts := strings.Split(id, ":")
	callFrameIds := make([]int, len(parts)-1)
	for i, part := range parts[1:] {
		callFrameId, err := strconv.Atoi(part)
		if err != nil {
			return nil, err
		}
		callFrameIds[i] = callFrameId
	}
	return CallId(callFrameIds), nil
}

func (id CallId) Depth() int {
	return len(id)
}

func (id CallId) IsRoot() bool {
	return len(id) == 0
}

func (id CallId) Parent() CallId {
	if len(id) == 0 {
		return nil
	}
	return id[:len(id)-1]
}

func (id CallId) String() string {
	// [1, 2, 3] -> "0:1:2:3"
	// Leading 0 is for the root call frame.
	if len(id) == 0 {
		return "0"
	}
	str := "0:"
	for i := 0; i < len(id)-1; i++ {
		str += strconv.Itoa(id[i]) + ":"
	}
	str += strconv.Itoa(id[len(id)-1])
	return str
}

func (id CallId) Compare(other CallId) int {
	for i := 0; i < len(id) && i < len(other); i++ {
		if id[i] < other[i] {
			return -1
		} else if id[i] > other[i] {
			return 1
		}
	}
	if len(id) < len(other) {
		return -1
	} else if len(id) > len(other) {
		return 1
	}
	return 0
}

func (id CallId) CommonParent(other CallId) CallId {
	i := 0
	for i < len(id) && i < len(other) && id[i] == other[i] {
		i++
	}
	return id[:i]
}

type DcfgId []int

func NewDcfgId(callId CallId, dcfgId int, callCount int) DcfgId {
	id := make(DcfgId, len(callId)+2)
	copy(id, callId)
	id[len(id)-2] = dcfgId
	id[len(id)-1] = callCount
	return id
}

func (id DcfgId) String() string {
	return NewCallId(id[:len(id)-2]).String() + "-" + strconv.Itoa(id[len(id)-2]) + "-" + strconv.Itoa(id[len(id)-1])
}

func (id DcfgId) CallId() CallId {
	return NewCallId(id[:len(id)-2])
}

func (id DcfgId) DcfgId() int {
	return id[len(id)-2]
}

func (id DcfgId) CallCount() int {
	return id[len(id)-1]
}

func StringToDcfgId(id string) (DcfgId, error) {
	parts := strings.Split(id, "-")
	callId, err := StringToCallId(parts[0])
	if err != nil {
		return nil, err
	}
	dcfgId := make(DcfgId, len(callId)+2)
	copy(dcfgId, callId)
	dcfgId[len(dcfgId)-2], err = strconv.Atoi(parts[1])
	if err != nil {
		return nil, err
	}
	dcfgId[len(dcfgId)-1], err = strconv.Atoi(parts[2])
	if err != nil {
		return nil, err
	}
	return dcfgId, nil
}

func toCompareValue(x int) int {
	if x < 0 {
		return -1
	}
	if x > 0 {
		return 1
	}
	return 0
}

func (id DcfgId) Compare(other DcfgId) int {
	callId := id.CallId()
	otherCallId := other.CallId()
	commonParent := callId.CommonParent(otherCallId)
	// One is the other's parent
	if len(commonParent) == len(callId) || len(commonParent) == len(otherCallId) {
		// Inside the same call
		if len(callId) == len(otherCallId) {
			return toCompareValue(id.DcfgId() - other.DcfgId())
		}
		// Current call is deeper
		if len(callId) > len(otherCallId) {
			if callId[len(otherCallId)] < other.CallCount() {
				return -1
			}
			return 1
		}
		// Other call is deeper
		if id.CallCount() <= otherCallId[len(callId)] {
			return -1
		}
		return 1
	}
	// Normal case
	return callId.Compare(otherCallId)
}
