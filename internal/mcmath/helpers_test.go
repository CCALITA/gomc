package mcmath_test

import (
	"testing"

	"github.com/fanxiyao/gomc/internal/mcmath"
)

func TestAbsInt(t *testing.T) {
	tests := []struct {
		name string
		in   int
		want int
	}{
		{"positive", 5, 5},
		{"negative", -7, 7},
		{"zero", 0, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := mcmath.Abs(tc.in)
			if got != tc.want {
				t.Errorf("Abs(%d) = %d, want %d", tc.in, got, tc.want)
			}
		})
	}
}

func TestAbsInt32(t *testing.T) {
	tests := []struct {
		name string
		in   int32
		want int32
	}{
		{"positive", 42, 42},
		{"negative", -42, 42},
		{"zero", 0, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := mcmath.Abs(tc.in)
			if got != tc.want {
				t.Errorf("Abs(%d) = %d, want %d", tc.in, got, tc.want)
			}
		})
	}
}

func TestAbsFloat32(t *testing.T) {
	tests := []struct {
		name string
		in   float32
		want float32
	}{
		{"positive", 3.14, 3.14},
		{"negative", -2.71, 2.71},
		{"zero", 0.0, 0.0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := mcmath.Abs(tc.in)
			if got != tc.want {
				t.Errorf("Abs(%v) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

func TestAbsFloat64(t *testing.T) {
	tests := []struct {
		name string
		in   float64
		want float64
	}{
		{"positive", 1.5, 1.5},
		{"negative", -99.9, 99.9},
		{"zero", 0.0, 0.0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := mcmath.Abs(tc.in)
			if got != tc.want {
				t.Errorf("Abs(%v) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}
