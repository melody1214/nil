package mlog

import (
	"os"
	"sync"

	"github.com/sirupsen/logrus"
)

var (
	l log
)

// log wraps logrus.Logger and holds information of logging file.
type log struct {
	*logrus.Logger

	file     *os.File
	location string
	mu       sync.Mutex
}

// Init creates log object.
// TODO: logging with linux logrotate.
func Init(location string) error {
	l.Logger = logrus.New()
	l.location = location

	if l.location == "stderr" {
		l.Out = os.Stderr
		l.file = nil
	} else {
		f, err := os.OpenFile(location, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			return err
		}
		l.Out = f
		l.file = f
	}

	return nil
}

// GetLogger returns static logger variable.
// Logging with this logger at any packages.
func GetLogger() *logrus.Logger {
	return l.Logger
}
