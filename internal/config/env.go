package config

import (
	"os"
	"strconv"
	"time"
)

func applyDuration(key string, dst *time.Duration) error {
	if v := os.Getenv(key); v != "" {
		d, e := time.ParseDuration(v)
		if e != nil {
			return e
		}
		*dst = d
	}
	return nil
}
func applyInt(key string, dst *int) error {
	if v := os.Getenv(key); v != "" {
		n, e := strconv.Atoi(v)
		if e != nil {
			return e
		}
		*dst = n
	}
	return nil
}
func ApplyEnvironment(c *Config) error {
	if v := os.Getenv("DHS_LISTEN"); v != "" {
		c.Listen = v
	}
	if v := os.Getenv("DHS_DATABASE"); v != "" {
		c.Database = v
	}
	if e := applyDuration("DHS_SCAN_INTERVAL", &c.ScanInterval); e != nil {
		return e
	}
	if e := applyDuration("DHS_HEARTBEAT_TIMEOUT", &c.HeartbeatTimeout); e != nil {
		return e
	}
	if e := applyDuration("DHS_RECOVERY_GRACE", &c.RecoveryGrace); e != nil {
		return e
	}
	if e := applyDuration("DHS_MAX_LOST_DURATION", &c.MaxLostDuration); e != nil {
		return e
	}
	if e := applyDuration("DHS_RETENTION", &c.Retention); e != nil {
		return e
	}
	return applyInt("DHS_CONFIRM_COUNT", &c.ConfirmCount)
}
