package client

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"

	"github.com/chanyoung/nil/pkg/cmap"
)

// GetChunkHandler handles the client request for downloading a chunk.
func (h *handlers) GetChunkHandler(w http.ResponseWriter, r *http.Request) {
	encGrp := r.Header.Get("Encoding-Group")
	if encGrp == "" {
		http.Error(w, "invalid header", http.StatusBadRequest)
		return
	}
	iEncGrp, err := strconv.ParseInt(encGrp, 10, 64)
	if err != nil {
		http.Error(w, "invalid header", http.StatusBadRequest)
		return
	}

	chunkName := r.Header.Get("Chunk-Name")
	if chunkName == "" {
		http.Error(w, "invalid header", http.StatusBadRequest)
		return
	}

	shardNum := r.Header.Get("Shard-Number")
	if shardNum == "" {
		http.Error(w, "invalid header", http.StatusBadRequest)
		return
	}
	iShardNum, err := strconv.Atoi(shardNum)
	if err != nil {
		http.Error(w, "invalid header", http.StatusBadRequest)
		return
	}

	eg, err := h.cmapAPI.SearchCallEncGrp().ID(cmap.ID(iEncGrp)).Status(cmap.EGAlive).Do()
	if err != nil {
		http.Error(w, "invalid header", http.StatusBadRequest)
		return
	}

	if len(eg.Vols) < iShardNum {
		http.Error(w, "invalid header", http.StatusBadRequest)
		return
	}

	v, err := h.cmapAPI.SearchCallVolume().ID(eg.Vols[iShardNum]).Status(cmap.Active).Do()
	if err != nil {
		http.Error(w, "invalid header", http.StatusBadRequest)
		return
	}
	r.Header.Add("Volume", v.ID.String())

	n, err := h.cmapAPI.SearchCallNode().ID(cmap.ID(v.Node)).Status(cmap.Alive).Do()
	if err != nil {
		http.Error(w, "invalid header", http.StatusBadRequest)
		return
	}

	rpURL, err := url.Parse("https://" + n.Addr.String())
	if err != nil {
		http.Error(w, "invalid header", http.StatusBadRequest)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(rpURL)
	proxy.ErrorLog = log.New(logger.Writer(), "http reverse proxy", log.Lshortfile)
	proxy.ServeHTTP(w, r)
}
