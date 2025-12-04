package logs

import (
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	reset  = "\033[0m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	blue   = "\033[34m"
	purple = "\033[35m"
	cyan   = "\033[36m"
)

func log(level, color string, parts ...any) {
	now := time.Now().Format("[15:04:05.000]")
	full := strings.TrimSpace(fmt.Sprintln(parts...))
	fmt.Println(color + now + " [" + level + "] " + full + reset)
}

func Debug(args ...any) { log("DEBUG", cyan, args...) }
func Info(args ...any)  { log("INFO", green, args...) }
func Warn(args ...any)  { log("WARN", yellow, args...) }
func Error(args ...any) { log("ERROR", purple, args...) }
func Fatal(args ...any) { log("FATAL", red, args...); os.Exit(0) }

func logf(level, color, msg string, args ...any) {
	now := time.Now().Format("[15:04:05.000]")
	fmt.Println(color + now + " [" + level + "] " + fmt.Sprintf(msg, args...) + reset)
}

func Debugf(msg string, args ...any) { logf("DEBUG", cyan, msg, args...) }
func Infof(msg string, args ...any)  { logf("INFO", green, msg, args...) }
func Warnf(msg string, args ...any)  { logf("WARN", yellow, msg, args...) }
func Errorf(msg string, args ...any) { logf("ERROR", purple, msg, args...) }
func Fatalf(msg string, args ...any) { logf("FATAL", red, msg, args...); os.Exit(0) }
