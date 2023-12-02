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
