package recovery

type Volume struct {
	ID         int
	Status     string
	NodeID     int
	Used       int
	Chain      int
	MaxChain   int
	Unbalanced bool
}

// ByFreeChain for sorting volumes by free chain.
type ByFreeChain []*Volume

func (c ByFreeChain) Len() int           { return len(c) }
func (c ByFreeChain) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c ByFreeChain) Less(i, j int) bool { return c[i].MaxChain-c[i].Chain < c[j].MaxChain-c[j].Chain }

type localChain struct {
	id        int
	status    string
	firstVol  int
	secondVol int
	thirdVol  int
	parityVol int
}
