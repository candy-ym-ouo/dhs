package service

import (
	"errors"
	"testing"
)

func TestBug003ErrorIdentitySurvivesServiceBoundary(t *testing.T) {
	sentinel := errors.New("storage unavailable")
	if errors.Is(flattenError(sentinel), sentinel) == false { t.Fatalf("service error lost its identity") }
}
