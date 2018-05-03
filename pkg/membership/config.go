package membership

import (
	"fmt"
	"strings"
	"time"
)

// Config contains configurations for running membership server.
type Config struct {
	// Time interval of generates ping message.
	// Swim server will sends ping periodically with this interval.
	PingPeriod time.Duration

	// Expire time of ping messages.
	PingExpire time.Duration

	// Name is a unique string for identifying node.
	Name NodeName

	// Address is a combination of host:port string of server network address.
	Address NodeAddress

	// Coordinator is the address of the membership server
	// which will ask to join the cluster.
	Coordinator NodeAddress

	// Type means the type of this node.
	Type NodeType
}

// DefaultConfig returns a config with the default time settings.
func DefaultConfig() *Config {
	return &Config{
		PingPeriod:  2000 * time.Millisecond,
		PingExpire:  500 * time.Millisecond,
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

	if len(config.Name) == 0 {
		return fmt.Errorf("empty server name")
	}

	if len(strings.Split(string(config.Address), ":")) != 2 {
		return fmt.Errorf("invalid address format")
	}

	if config.Type.String() == "Unknown" {
		return fmt.Errorf("invalid server type")
	}

	return nil
}
