package handlers

import (
	"html/template"
	"strconv"
)

func parseIntSlice(vals []string) []int {
	out := make([]int, 0, len(vals))
	for _, v := range vals {
		i, err := strconv.Atoi(v)
		if err == nil {
			out = append(out, i)
		}
	}
	return out
}

func isSubset(sub, set []int) bool {
	m := map[int]bool{}
	for _, v := range set {
		m[v] = true
	}
	for _, v := range sub {
		if !m[v] {
			return false
		}
	}
	return true
}

func containsInt(xs []int, v int) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}

func mustParse(files ...string) *template.Template {
	t := template.New("").Funcs(template.FuncMap{
		"derefInt": func(p *int) int {
			if p == nil {
				return 0
			}
			return *p
		},
	})
	return template.Must(t.ParseFiles(files...))
}
