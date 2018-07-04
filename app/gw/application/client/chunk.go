package client

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/rpc"
	"net/url"
	"strconv"
	"time"

	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/nilrpc"
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

	c := h.cmapAPI.SearchCall()
	eg, err := c.EncGrp().ID(cmap.ID(iEncGrp)).Status(cmap.EGAlive).Do()
	if err != nil {
		http.Error(w, "invalid header", http.StatusBadRequest)
		return
	}

	if len(eg.Vols) < iShardNum {
		http.Error(w, "invalid header", http.StatusBadRequest)
		return
	}

	v, err := c.Volume().ID(eg.Vols[iShardNum].ID).Status(cmap.VolActive).Do()
	if err != nil {
		http.Error(w, "invalid header", http.StatusBadRequest)
		return
	}
	r.Header.Add("Volume", v.ID.String())

	n, err := c.Node().ID(cmap.ID(v.Node)).Status(cmap.NodeAlive).Do()
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

func (h *handlers) RenameChunkHandler(w http.ResponseWriter, r *http.Request) {
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

	oldChunkName := r.Header.Get("Old-Chunk-Name")
	if oldChunkName == "" {
		http.Error(w, "invalid header", http.StatusBadRequest)
		return
	}

	newChunkName := r.Header.Get("New-Chunk-Name")
	if newChunkName == "" {
		http.Error(w, "invalid header", http.StatusBadRequest)
		return
	}

	c := h.cmapAPI.SearchCall()
	eg, err := c.EncGrp().ID(cmap.ID(iEncGrp)).Status(cmap.EGAlive).Do()
	if err != nil {
		http.Error(w, "invalid header", http.StatusBadRequest)
		return
	}

	for _, egv := range eg.Vols {
		v, err := c.Volume().ID(egv.ID).Status(cmap.VolActive).Do()
		if err != nil {
			http.Error(w, "invalid header", http.StatusBadRequest)
			return
		}

		n, err := c.Node().ID(cmap.ID(v.Node)).Status(cmap.NodeAlive).Do()
		if err != nil {
			http.Error(w, "invalid header", http.StatusBadRequest)
			return
		}

		conn, err := nilrpc.Dial(n.Addr.String(), nilrpc.RPCNil, time.Duration(2*time.Second))
		if err != nil {
			http.Error(w, "invalid header", http.StatusBadRequest)
			return
		}
		defer conn.Close()

		req := &nilrpc.DGERenameChunkRequest{
			Vol:      egv.ID.String(),
			EncGrp:   eg.ID.String(),
			OldChunk: oldChunkName,
			NewChunk: newChunkName,
		}
		res := &nilrpc.DGERenameChunkResponse{}

		cli := rpc.NewClient(conn)
		if err := cli.Call(nilrpc.DsGencodingRenameChunk.String(), req, res); err != nil {
			http.Error(w, "invalid header", http.StatusBadRequest)
			return
		}
	}
}
