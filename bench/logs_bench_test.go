package logs

import (
	"log"
	"log/slog"
	"testing"

	. "github.com/rhydori/logs"
	"github.com/rs/zerolog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type benchmarkWriter struct{}

func (benchmarkWriter) Write(buffer []byte) (int, error) {
	return len(buffer), nil
}

func BenchmarkSimpleMessage(b *testing.B) {
	stdLogger := log.New(benchmarkWriter{}, "", 0)
	slogLogger := slog.New(slog.NewTextHandler(benchmarkWriter{}, &slog.HandlerOptions{}))
	zapLogger := newZapBenchmarkLogger()
	zeroLogger := zerolog.New(benchmarkWriter{})

	benchmarks := []struct {
		name  string
		setup func()
		fn    func()
	}{
		{
			name:  "v2-default",
			setup: configureV2Default,
			fn: func() {
				Info("hello")
			},
		},
		{
			name:  "v2-minimal",
			setup: configureV2Minimal,
			fn: func() {
				Info("hello")
			},
		},
		{
			name: "stdlib-log",
			fn:   func() { stdLogger.Print("hello") },
		},
		{
			name: "slog-text",
			fn:   func() { slogLogger.Info("hello") },
		},
		{
			name: "zap",
			fn:   func() { zapLogger.Info("hello") },
		},
		{
			name: "zerolog",
			fn:   func() { zeroLogger.Info().Msg("hello") },
		},
	}

	for _, benchmark := range benchmarks {
		b.Run(benchmark.name, func(b *testing.B) {
			if benchmark.setup != nil {
				benchmark.setup()
			}
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				benchmark.fn()
			}
		})
	}
}

func BenchmarkFormattedMessage(b *testing.B) {
	stdLogger := log.New(benchmarkWriter{}, "", 0)
	zapLogger := newZapBenchmarkLogger().Sugar()
	zeroLogger := zerolog.New(benchmarkWriter{})

	benchmarks := []struct {
		name string
		fn   func()
	}{
		{
			name: "v2",
			fn:   func() { Info("hello %d", 42) },
		},
		{
			name: "stdlib-log",
			fn:   func() { stdLogger.Printf("hello %d", 42) },
		},
		{
			name: "zap-sugared",
			fn:   func() { zapLogger.Infof("hello %d", 42) },
		},
		{
			name: "zerolog",
			fn:   func() { zeroLogger.Info().Msgf("hello %d", 42) },
		},
	}

	configureV2Minimal()
	for _, benchmark := range benchmarks {
		b.Run(benchmark.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				benchmark.fn()
			}
		})
	}
}

func BenchmarkDisabledDebug(b *testing.B) {
	SetMode(MODE_PRODUCTION)
	SetWrite(0)
	b.ReportAllocs()

	for b.Loop() {
		Debug("hello")
	}
}

func configureV2Default() {
	SetMode(MODE_DEVELOPMENT)
	SetWrite(0)
	SetFileLine(true)
	SetDate(DATE_DAY_MONTH_YEAR)
	SetTimer(TIMER_HOUR | TIMER_MINUTE | TIMER_SECOND)
	SetSecondPrecision(SECPRECISION_SECOND)
}

func configureV2Minimal() {
	SetMode(MODE_DEVELOPMENT)
	SetWrite(0)
	SetFileLine(false)
	SetDate(DATE_DAY_MONTH_YEAR)
	SetTimer(0)
}

func newZapBenchmarkLogger() *zap.Logger {
	config := zap.NewProductionEncoderConfig()
	config.TimeKey = ""
	return zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(config),
		zapcore.AddSync(benchmarkWriter{}),
		zap.InfoLevel,
	))
}
