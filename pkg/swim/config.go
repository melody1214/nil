package swim

import (
	"fmt"
	"strings"
	"time"
)

// ServerID is a unique string for identifying servers.
type ServerID string

// ServerAddress is a combination of host:port string of server network address.
type ServerAddress string

// ServerType represents the type of swim node.
type ServerType int

const (
	// MDS : Metadata server.
	MDS ServerType = iota
	// OSD : Object storage daemon.
	OSD
)

// String returns the string of server type.
func (s *ServerType) String() string {
	if *s == MDS {
		return "MDS"
	} else if *s == OSD {
		return "OSD"
	} else {
		return "Unknown"
	}
}

// Config contains configurations for running swim server.
type Config struct {
	// Time interval of generates ping message.
	// Swim server will sends ping periodically with this interval.
	PingPeriod time.Duration

	// Expire time of ping messages.
	PingExpire time.Duration

	// ID is a unique string for identifying servers.
	ID ServerID

	// Address is a combination of host:port string of server network address.
	Address ServerAddress

	// Coordinator is the address of the swim node
	// which will ask to join the cluster.
	Coordinator ServerAddress

	// Type indicates this server is MDS or OSD.
	Type ServerType
}

// DefaultConfig returns a config with the default time settings.
func DefaultConfig() *Config {
	return &Config{
		PingPeriod:  5 * time.Second,
		PingExpire:  5 * time.Second,
		Coordinator: "localhost:51000",
	}
}

func validateConfig(config *Config) error {
	if config.PingPeriod <= 0*time.Second {
		return fmt.Errorf("ping period is too short")
	}

	if config.PingExpire <= 0*time.Second {
		return fmt.Errorf("ping expire time is too short")
	}

	if len(config.ID) == 0 {
		return fmt.Errorf("empty server id")
	}

	if len(strings.Split(string(config.Address), ":")) != 2 {
		return fmt.Errorf("invalid address format")
	}

	if config.Type.String() == "Unknown" {
		return fmt.Errorf("invalid server type")
	}

	return nil
}
