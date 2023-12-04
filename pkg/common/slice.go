package common

import (
	"regexp"
	"slices"
	"strings"
)

type Common struct {
	valueAsStr  string
	valueAsList []string
}

func DeleteEmpty(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}

func IsEquals[V int | string](a, b []V) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func Split(src, expr string) []string {
	return regexp.MustCompile(expr).Split(src, -1)
}

func GrepInList(source []string, pattern string) string {
	idx := slices.IndexFunc(source, func(row string) bool {
		return strings.Contains(row, pattern)
	})
	if idx == -1 {
		return ""
	}
	return source[idx]
}

func Map[T, V any](ts []T, fn func(T, int) V) []V {
	result := make([]V, len(ts))
	for idx, t := range ts {
		result[idx] = fn(t, idx)
	}
	return result
}

// func Filter(vs []string, f func(string) bool) []string {
// 	vsf := make([]string, 0)
// 	for _, v := range vs {
// 		if f(v) {
// 			vsf = append(vsf, v)
// 		}
// 	}
// 	return vsf
// }

func Filter[T any](items []T, fn func(item T) bool) []T {
	filteredItems := []T{}
	for _, value := range items {
			if fn(value) {
					filteredItems = append(filteredItems, value)
			}
	}
	return filteredItems
}

// Nomad test

func Maybe(value string) *Common {
	return &Common{valueAsStr: value}
}

func (c *Common) Split(expr string) *Common {
	c.valueAsList = regexp.MustCompile(expr).Split(c.valueAsStr, -1)
	return c
}

func (c *Common) DeleteEmpty(andMore string) *Common {
	var r []string
	for _, str := range c.valueAsList {
		if str != "" && str != andMore {
			r = append(r, str)
		}
	}
	c.valueAsList = r
	return c
}

func (c *Common) GrepInList(pattern string) *Common {
	idx := slices.IndexFunc(c.valueAsList, func(row string) bool {
		return strings.Contains(row, pattern)
	})
	if idx == -1 {
		c.valueAsStr = ""
	}
	c.valueAsStr = c.valueAsList[idx]
	return c
}

func (c *Common) ToStr() *Common {
	c.valueAsStr = strings.Join(c.valueAsList, " ")
	return c
}

func (c *Common) GetStr() string {
	return c.valueAsStr
}

func (c *Common) GetList() []string {
	return c.valueAsList
}
