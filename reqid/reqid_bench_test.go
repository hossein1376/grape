package reqid

import (
	"crypto/rand"
	"testing"
)

func BenchmarkRequestID(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = string(NewRequestID())
	}
}

func BenchmarkRandText(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = rand.Text()
	}
}

func BenchmarkRequestIDParallel(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = string(NewRequestID())
		}
	})
}

func BenchmarkRandTextParallel(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = rand.Text()
		}
	})
}
