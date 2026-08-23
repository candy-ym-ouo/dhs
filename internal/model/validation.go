package model

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var idPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$`)

func (r RegisterRequest) Validate() error {
	if !idPattern.MatchString(r.ID) {
		return errors.New("id must contain 1-128 safe characters")
	}
	if strings.TrimSpace(r.Name) == "" {
		return errors.New("name is required")
	}
	if len(r.Name) > 200 {
		return errors.New("name is too long")
	}
	if len(r.Address) > 255 {
		return errors.New("address is too long")
	}
	if len(r.TaskTypes) > 100 {
		return errors.New("too many task types")
	}
	for _, t := range r.TaskTypes {
		if strings.TrimSpace(t) == "" {
			return errors.New("task type cannot be empty")
		}
	}
	for k, v := range r.Labels {
		if strings.TrimSpace(k) == "" || len(k) > 64 || len(v) > 256 {
			return fmt.Errorf("invalid label %q", k)
		}
	}
	return nil
}

func (r HeartbeatRequest) Validate() error {
	if r.Load < 0 || r.Load > 1 {
		return errors.New("load must be between 0 and 1")
	}
	if len(r.Extra) > 64 {
		return errors.New("too many extra fields")
	}
	return nil
}
