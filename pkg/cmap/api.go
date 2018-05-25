package cmap

// Service is the root manager of membership package.
// The service consists of four parts described below.
//
// 1. Server
// The server is a membership management server based on the swim membership
// protocol. It sends new updates to randomly selected nodes and updates its
// membership information.
//
// 2. Cluster map
// Cluster map contains the information of each node, volume, encoding group
// and etc. It is versioned for every significant changes are occurred.
//
// 3. Slave cluster map api
// Slave cluster map api provides functions to search various elements of
// cluster map with the given conditions. It also provides the functions
// to update volume status or capacity information. (for DS functions)
//
// 4. Master cluster map api
// Master cluster map api is the superset of slave api. Additionally provides add,
// remove, update node functions and all of this kind of changes will increment
// the version number of the cluster map. (for MDS functions)
type Service struct {
	// Configuration provided at service initialization.
	cfg Config

	// Managing cmap with membership server and client APIs.
	manager *manager

	// Membership protocol server.
	server *server
}

// MasterAPI is the interface for access the membership service with master mode.
type MasterAPI interface {
	SearchCall() *SearchCall
	SearchCallNode() *SearchCallNode
	SearchCallVolume() *SearchCallVolume
	SearchCallEncGrp() *SearchCallEncGrp
	GetLatestCMap() CMap
	UpdateCMap(cmap *CMap) error
	GetStateChangedNoti() <-chan interface{}
	GetLatestCMapVersion() Version
	GetUpdatedNoti(ver Version) <-chan interface{}
}

// SlaveAPI is the interface for access the membership service with slave mode.
type SlaveAPI interface {
	SearchCall() *SearchCall
	SearchCallNode() *SearchCallNode
	SearchCallVolume() *SearchCallVolume
	SearchCallEncGrp() *SearchCallEncGrp
	UpdateNodeStatus(nID ID, stat NodeStatus) error
	UpdateVolume(volume Volume) error
	UpdateEncodingGroupStatus(egID ID, stat EncodingGroupStatus) error
	UpdateEncodingGroupUsed(egID ID, used uint64) error
	GetLatestCMapVersion() Version
	GetLatestCMap() CMap
	GetUpdatedNoti(ver Version) <-chan interface{}
	FindEncodingGroupByLeader(leaderNode ID) []EncodingGroup
	UpdateEncodingGroupUnencoded(eg EncodingGroup) error
}

// MasterAPI returns a set of APIs that can be used by nodes in master mode.
func (s *Service) MasterAPI() MasterAPI {
	return s
}

// SlaveAPI returns a set of APIs that can be used by nodes in slave mode.
func (s *Service) SlaveAPI() SlaveAPI {
	return s
}

// SearchCall returns a SearchCall object which can support convenient
// searching some members in the cluster.
func (s *Service) SearchCall() *SearchCall {
	return &SearchCall{
		// Use copied the latest cluster map.
		cmap: s.manager.LatestCMap(),
	}
}
