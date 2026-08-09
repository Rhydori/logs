package logs

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"time"
)

type (
	mode  uint8
	write uint8

	level string

	date            uint8
	dateFormat      uint8
	timer           uint16
	secondPrecision uint8
)

const (
	MODE_DEVELOPMENT mode = iota
	MODE_PRODUCTION
)

const (
	WRITE_STDIO write = 1 << iota
	WRITE_STDERR
)

const (
	LEVEL_INFO  level = "INFO"
	LEVEL_ERROR level = "ERROR"
	LEVEL_WARN  level = "WARN"
	LEVEL_DEBUG level = "DEBUG"
)

const (
	DATE_DAY_MONTH_YEAR date = iota
	DATE_DAY_MONTH
	DATE_MONTH_YEAR
)

const (
	DATEFORMAT_DAY_MONTH_YEAR dateFormat = iota
	DATEFORMAT_MONTH_DAY_YEAR
	DATEFORMAT_YEAR_MONTH_DAY
)

const (
	TIMER_HOUR timer = 1 << iota
	TIMER_MINUTE
	TIMER_SECOND
)

const (
	SECPRECISION_SECOND secondPrecision = iota
	SECPRECISION_MILLI
	SECPRECISION_MICRO
)

const (
	reset  = "\033[0m"
	gray   = "\033[30m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	purple = "\033[35m"
)

type state struct {
	Mode  mode
	Write write

	FileLine     bool
	Date         date
	DateFormat   dateFormat
	Timer        timer
	SecondFormat secondPrecision
}

var logger = state{
	Mode:  MODE_DEVELOPMENT,
	Write: WRITE_STDIO,

	FileLine:     true,
	Date:         DATE_DAY_MONTH_YEAR,
	DateFormat:   DATEFORMAT_DAY_MONTH_YEAR,
	Timer:        TIMER_HOUR | TIMER_MINUTE | TIMER_SECOND,
	SecondFormat: SECPRECISION_SECOND,
}

var bufferPool = sync.Pool{
	New: func() any {
		buffer := make([]byte, 0, 256)
		return &buffer
	},
}

func SetMode(mode mode) {
	logger.Mode = mode
}

func SetWrite(flags write) {
	logger.Write = flags
}

func SetFileLine(caller bool) {
	logger.FileLine = caller
}

func SetDate(date date) {
	logger.Date = date
}

func SetDateFormat(format dateFormat) {
	logger.DateFormat = format
}

func SetTimer(timer timer) {
	logger.Timer = timer
}

func SetSecondFormat(format secondPrecision) {
	logger.SecondFormat = format
}

//go:noinline
func Info(format string, args ...any) {
	createLog(LEVEL_INFO, format, args...)
}

//go:noinline
func Error(format string, args ...any) {
	createLog(LEVEL_ERROR, format, args...)
}

//go:noinline
func Warn(format string, args ...any) {
	createLog(LEVEL_WARN, format, args...)
}

//go:noinline
func Debug(format string, args ...any) {
	createLog(LEVEL_DEBUG, format, args...)
}

func createLog(level level, format string, args ...any) {
	if level == LEVEL_DEBUG && logger.Mode == MODE_PRODUCTION {
		return
	}

	var now time.Time
	if logger.Date != 0 || logger.Timer != 0 {
		now = time.Now()
	}

	bufferPointer := getBuffer()
	buffer := (*bufferPointer)[:0]

	// Level
	buffer = appendLevel(buffer, level)

	// Message
	buffer = append(buffer, '\'')
	if len(args) > 0 {
		buffer = fmt.Appendf(buffer, format, args...)
	} else {
		buffer = append(buffer, format...)
	}
	buffer = append(buffer, '\'')

	// Separator
	buffer = append(buffer, reset...)
	buffer = append(buffer, " - "...)

	// FileLine
	if logger.FileLine {
		buffer = append(buffer, gray...)
		buffer = appendCaller(buffer)
		buffer = append(buffer, reset...)
		buffer = append(buffer, " - "...)
	}

	// Date
	if logger.Date != 0 {
		buffer = append(buffer, gray...)
		buffer = appendDate(buffer, now)
		buffer = append(buffer, reset...)
		buffer = append(buffer, " - "...)
	}

	// Time
	if logger.Timer != 0 {
		buffer = append(buffer, gray...)
		buffer = appendTimer(buffer, now)
		buffer = append(buffer, reset...)
	}

	// New Line
	buffer = append(buffer, '\n')

	writeInStd(buffer)

	*bufferPointer = buffer
	putBuffer(bufferPointer)
}

func appendCaller(buffer []byte) []byte {
	var pcs [1]uintptr
	if runtime.Callers(4, pcs[:]) == 0 {
		return buffer
	}

	function := runtime.FuncForPC(pcs[0])
	if function == nil {
		return buffer
	}

	file, line := function.FileLine(pcs[0])
	buffer = append(buffer, filepath.Base(file)...)
	buffer = append(buffer, ':')
	return strconv.AppendInt(buffer, int64(line), 10)
}

func appendLevel(buffer []byte, level level) []byte {
	switch level {
	case LEVEL_INFO:
		buffer = append(buffer, green...)
	case LEVEL_ERROR:
		buffer = append(buffer, red...)
	case LEVEL_WARN:
		buffer = append(buffer, yellow...)
	case LEVEL_DEBUG:
		buffer = append(buffer, purple...)
	}

	buffer = append(buffer, level...)
	buffer = append(buffer, ' ')

	return buffer
}

func appendDate(buffer []byte, now time.Time) []byte {
	layout := ""
	switch logger.Date {
	case DATE_DAY_MONTH_YEAR:
		switch logger.DateFormat {
		case DATEFORMAT_DAY_MONTH_YEAR:
			layout = "02/01/2006"
		case DATEFORMAT_MONTH_DAY_YEAR:
			layout = "01/02/2006"
		case DATEFORMAT_YEAR_MONTH_DAY:
			layout = "2006/01/02"
		}

	case DATE_DAY_MONTH:
		if logger.DateFormat == DATEFORMAT_MONTH_DAY_YEAR {
			layout = "01/02"
		} else {
			layout = "02/01"
		}

	case DATE_MONTH_YEAR:
		if logger.DateFormat == DATEFORMAT_YEAR_MONTH_DAY {
			layout = "2006/01"
		} else {
			layout = "01/2006"
		}
	}

	return now.AppendFormat(buffer, layout)
}

func appendTimer(buffer []byte, now time.Time) []byte {
	hour, minute, second := now.Clock()

	if logger.Timer&TIMER_HOUR != 0 {
		buffer = append2(buffer, hour)
		buffer = append(buffer, 'h')
	}

	if logger.Timer&TIMER_MINUTE != 0 {
		buffer = append2(buffer, minute)
		buffer = append(buffer, 'm')
	}

	if logger.Timer&TIMER_SECOND != 0 {
		buffer = append2(buffer, second)
		buffer = append(buffer, 's')

		switch logger.SecondFormat {
		case SECPRECISION_MILLI:
			buffer = append3(buffer, now.Nanosecond()/1e6)
			buffer = append(buffer, "ms"...)

		case SECPRECISION_MICRO:
			buffer = append6(buffer, now.Nanosecond()/1e3)
			buffer = append(buffer, "us"...)
		}
	}

	return buffer
}

func getBuffer() *[]byte {
	buffer := bufferPool.Get().(*[]byte)
	*buffer = (*buffer)[:0]
	return buffer
}

func putBuffer(buffer *[]byte) {
	bufferPool.Put(buffer)
}

func writeInStd(buffer []byte) {
	if logger.Write&WRITE_STDIO != 0 {
		os.Stdout.Write(buffer)
	}
	if logger.Write&WRITE_STDERR != 0 {
		os.Stderr.Write(buffer)
	}
}

func append2(buffer []byte, value int) []byte {
	return append(
		buffer,
		byte('0'+value/10),
		byte('0'+value%10),
	)
}

func append3(buffer []byte, value int) []byte {
	return append(
		buffer,
		byte('0'+value/100),
		byte('0'+(value/10)%10),
		byte('0'+value%10),
	)
}

func append6(buffer []byte, value int) []byte {
	return append(
		buffer,
		byte('0'+value/100000),
		byte('0'+(value/10000)%10),
		byte('0'+(value/1000)%10),
		byte('0'+(value/100)%10),
		byte('0'+(value/10)%10),
		byte('0'+value%10),
	)
}
