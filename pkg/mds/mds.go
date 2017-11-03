package mds

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/chanyoung/nil/pkg/db"
	"github.com/chanyoung/nil/pkg/util/uuid"
)

// Mds is the [project name] meta-data server node.
type Mds struct {
	id     uuid.UUID
	server *server
	db     db.DB
}

// New creates a mds object.
func New(addr, port string) (*Mds, error) {
	m := &Mds{
		id:     uuid.Gen(),
		server: newServer(addr, port),
		db:     db.New(),
	}

	return m, nil
}

// Start starts the mds.
func (m *Mds) Start() {
	// Make channel for Ctrl-C or other terminate signal is received.
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	// Make channel for message from server.
	mc := make(chan error, 1)
	go m.server.start(mc)

	for {
		select {
		case <-sc:
			m.stop()
			return
		case err := <-mc:
			fmt.Println(err)
		}
	}
}

func (m *Mds) stop() {
	// Clean up mds works.
	fmt.Println("stop mds")
}
