package logger

import (
	"fmt"
)

var prefix string = Cyan + "[NiimbotGO] " + Reset
var errorPrefix string = Red + "[NiimbotGO - ERROR] " + Reset
var debugPrefix string = Yellow + "[NiimbotGO - DEBUG] " + Reset

var Reset = "\033[0m"
var Red = "\033[31m"
var Yellow = "\033[33m"
var Cyan = "\033[36m"

var enabledInfo bool = true
var enabledError bool = true
var enabledDebug bool = true

func ConfigureLogger(enableInfo bool, enableError bool, enableDebug bool, enableColor bool) {
	enabledInfo = enableInfo
	enabledError = enableError
	enabledDebug = enableDebug
    if !enableColor {
        Reset = ""
        Red = ""
        Yellow = ""
        Cyan = ""
    }
}

func LogError(args ...interface{}) {
	if !enabledError {
		return
	}
	fmt.Print(errorPrefix)
	fmt.Println(args...)
}

func LogInfo(args ...interface{}) {
	if !enabledInfo {
		return
	}
	fmt.Print(prefix)
	fmt.Println(args...)
}

func LogDebug(args ...interface{}) {
	if !enabledDebug {
		return
	}
	fmt.Print(debugPrefix)
	fmt.Println(args...)
}
