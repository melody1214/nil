package membership

// EncodingGroupStatus is the status of the node.
type EncodingGroupStatus string

const (
	// EGAlive : healthy node
	EGAlive EncodingGroupStatus = "Alive"
	// EGSuspect : maybe faulty
	EGSuspect = "Suspect"
	// EGFaulty : faulty
	EGFaulty = "Faulty"
	// EGRdonly : readonly, maybe rebalancing or recovering.
	EGRdonly = "Rdonly"
)

// String returns a string of the node status.
func (s EncodingGroupStatus) String() string {
	switch s {
	case EGAlive, EGSuspect, EGFaulty, EGRdonly:
		return string(s)
	default:
		return unknown
	}
}

// EncodingGroup is the logical group for making local parity.
type EncodingGroup struct {
	ID   ID                  `xml:"id"`
	Incr Incarnation         `xml:"incarnation"`
	Stat EncodingGroupStatus `xml:"status"`
	Size int64               `xml:"size"`
	Used int64               `xml:"used"`
	Free int64               `xml:"free"`
	Vols []ID                `xml:"volume"`
}
