package config

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	Listen                                                                    string
	Database                                                                  string
	ScanInterval, HeartbeatTimeout, RecoveryGrace, MaxLostDuration, Retention time.Duration
	ConfirmCount                                                              int
}

func Default() Config {
	return Config{Listen: ":8080", Database: "./data/heartbeat.db", ScanInterval: 5 * time.Second, HeartbeatTimeout: 30 * time.Second, RecoveryGrace: 60 * time.Second, MaxLostDuration: 30 * time.Minute, Retention: 7 * 24 * time.Hour, ConfirmCount: 1}
}
func Load(path string) (Config, error) {
	c := Default()
	if path != "" {
		b, e := os.ReadFile(path)
		if e != nil {
			return c, e
		}
		for _, line := range strings.Split(string(b), "\n") {
			p := strings.SplitN(strings.TrimSpace(line), ":", 2)
			if len(p) != 2 {
				continue
			}
			k, v := strings.TrimSpace(p[0]), strings.Trim(strings.TrimSpace(p[1]), "\"'")
			switch k {
			case "listen":
				c.Listen = v
			case "database":
				c.Database = v
			case "scan_interval":
				d, e := time.ParseDuration(v)
				if e != nil {
					return c, e
				}
				c.ScanInterval = d
			case "heartbeat_timeout":
				d, e := time.ParseDuration(v)
				if e != nil {
					return c, e
				}
				c.HeartbeatTimeout = d
			case "recovery_grace":
				d, e := time.ParseDuration(v)
				if e != nil {
					return c, e
				}
				c.RecoveryGrace = d
			case "max_lost_duration":
				d, e := time.ParseDuration(v)
				if e != nil {
					return c, e
				}
				c.MaxLostDuration = d
			case "retention":
				d, e := time.ParseDuration(v)
				if e != nil {
					return c, e
				}
				c.Retention = d
			case "confirm_count":
				fmt.Sscan(v, &c.ConfirmCount)
			}
		}
	}
	if e := ApplyEnvironment(&c); e != nil {
		return c, e
	}
	if e := c.Validate(); e != nil {
		return c, e
	}
	return c, nil
}
func FromFlags() (Config, error) {
	p := flag.String("config", "config.yaml", "configuration file")
	flag.Parse()
	return Load(*p)
}
