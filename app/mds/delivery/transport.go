package delivery

// rpcTypeBytes returns rpc type bytes which is used to multiplexing.
func rpcTypeBytes() []byte {
	return []byte{
		0x02, // rpcNil
	}
}

// raftTypeBytes returns rpc type bytes which is used to raft protocol.
func raftTypeBytes() []byte {
	return []byte{
		0x01, // rpcRaft
	}
}

// membershipTypeBytes returns rpc type bytes which is used to multiplexing.
func membershipTypeBytes() []byte {
	return []byte{
		0x03, // rpcSwim
	}
}
