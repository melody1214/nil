package membership

import (
	"fmt"
	"testing"
	"time"
)

func TestConfig(t *testing.T) {
	testCases := []struct {
		cfg      *Config
		expected error
	}{
		{
			&Config{
				Name:       NodeName("TEST_SRV"),
				PingPeriod: 0 * time.Second,
			},
			fmt.Errorf("ping period is too short"),
		},
		{
			&Config{
				Name:       NodeName("TEST_SRV"),
				PingPeriod: 5 * time.Second,
				PingExpire: 0 * time.Second,
			},
			fmt.Errorf("ping expire time is too short"),
		},
		{
			&Config{
				Name:       NodeName("TEST_SRV"),
				PingPeriod: 5 * time.Second,
				PingExpire: 5 * time.Second,
				Address:    NodeAddress("address"),
			},
			fmt.Errorf("invalid address format"),
		},
		{
			&Config{
				Name:       NodeName("TEST_SRV"),
				PingPeriod: 5 * time.Second,
				PingExpire: 5 * time.Second,
				Address:    NodeAddress("localhost:8000"),
				Type:       NodeType(-1),
			},
			fmt.Errorf("invalid server type"),
		},
	}

	for _, c := range testCases {
		if err := validateConfig(c.cfg); err.Error() != c.expected.Error() {
			t.Errorf("expected error %v but got %v", c.expected, err)
		}
	}
}
