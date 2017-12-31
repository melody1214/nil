package swim

// Status is the status of the node which managemented by swim protocol.
type Status int

const (
	// Alive : healthy node
	Alive Status = 0
	// Suspect : maybe faulty
	Suspect = 1
	// Faulty : faulty
	Faulty = 2
)

// String converts integer member type to string.
func (s *Status) String() string {
	if *s == Alive {
		return "Alive"
	} else if *s == Suspect {
		return "Suspect"
	} else if *s == Faulty {
		return "Faulty"
	}
	return "Unknown"
}

// MemberType indicates the type of the swim member node.
type MemberType int

const (
	// MDS : mds node
	MDS MemberType = 0
	// OSD : osd node
	OSD = 1
)

// String converts integer member type to string.
func (m *MemberType) String() string {
	if *m == MDS {
		return "MDS"
	} else if *m == OSD {
		return "OSD"
	}
	return "Unknown"
}

// Member contains the node information about swim node.
type Member struct {
	UUID        string
	Type        MemberType
	Addr        string
	Port        string
	Status      Status
	Incarnation uint32
}
