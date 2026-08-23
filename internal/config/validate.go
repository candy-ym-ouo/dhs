package config

import "fmt"

func (c Config) Validate() error {
	if c.Listen == "" {
		return fmt.Errorf("listen cannot be empty")
	}
	if c.Database == "" {
		return fmt.Errorf("database cannot be empty")
	}
	if c.ScanInterval <= 0 || c.HeartbeatTimeout <= 0 {
		return fmt.Errorf("scan and heartbeat durations must be positive")
	}
	if c.RecoveryGrace <= 0 || c.MaxLostDuration <= 0 || c.Retention <= 0 {
		return fmt.Errorf("retention durations must be positive")
	}
	if c.ConfirmCount < 1 {
		return fmt.Errorf("confirm_count must be at least 1")
	}
	return nil
}
