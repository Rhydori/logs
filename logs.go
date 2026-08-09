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

const levelWidth = 5

const (
	LEVEL_INFO  level = "INFO"
	LEVEL_ERROR level = "ERROR"
	LEVEL_WARN  level = "WARN"
	LEVEL_DEBUG level = "DEBUG"
)

const (
	DATE_DAY_MONTH_YEAR date = iota + 1
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

	MessageLevelColor bool
	FileLine          bool
	Date              date
	DateFormat        dateFormat
	Timer             timer
	SecondFormat      secondPrecision
}

var logger = state{
	Mode:  MODE_DEVELOPMENT,
	Write: WRITE_STDIO,

	FileLine:     true,
	DateFormat:   DATEFORMAT_DAY_MONTH_YEAR,
	Timer:        TIMER_HOUR | TIMER_MINUTE | TIMER_SECOND,
	SecondFormat: SECPRECISION_SECOND,
}

var bufferPool = sync.Pool{
	New: func() any {
		buffer := make([]byte, 0, 512)
		return &buffer
	},
}

func SetMode(mode mode) {
	logger.Mode = mode
}

func SetWrite(flags write) {
	logger.Write = flags
}

func SetMessageLevelColor(enabled bool) {
	logger.MessageLevelColor = enabled
}

func SetFileLine(enabled bool) {
	logger.FileLine = enabled
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

func SetSecondPrecision(format secondPrecision) {
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

	buffer := getBuffer()

	// Level
	appendLevel(buffer, level)
	appendSeparator(buffer)

	// Message
	appendMessage(buffer, level, format, args...)

	// FileLine
	if logger.FileLine {
		appendSeparator(buffer)
		appendCaller(buffer)
	}

	// Date
	if logger.Date != 0 {
		appendSeparator(buffer)
		appendDate(buffer, now)
	}

	// Time
	if logger.Timer != 0 {
		appendSeparator(buffer)
		appendTimer(buffer, now)
	}

	// Reset + New Line
	*buffer = append(*buffer, reset...)
	*buffer = append(*buffer, '\n')

	writeInStd(buffer)

	putBuffer(buffer)
}

func appendLevel(buffer *[]byte, level level) {
	appendLevelColor(buffer, level)

	*buffer = append(*buffer, level...)

	for i := len(level); i < levelWidth; i++ {
		*buffer = append(*buffer, ' ')
	}
}

func appendMessage(buffer *[]byte, level level, format string, args ...any) {
	if logger.MessageLevelColor {
		appendLevelColor(buffer, level)
	}

	if len(args) > 0 {
		*buffer = fmt.Appendf(*buffer, format, args...)
	} else {
		*buffer = append(*buffer, format...)
	}
}

func appendCaller(buffer *[]byte) {
	var pcs [1]uintptr
	if runtime.Callers(4, pcs[:]) == 0 {
		return
	}

	function := runtime.FuncForPC(pcs[0])
	if function == nil {
		return
	}

	*buffer = append(*buffer, gray...)
	file, line := function.FileLine(pcs[0])
	*buffer = append(*buffer, filepath.Base(file)...)
	*buffer = append(*buffer, ':')
	*buffer = strconv.AppendInt(*buffer, int64(line), 10)
}

func appendDate(buffer *[]byte, now time.Time) {
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

	*buffer = append(*buffer, gray...)
	*buffer = now.AppendFormat(*buffer, layout)
}

func appendTimer(buffer *[]byte, now time.Time) {
	*buffer = append(*buffer, gray...)

	hour, minute, second := now.Clock()
	if logger.Timer&TIMER_HOUR != 0 {
		*buffer = strconv.AppendInt(*buffer, int64(hour), 10)
		*buffer = append(*buffer, 'h')
	}

	if logger.Timer&TIMER_MINUTE != 0 {
		if logger.Timer&TIMER_HOUR != 0 {
			*buffer = append(*buffer, ' ')
		}
		*buffer = strconv.AppendInt(*buffer, int64(minute), 10)
		*buffer = append(*buffer, 'm')
	}

	if logger.Timer&TIMER_SECOND != 0 {
		if logger.Timer&(TIMER_HOUR|TIMER_MINUTE) != 0 {
			*buffer = append(*buffer, ' ')
		}
		*buffer = strconv.AppendInt(*buffer, int64(second), 10)

		switch logger.SecondFormat {
		default:
			*buffer = append(*buffer, 's')

		case SECPRECISION_MILLI:
			*buffer = append(*buffer, '.')
			append3(buffer, now.Nanosecond()/1e6)
			*buffer = append(*buffer, "ms"...)

		case SECPRECISION_MICRO:
			*buffer = append(*buffer, '.')
			append6(buffer, now.Nanosecond()/1e3)
			*buffer = append(*buffer, "us"...)
		}
	}
}

func appendSeparator(buffer *[]byte) {
	*buffer = append(*buffer, reset...)
	*buffer = append(*buffer, " -- "...)
}

func appendLevelColor(buffer *[]byte, level level) {
	switch level {
	case LEVEL_INFO:
		*buffer = append(*buffer, green...)
	case LEVEL_ERROR:
		*buffer = append(*buffer, red...)
	case LEVEL_WARN:
		*buffer = append(*buffer, yellow...)
	case LEVEL_DEBUG:
		*buffer = append(*buffer, purple...)
	}
}

func getBuffer() *[]byte {
	buffer := bufferPool.Get().(*[]byte)
	*buffer = (*buffer)[:0]

	return buffer
}

func putBuffer(buffer *[]byte) {
	bufferPool.Put(buffer)
}

func writeInStd(buffer *[]byte) {
	if logger.Write&WRITE_STDIO != 0 {
		os.Stdout.Write(*buffer)
	}
	if logger.Write&WRITE_STDERR != 0 {
		os.Stderr.Write(*buffer)
	}
}

func append3(buffer *[]byte, value int) {
	*buffer = append(
		*buffer,
		byte('0'+value/100),
		byte('0'+(value/10)%10),
		byte('0'+value%10),
	)
}

func append6(buffer *[]byte, value int) {
	*buffer = append(
		*buffer,
		byte('0'+value/100000),
		byte('0'+(value/10000)%10),
		byte('0'+(value/1000)%10),
		byte('0'+(value/100)%10),
		byte('0'+(value/10)%10),
		byte('0'+value%10),
	)
}
