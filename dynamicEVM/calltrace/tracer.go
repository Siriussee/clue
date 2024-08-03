package calltrace

import "github.com/anon/execution-property-graph/dynamicEVM/types"

type CallIdTracer interface {
	GetCurrentCallId() types.CallId
	GetCurrentCallCount() int
}
