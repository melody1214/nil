package nilrpc

// RecoveryType represents which recovery operation is requested.
type RecoveryType int

const (
	// Recover means do recover.
	Recover RecoveryType = iota
	// Rebalance means do rebalance.
	Rebalance
)

// RecoveryRequest is a request message to recovery domain.
type RecoveryRequest struct {
	Type RecoveryType
}

// RecoveryResponse is a response message from recovery domain.
type RecoveryResponse struct{}
