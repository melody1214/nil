package mlog

import (
	"os"
	"sync"

	"github.com/sirupsen/logrus"
)

// Log wraps logrus.Logger and holds information of logging file.
type Log struct {
	*logrus.Logger

	file     *os.File
	location string
	mu       sync.Mutex
}

// New creates Log object.
// TODO: logging with linux logrotate.
func New(location string) (*Log, error) {
	l := &Log{}

	l.Logger = logrus.New()
	l.location = location

	if l.location == "stderr" {
		l.Out = os.Stderr
		l.file = nil
	} else {
		f, err := os.OpenFile(location, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			return nil, err
		}
		l.Out = f
		l.file = f
	}

	return l, nil
}
