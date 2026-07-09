package logger

import (
	"fmt"
	"strings"
)

func formatLevel(i interface{}) string {
	level := strings.ToUpper(fmt.Sprintf("%s", i))

	switch level {
	case "TRACE":
		return "\033[37mTRACE\033[0m"
	case "DEBUG":
		return "\033[36mDEBUG\033[0m"
	case "INFO":
		return "\033[32mINFO \033[0m"
	case "WARN":
		return "\033[33mWARN \033[0m"
	case "ERROR":
		return "\033[31mERROR\033[0m"
	default:
		return level
	}
}
