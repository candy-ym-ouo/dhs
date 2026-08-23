package sqlite

import (
	"database/sql/driver"
	"encoding/json"
)

type JSON map[string]any

func (j JSON) Value() (driver.Value, error) { return json.Marshal(j) }
func (j *JSON) Scan(src any) error {
	if src == nil {
		*j = nil
		return nil
	}
	b, ok := src.([]byte)
	if !ok {
		b = []byte(src.(string))
	}
	return json.Unmarshal(b, j)
}
