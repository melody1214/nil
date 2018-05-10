package recovery

// // Volume is an entity that represents a specific volume attached to ds.
// // It has some attributes for rebalancing.
// type Volume struct {
// 	cmap.Volume
// 	NodeID   int
// 	Chain    int
// 	MaxChain int
// }

// // isUnbalanced checks if the volume has unbalanced encoding group ratio.
// func (v Volume) isUnbalanced() bool {
// 	if v.MaxChain == 0 {
// 		return false
// 	}

// 	if v.Chain == 0 {
// 		return true
// 	}

// 	return (v.Chain*100)/v.MaxChain < 70
// }

// // ByFreeChain for sorting volumes by free chain.
// type ByFreeChain []*Volume

// func (c ByFreeChain) Len() int           { return len(c) }
// func (c ByFreeChain) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
// func (c ByFreeChain) Less(i, j int) bool { return c[i].MaxChain-c[i].Chain < c[j].MaxChain-c[j].Chain }
