package strutils

import (
	"fmt"
	"strings"
)

func Strings[S fmt.Stringer](elems []S) []string {
	out := make([]string, 0, len(elems))
	for i := range elems {
		out = append(out, elems[i].String())
	}
	return out
}

func Join[S fmt.Stringer](elems []S, sep string) string {
	return strings.Join(Strings(elems), sep)
}
