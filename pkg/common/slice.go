package common

import "regexp"

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
