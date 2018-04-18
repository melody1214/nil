package nilrpc

// RecoveryType represents which recovery operation is requested.
type RecoveryType int

const (
	// Recover means do recover.
	Recover RecoveryType = iota
	// Rebalance means do rebalance.
	Rebalance
)

// MRERecoveryRequest is a request message to recovery domain.
type MRERecoveryRequest struct {
	Type RecoveryType
}

// MRERecoveryResponse is a response message from recovery domain.
type MRERecoveryResponse struct{}
