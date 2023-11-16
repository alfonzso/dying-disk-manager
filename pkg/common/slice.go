package common

import (
	"regexp"
	"slices"
	"strings"
)

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
