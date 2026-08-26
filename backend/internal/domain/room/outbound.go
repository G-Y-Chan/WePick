package room

// OutboundEvent is what actually gets pushed to a connected realtime Client.
// Distinct from Event: Event is the cross-process bus message; OutboundEvent
// is the per-connection push payload derived from it.
type OutboundType string

const (
	OutboundStart         OutboundType = "START"
	OutboundMajorityFound OutboundType = "MAJORITY_FOUND"
	OutboundError         OutboundType = "ERROR"
)

// OutboundEvent is a single push payload destined for a connected realtime Client.
type OutboundEvent struct {
	Type    OutboundType
	CardID  string
	Message string // populated only for OutboundError
}
