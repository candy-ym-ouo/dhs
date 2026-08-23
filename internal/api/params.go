package api

import (
	"net/http"
	"strconv"
	"strings"
)

type Params struct {
	Page    int
	Size    int
	Limit   int
	Keyword string
}

func parseParams(r *http.Request) Params {
	q := r.URL.Query()
	p, _ := strconv.Atoi(q.Get("page"))
	s, _ := strconv.Atoi(q.Get("page_size"))
	l, _ := strconv.Atoi(q.Get("limit"))
	if p < 1 {
		p = 1
	}
	if s < 1 {
		s = 20
	}
	if s > 100 {
		s = 100
	}
	if l < 1 {
		l = 50
	}
	if l > 500 {
		l = 500
	}
	return Params{p, s, l, strings.TrimSpace(q.Get("keyword"))}
}
func pathParts(path string) []string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	out := parts[:0]
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
