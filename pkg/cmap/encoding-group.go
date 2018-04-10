package cmap

// EncodingGroup is the logical group for making local parity.
type EncodingGroup struct {
	ID   ID    `xml:"id"`
	Size int64 `xml:"size"`
	Used int64 `xml:"used"`
	Free int64 `xml:"free"`
	Vols []ID  `xml:"volumes"`
}
