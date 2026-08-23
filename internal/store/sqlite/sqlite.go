package sqlite

import (
	"context"
	"database/sql"
	"dhs/internal/model"
	"dhs/internal/store"
	"encoding/json"
	"fmt"
	_ "modernc.org/sqlite"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type DB struct{ db *sql.DB }

func Open(path string) (*DB, error) {
	if path != "" && path != ":memory:" {
		if e := os.MkdirAll(filepath.Dir(path), 0755); e != nil {
			return nil, e
		}
	}
	d, e := sql.Open("sqlite", path)
	if e != nil {
		return nil, e
	}
	d.SetMaxOpenConns(1)
	x := &DB{d}
	if e = x.migrate(); e != nil {
		d.Close()
		return nil, e
	}
	return x, nil
}
func (s *DB) Close() error { return s.db.Close() }
func (s *DB) migrate() error {
	_, e := s.db.Exec(`CREATE TABLE IF NOT EXISTS nodes(id TEXT PRIMARY KEY,name TEXT,address TEXT,labels TEXT,task_types TEXT,version TEXT,status TEXT NOT NULL,last_heartbeat_at TEXT,registered_at TEXT NOT NULL,lost_at TEXT,recover_attempts INTEGER NOT NULL DEFAULT 0); CREATE TABLE IF NOT EXISTS heartbeats(id INTEGER PRIMARY KEY AUTOINCREMENT,node_id TEXT,load REAL,extra TEXT,reported_at TEXT); CREATE TABLE IF NOT EXISTS state_transitions(id INTEGER PRIMARY KEY AUTOINCREMENT,node_id TEXT,from_status TEXT,to_status TEXT,reason TEXT,trigger TEXT,detail TEXT,created_at TEXT); CREATE INDEX IF NOT EXISTS idx_hb_node ON heartbeats(node_id,reported_at); CREATE INDEX IF NOT EXISTS idx_tr_node ON state_transitions(node_id,created_at)`)
	return e
}
func enc(v any) string    { b, _ := json.Marshal(v); return string(b) }
func dec(s string, v any) { _ = json.Unmarshal([]byte(s), v) }
func scanNode(r interface{ Scan(...any) error }) (n model.Node, e error) {
	var l, t string
	var hb, lost, reg sql.NullString
	e = r.Scan(&n.ID, &n.Name, &n.Address, &l, &t, &n.Version, &n.Status, &hb, &reg, &lost, &n.RecoverAttempts)
	if e != nil {
		return
	}
	dec(l, &n.Labels)
	dec(t, &n.TaskTypes)
	n.RegisteredAt, _ = time.Parse(time.RFC3339Nano, reg.String)
	if hb.Valid {
		x, _ := time.Parse(time.RFC3339Nano, hb.String)
		n.LastHeartbeatAt = &x
	}
	if lost.Valid {
		x, _ := time.Parse(time.RFC3339Nano, lost.String)
		n.LostAt = &x
	}
	return
}

var nodeCols = "id,name,address,labels,task_types,version,status,last_heartbeat_at,registered_at,lost_at,recover_attempts"

func (s *DB) Register(ctx context.Context, r model.RegisterRequest) (model.Node, error) {
	now := time.Now().UTC()
	labels, tasks := enc(r.Labels), enc(r.TaskTypes)
	var previous model.Status
	err := s.db.QueryRowContext(ctx, "SELECT status FROM nodes WHERE id=?", r.ID).Scan(&previous)
	isNew := err == sql.ErrNoRows
	if err != nil && !isNew {
		return model.Node{}, err
	}
	_, e := s.db.ExecContext(ctx, `INSERT INTO nodes(id,name,address,labels,task_types,version,status,registered_at) VALUES(?,?,?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET name=excluded.name,address=excluded.address,labels=excluded.labels,task_types=excluded.task_types,version=excluded.version,status=CASE WHEN nodes.status=? THEN excluded.status ELSE nodes.status END,lost_at=CASE WHEN nodes.status=? THEN NULL ELSE nodes.lost_at END`, r.ID, r.Name, r.Address, labels, tasks, r.Version, model.Registered, now.Format(time.RFC3339Nano), model.Offline, model.Offline)
	if e != nil {
		return model.Node{}, e
	}
	if isNew {
		_, e = s.db.ExecContext(ctx, "INSERT INTO state_transitions(node_id,from_status,to_status,reason,trigger,detail,created_at) VALUES(?,?,?,?,?,?,?)", r.ID, "", model.Registered, "register", "api", "", now.Format(time.RFC3339Nano))
	} else if previous == model.Offline {
		_, e = s.db.ExecContext(ctx, "INSERT INTO state_transitions(node_id,from_status,to_status,reason,trigger,detail,created_at) VALUES(?,?,?,?,?,?,?)", r.ID, previous, model.Registered, "re_register", "api", "", now.Format(time.RFC3339Nano))
	}
	if e != nil {
		return model.Node{}, e
	}
	return s.GetNode(ctx, r.ID)
}
func (s *DB) GetNode(ctx context.Context, id string) (model.Node, error) {
	return scanNode(s.db.QueryRowContext(ctx, "SELECT "+nodeCols+" FROM nodes WHERE id=?", id))
}
func (s *DB) ListNodes(ctx context.Context, f store.Filter) ([]model.Node, int, error) {
	where := []string{"1=1"}
	args := []any{}
	if f.Status != "" {
		where = append(where, "status=?")
		args = append(args, f.Status)
	}
	if f.Keyword != "" {
		where = append(where, "(id LIKE ? OR name LIKE ?)")
		q := "%" + f.Keyword + "%"
		args = append(args, q, q)
	}
	var total int
	if e := s.db.QueryRowContext(ctx, "SELECT count(*) FROM nodes WHERE "+strings.Join(where, " AND "), args...).Scan(&total); e != nil {
		return nil, 0, e
	}
	if f.PageSize <= 0 {
		f.PageSize = 20
	}
	if f.Page <= 0 {
		f.Page = 1
	}
	args = append(args, f.PageSize, (f.Page-1)*f.PageSize)
	rows, e := s.db.QueryContext(ctx, "SELECT "+nodeCols+" FROM nodes WHERE "+strings.Join(where, " AND ")+" ORDER BY id LIMIT ? OFFSET ?", args...)
	if e != nil {
		return nil, 0, e
	}
	defer rows.Close()
	var out []model.Node
	for rows.Next() {
		n, e := scanNode(rows)
		if e != nil {
			return nil, 0, e
		}
		out = append(out, n)
	}
	return out, total, rows.Err()
}
func (s *DB) RecordHeartbeat(ctx context.Context, id string, h model.HeartbeatRequest, now time.Time) (model.Heartbeat, error) {
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return model.Heartbeat{}, e
	}
	defer tx.Rollback()
	_, e = tx.ExecContext(ctx, "INSERT INTO heartbeats(node_id,load,extra,reported_at) VALUES(?,?,?,?)", id, h.Load, enc(h.Extra), now.Format(time.RFC3339Nano))
	if e != nil {
		return model.Heartbeat{}, e
	}
	_, e = tx.ExecContext(ctx, "UPDATE nodes SET last_heartbeat_at=? WHERE id=?", now.Format(time.RFC3339Nano), id)
	if e != nil {
		return model.Heartbeat{}, e
	}
	if e = tx.Commit(); e != nil {
		return model.Heartbeat{}, e
	}
	return model.Heartbeat{NodeID: id, Load: h.Load, Extra: h.Extra, ReportedAt: now}, e
}
func (s *DB) SetStatus(ctx context.Context, id string, to model.Status, reason, trigger, detail string, now time.Time) (bool, error) {
	n, e := s.GetNode(ctx, id)
	if e != nil {
		return false, e
	}
	if n.Status == to {
		return false, nil
	}
	res, e := s.db.ExecContext(ctx, "UPDATE nodes SET status=?,lost_at=CASE WHEN ?=? THEN ? WHEN ?=? THEN lost_at ELSE NULL END WHERE id=? AND status=?", to, to, model.Lost, now.Format(time.RFC3339Nano), to, model.Recovering, id, n.Status)
	if e != nil {
		return false, e
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		return false, nil
	}
	_, e = s.db.ExecContext(ctx, "INSERT INTO state_transitions(node_id,from_status,to_status,reason,trigger,detail,created_at) VALUES(?,?,?,?,?,?,?)", id, n.Status, to, reason, trigger, detail, now.Format(time.RFC3339Nano))
	return aff == 1, e
}
func (s *DB) StartRecovery(ctx context.Context, id string, now time.Time) (bool, error) {
	r, e := s.db.ExecContext(ctx, "UPDATE nodes SET status=?,recover_attempts=recover_attempts+1 WHERE id=? AND status=?", model.Recovering, id, model.Lost)
	if e != nil {
		return false, e
	}
	a, _ := r.RowsAffected()
	if a == 1 {
		_, e = s.db.ExecContext(ctx, "INSERT INTO state_transitions(node_id,from_status,to_status,reason,trigger,detail,created_at) VALUES(?,?,?,?,?,?,?)", id, model.Lost, model.Recovering, "recovery_start", "scanner", "", now.Format(time.RFC3339Nano))
	}
	return a == 1, e
}
func (s *DB) ListTransitions(ctx context.Context, id string, limit int) ([]model.Transition, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, e := s.db.QueryContext(ctx, "SELECT id,node_id,from_status,to_status,reason,trigger,detail,created_at FROM state_transitions WHERE node_id=? ORDER BY id DESC LIMIT ?", id, limit)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var out []model.Transition
	for rows.Next() {
		var x model.Transition
		var ts string
		if e := rows.Scan(&x.ID, &x.NodeID, &x.From, &x.To, &x.Reason, &x.Trigger, &x.Detail, &ts); e != nil {
			return nil, e
		}
		x.CreatedAt, _ = time.Parse(time.RFC3339Nano, ts)
		out = append(out, x)
	}
	return out, rows.Err()
}
func (s *DB) ListHeartbeats(ctx context.Context, id string, limit int) ([]model.Heartbeat, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, e := s.db.QueryContext(ctx, "SELECT id,node_id,load,extra,reported_at FROM heartbeats WHERE node_id=? ORDER BY id DESC LIMIT ?", id, limit)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var out []model.Heartbeat
	for rows.Next() {
		var x model.Heartbeat
		var extra, ts string
		if e := rows.Scan(&x.ID, &x.NodeID, &x.Load, &extra, &ts); e != nil {
			return nil, e
		}
		dec(extra, &x.Extra)
		x.ReportedAt, _ = time.Parse(time.RFC3339Nano, ts)
		out = append(out, x)
	}
	return out, rows.Err()
}
func (s *DB) Stats(ctx context.Context) (map[model.Status]int, error) {
	rows, e := s.db.QueryContext(ctx, "SELECT status,count(*) FROM nodes GROUP BY status")
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	m := map[model.Status]int{}
	for rows.Next() {
		var st model.Status
		var c int
		if e := rows.Scan(&st, &c); e != nil {
			return nil, e
		}
		m[st] = c
	}
	return m, rows.Err()
}
func (s *DB) ScanCandidates(ctx context.Context, now time.Time, d time.Duration) ([]model.Node, error) {
	rows, e := s.db.QueryContext(ctx, "SELECT "+nodeCols+" FROM nodes WHERE status IN (?,?,?) AND (last_heartbeat_at IS NULL OR last_heartbeat_at<?)", model.Registered, model.Online, model.Recovering, now.Add(-d).Format(time.RFC3339Nano))
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var out []model.Node
	for rows.Next() {
		n, e := scanNode(rows)
		if e != nil {
			return nil, e
		}
		out = append(out, n)
	}
	return out, rows.Err()
}
func (s *DB) Cleanup(ctx context.Context, before time.Time) error {
	_, e := s.db.ExecContext(ctx, "DELETE FROM heartbeats WHERE reported_at<?; DELETE FROM state_transitions WHERE created_at<?", before.Format(time.RFC3339Nano), before.Format(time.RFC3339Nano))
	return e
}

var _ store.Store = (*DB)(nil)
var _ = fmt.Sprintf
