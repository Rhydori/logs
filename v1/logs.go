package logs

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/rhydori/logs/ansi"
	"golang.org/x/term"
)

var (
	logChan = make(chan *[]byte, 1024)
	bufPool = sync.Pool{
		New: func() any {
			b := make([]byte, 0, 256)
			return &b
		},
	}

	wg   sync.WaitGroup
	once sync.Once

	reset = []byte("\033[0m")

	red    = []byte("\033[31m")
	green  = []byte("\033[32m")
	yellow = []byte("\033[33m")
	blue   = []byte("\033[34m")
	purple = []byte("\033[35m")
	cyan   = []byte("\033[36m")

	lDEBUG = []byte("DEBUG")
	lINFO  = []byte("INFO")
	lWARN  = []byte("WARN")
	lERROR = []byte("ERROR")
	lFATAL = []byte("FATAL")
)

func init() {
	isTerm := term.IsTerminal(int(os.Stdout.Fd()))
	noColorEnv := os.Getenv("NO_COLOR") == "true"
	if !noColorEnv && isTerm {
		ansi.EnableANSI()
	} else {
		reset, red, green, yellow, blue, purple, cyan = nil, nil, nil, nil, nil, nil, nil
	}

	wg.Add(1)
	go writer()
}

func writer() {
	defer wg.Done()
	for bp := range logChan {
		os.Stdout.Write(*bp)
		*bp = (*bp)[:0]

		bufPool.Put(bp)
	}
}

func Shutdown() {
	once.Do(func() { close(logChan) })
	wg.Wait()
}

func log(level, color []byte, msg string) {
	bp := bufPool.Get().(*[]byte)
	buf := *bp

	now := time.Now()
	buf = append(buf, color...)
	buf = append(buf, '[')
	appendTime(&buf, now)
	buf = append(buf, "] ["...)
	buf = append(buf, level...)
	buf = append(buf, "] "...)
	buf = append(buf, msg...)
	buf = append(buf, reset...)
	buf = append(buf, '\n')

	*bp = buf
	select {
	case logChan <- bp:
	default:
		bufPool.Put(bp)
	}
}

func logf(level, color []byte, format string, args ...any) {
	bp := bufPool.Get().(*[]byte)
	buf := *bp

	now := time.Now()
	buf = append(buf, color...)
	buf = append(buf, '[')
	appendTime(&buf, now)
	buf = append(buf, "] ["...)
	buf = append(buf, level...)
	buf = append(buf, "] "...)

	buf = fmt.Appendf(buf, format, args...)

	buf = append(buf, reset...)
	buf = append(buf, '\n')

	*bp = buf
	select {
	case logChan <- bp:
	default:
		bufPool.Put(bp)
	}
}

func append2(b *[]byte, v int) {
	*b = append(*b, byte('0'+v/10), byte('0'+v%10))
}

func append3(b *[]byte, v int) {
	*b = append(*b,
		byte('0'+v/100),
		byte('0'+(v/10)%10),
		byte('0'+v%10),
	)
}

func appendTime(b *[]byte, t time.Time) {
	h, m, s := t.Clock()
	ms := t.Nanosecond() / 1e6
	append2(b, h)
	*b = append(*b, ':')
	append2(b, m)
	*b = append(*b, ':')
	append2(b, s)
	*b = append(*b, '.')
	append3(b, ms)
}

func Debug(msg string) { log(lDEBUG, cyan, msg) }
func Info(msg string)  { log(lINFO, green, msg) }
func Warn(msg string)  { log(lWARN, yellow, msg) }
func Error(msg string) { log(lERROR, purple, msg) }
func Fatal(msg string) { log(lFATAL, red, msg); Shutdown(); os.Exit(1) }

func Debugf(format string, args ...any) { logf(lDEBUG, cyan, format, args...) }
func Infof(format string, args ...any)  { logf(lINFO, green, format, args...) }
func Warnf(format string, args ...any)  { logf(lWARN, yellow, format, args...) }
func Errorf(format string, args ...any) { logf(lERROR, purple, format, args...) }
func Fatalf(format string, args ...any) { logf(lFATAL, red, format, args...); Shutdown(); os.Exit(1) }
