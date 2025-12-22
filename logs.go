package logs

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/sys/windows"
	"golang.org/x/term"
)

type logger struct {
	logChan   chan logEntry
	wg        sync.WaitGroup
	closeOnce sync.Once
}

type logColors struct {
	reset  string
	red    string
	green  string
	yellow string
	blue   string
	purple string
	cyan   string
}

type logEntry struct {
	level string
	color string
	msg   string
}

var rlog = logger{
	logChan: make(chan logEntry, 1024),
}
var colors logColors

func enableANSI() {
	h := windows.Handle(os.Stdout.Fd())

	var mode uint32
	windows.GetConsoleMode(h, &mode)
	mode |= windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING
	windows.SetConsoleMode(h, mode)
}

func init() {
	if term.IsTerminal(int(os.Stdout.Fd())) {
		enableANSI()

		colors = logColors{
			reset:  "\033[0m",
			red:    "\033[31m",
			green:  "\033[32m",
			yellow: "\033[33m",
			blue:   "\033[34m",
			purple: "\033[35m",
			cyan:   "\033[36m",
		}
	}

	rlog.wg.Add(1)
	go func() {
		defer rlog.wg.Done()
		for entry := range rlog.logChan {
			now := time.Now().Format("[15:04:05.000]")
			fmt.Printf("%s%s [%s] %s%s\n", entry.color, now, entry.level, entry.msg, colors.reset)
		}
	}()
}

func trySend(entry logEntry) {
	select {
	case rlog.logChan <- entry:
	default:
		os.Stderr.WriteString("rlog: channel full, dropping log\n")
	}
}

func sendLog(level, color string, parts ...any) {
	full := strings.TrimSpace(fmt.Sprintln(parts...))
	trySend(logEntry{level, color, full})
}

func sendLogf(level, color, msg string, args ...any) {
	full := fmt.Sprintf(msg, args...)
	trySend(logEntry{level, color, full})
}

func Debug(args ...any) { sendLog("DEBUG", colors.cyan, args...) }
func Info(args ...any)  { sendLog("INFO", colors.green, args...) }
func Warn(args ...any)  { sendLog("WARN", colors.yellow, args...) }
func Error(args ...any) { sendLog("ERROR", colors.purple, args...) }
func Fatal(args ...any) {
	sendLog("FATAL", colors.red, args...)
	shutdown()
}

func Debugf(msg string, args ...any) { sendLogf("DEBUG", colors.cyan, msg, args...) }
func Infof(msg string, args ...any)  { sendLogf("INFO", colors.green, msg, args...) }
func Warnf(msg string, args ...any)  { sendLogf("WARN", colors.yellow, msg, args...) }
func Errorf(msg string, args ...any) { sendLogf("ERROR", colors.purple, msg, args...) }
func Fatalf(msg string, args ...any) {
	sendLogf("FATAL", colors.red, msg, args...)
	shutdown()
}

func shutdown() {
	rlog.closeOnce.Do(func() {
		close(rlog.logChan)
		rlog.wg.Wait()
		os.Exit(0)
	})
}
