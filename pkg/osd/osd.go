package osd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/chanyoung/nil/pkg/rest"
)

type Osd struct {
	restServer *rest.Server
}

// New creates a osd object.
func New() (*Osd, error) {
	o := &Osd{
		restServer: rest.NewServer(),
	}

	return o, nil
}

// Start starts the osd.
func (o *Osd) Start() {
	if err := o.restServer.Start(); err != nil {
		log.Fatal(err)
	}

	// Wait until Ctrl-C or other terminate signal is received.
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	o.stop()
}

func (o *Osd) stop() {
	// Clean up osd works.
	fmt.Println("stop osd")
}
