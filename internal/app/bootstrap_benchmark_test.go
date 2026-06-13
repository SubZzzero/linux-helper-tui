package app_test

import (
	"testing"

	"linux-helper/internal/app"
)

// BenchmarkBootstrap measures startup for the embedded catalog-first application.
func BenchmarkBootstrap(b *testing.B) {
	home := b.TempDir()
	b.Setenv("HOME", home)
	b.ReportAllocs()

	for range b.N {
		model, closeLog, err := app.Bootstrap()
		if err != nil {
			b.Fatalf("bootstrap app: %v", err)
		}
		if view := model.View(); view == "" {
			b.Fatal("expected non-empty startup view")
		}
		if err := closeLog(); err != nil {
			b.Fatalf("close benchmark log: %v", err)
		}
	}
}
