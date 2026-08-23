package config

import "time"

func (c Config) Example() map[string]string {
	return map[string]string{"listen": c.Listen, "database": c.Database, "scan_interval": c.ScanInterval.String(), "heartbeat_timeout": c.HeartbeatTimeout.String(), "recovery_grace": c.RecoveryGrace.String(), "max_lost_duration": c.MaxLostDuration.String(), "retention": c.Retention.String()}
}
func DurationMap(c Config) map[string]time.Duration {
	return map[string]time.Duration{"scan_interval": c.ScanInterval, "heartbeat_timeout": c.HeartbeatTimeout, "recovery_grace": c.RecoveryGrace, "max_lost_duration": c.MaxLostDuration, "retention": c.Retention}
}
func Merge(base, override Config) Config {
	if override.Listen != "" {
		base.Listen = override.Listen
	}
	if override.Database != "" {
		base.Database = override.Database
	}
	if override.ScanInterval > 0 {
		base.ScanInterval = override.ScanInterval
	}
	if override.HeartbeatTimeout > 0 {
		base.HeartbeatTimeout = override.HeartbeatTimeout
	}
	if override.RecoveryGrace > 0 {
		base.RecoveryGrace = override.RecoveryGrace
	}
	if override.MaxLostDuration > 0 {
		base.MaxLostDuration = override.MaxLostDuration
	}
	if override.Retention > 0 {
		base.Retention = override.Retention
	}
	if override.ConfirmCount > 0 {
		base.ConfirmCount = override.ConfirmCount
	}
	return base
}
