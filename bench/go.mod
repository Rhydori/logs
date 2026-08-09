module github.com/rhydori/logs/bench

go 1.25.3

require (
	github.com/rhydori/logs v0.0.0-00010101000000-000000000000
	github.com/rs/zerolog v1.35.1
	go.uber.org/zap v1.28.0
)

require (
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
)

replace github.com/rhydori/logs => ..
