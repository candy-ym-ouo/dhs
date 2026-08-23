package store

import "math"

type Page struct {
	Number int `json:"page"`
	Size   int `json:"page_size"`
	Total  int `json:"total"`
	Pages  int `json:"pages"`
}

func NormalizePage(number, size int) Page {
	if number < 1 {
		number = 1
	}
	if size <= 0 {
		size = 20
	}
	if size > 100 {
		size = 100
	}
	return Page{Number: number, Size: size}
}
func (p Page) Offset() int { return (p.Number - 1) * p.Size }
func (p *Page) SetTotal(total int) {
	p.Total = total
	p.Pages = int(math.Ceil(float64(total) / float64(p.Size)))
}
