package mds

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/chanyoung/nil/pkg/db"
)

// Mds is the [project name] meta-data server node.
type Mds struct {
	db db.DB
}

// New creates a mds object.
func New() (*Mds, error) {
	m := &Mds{
		db: db.New(),
	}

	return m, nil
}

// Start starts the mds.
func (m *Mds) Start() {
	// Wait until Ctrl-C or other terminate signal is received.
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	m.stop()
}

func (m *Mds) stop() {
	// Clean up mds works.
	fmt.Println("stop mds")
}
