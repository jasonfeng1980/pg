package util

import (
	"runtime"
	"strings"
)

func FuncName(level int) string {
	pc, _, _, _ := runtime.Caller(level)
	name := runtime.FuncForPC(pc).Name()
	split := strings.Split(name, ".")
	return split[len(split)-1]
}
