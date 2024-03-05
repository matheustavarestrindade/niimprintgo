package logger

import (
	"fmt"
)

var prefix string = "[NiimbotGO] "

var enabledInfo bool = true

func LogInfo(args ...interface{}) {
	if !enabledInfo {
		return
	}
    fmt.Print(prefix)
    fmt.Println(args...)

}
