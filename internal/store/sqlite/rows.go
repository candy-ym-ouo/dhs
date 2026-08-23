package sqlite

import (
	"database/sql"
	"dhs/internal/model"
	"time"
)

func nullableTime(v *time.Time) any {
	if v == nil {
		return nil
	}
	return v.UTC().Format(time.RFC3339Nano)
}
func nullableString(v sql.NullString) string {
	if !v.Valid {
		return ""
	}
	return v.String
}
func parseTime(s string) time.Time { t, _ := time.Parse(time.RFC3339Nano, s); return t }
func statusList() []model.Status {
	return []model.Status{model.Registered, model.Online, model.Lost, model.Recovering, model.Offline}
}
func isKnownStatus(s model.Status) bool {
	for _, x := range statusList() {
		if s == x {
			return true
		}
	}
	return false
}
