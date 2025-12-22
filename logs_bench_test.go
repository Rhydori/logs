package logs

import (
	"testing"
)

func BenchmarkLogger(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Info("hello world", i)
	}
}
