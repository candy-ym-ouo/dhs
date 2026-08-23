package model

import "encoding/json"

func EncodeLabels(v map[string]string) string { b, _ := json.Marshal(v); return string(b) }
func DecodeLabels(s string) map[string]string {
	var v map[string]string
	_ = json.Unmarshal([]byte(s), &v)
	return v
}
func EncodeTasks(v []string) string       { b, _ := json.Marshal(v); return string(b) }
func DecodeTasks(s string) []string       { var v []string; _ = json.Unmarshal([]byte(s), &v); return v }
func EncodeExtra(v map[string]any) string { b, _ := json.Marshal(v); return string(b) }
func DecodeExtra(s string) map[string]any {
	var v map[string]any
	_ = json.Unmarshal([]byte(s), &v)
	return v
}
func Pretty(v any) string { b, _ := json.MarshalIndent(v, "", "  "); return string(b) }
