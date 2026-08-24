package sqlite

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"dhs/internal/api"
	"dhs/internal/model"
	"dhs/internal/service"
	"dhs/internal/store"
)

func newDB(t *testing.T) *DB {
	t.Helper()
	db, err := Open("")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestGetNodeNotFoundPreservesSentinel(t *testing.T) {
	ctx := context.Background()
	db := newDB(t)

	_, err := db.GetNode(ctx, "missing")
	if err == nil {
		t.Fatal("expected not-found error, got nil")
	}
	if !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("errors.Is(store.ErrNotFound) = false; err=%v", err)
	}
	if !store.IsNotFound(err) {
		t.Fatalf("store.IsNotFound = false; err=%v", err)
	}
	if errors.Is(err, store.ErrConflict) {
		t.Fatalf("not-found must not match conflict; err=%v", err)
	}
}

func TestStartRecoveryConflictPreservesSentinel(t *testing.T) {
	ctx := context.Background()
	db := newDB(t)

	// A freshly registered node is not Lost, so StartRecovery must report a
	// conflict (status precondition unmet) rather than a silent no-op.
	n, err := db.Register(ctx, model.RegisterRequest{ID: "n1", Name: "node-1"})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if n.Status != model.Registered {
		t.Fatalf("expected REGISTERED, got %s", n.Status)
	}

	_, err = db.StartRecovery(ctx, "n1", time.Now().UTC())
	if err == nil {
		t.Fatal("expected conflict error, got nil")
	}
	if !errors.Is(err, store.ErrConflict) {
		t.Fatalf("errors.Is(store.ErrConflict) = false; err=%v", err)
	}
	if !store.IsConflict(err) {
		t.Fatalf("store.IsConflict = false; err=%v", err)
	}
}

func TestServiceNodeChainMatchesBothSentinels(t *testing.T) {
	ctx := context.Background()
	svc := &service.Service{Store: newDB(t)}

	_, err := svc.Node(ctx, "missing")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// The service must wrap the store sentinel, so both layers match.
	if !errors.Is(err, service.ErrNotFound) {
		t.Fatalf("errors.Is(service.ErrNotFound) = false; err=%v", err)
	}
	if !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("errors.Is(store.ErrNotFound) = false; err=%v", err)
	}
}

func TestServiceConflictChainMatchesBothSentinels(t *testing.T) {
	ctx := context.Background()
	svc := &service.Service{Store: newDB(t)}

	n, err := svc.Register(ctx, model.RegisterRequest{ID: "n2", Name: "node-2"})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if n.Status != model.Registered {
		t.Fatalf("expected REGISTERED, got %s", n.Status)
	}

	// Recover on a non-Lost/non-Recovering node is an illegal transition.
	_, err = svc.Recover(ctx, "n2")
	if err == nil {
		t.Fatal("expected conflict error, got nil")
	}
	if !errors.Is(err, service.ErrConflict) {
		t.Fatalf("errors.Is(service.ErrConflict) = false; err=%v", err)
	}
	if !errors.Is(err, store.ErrConflict) {
		t.Fatalf("errors.Is(store.ErrConflict) = false; err=%v", err)
	}
}

func TestHTTPStatusMapping(t *testing.T) {
	ctx := context.Background()
	db := newDB(t)
	svc := &service.Service{Store: db}
	a := &api.API{S: svc}
	router := a.Router(http.NewServeMux())

	// not found -> 404
	req := httptest.NewRequest(http.MethodGet, "/api/v1/nodes/missing", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("not found status = %d, want 404", rr.Code)
	}

	// Register a node, then attempt an illegal transition -> 409.
	n, err := svc.Register(ctx, model.RegisterRequest{ID: "n3", Name: "node-3"})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if n.Status != model.Registered {
		t.Fatalf("expected REGISTERED, got %s", n.Status)
	}
	req = httptest.NewRequest(http.MethodPost, "/api/v1/nodes/n3/recover", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusConflict {
		t.Fatalf("recover status = %d, want 409", rr.Code)
	}

	// A closed store surfaces a non-sentinel DB error -> 500.
	db.Close()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/nodes/missing", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("closed-db status = %d, want 500", rr.Code)
	}
}
