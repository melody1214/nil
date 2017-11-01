package osd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

type Osd struct {
}

// New creates a osd object.
func New() (*Osd, error) {
	o := &Osd{}

	return o, nil
}

// Start starts the osd.
func (o *Osd) Start() {
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
