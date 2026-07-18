package util

import (
	"testing"
)

func TestFormatSize(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{842, "842 B"},
		{1000, "1,000 B"},
		{1023, "1,023 B"},
		{1024, "1.00 KB"},
		{1048576, "1.00 MB"},
		{1073741824, "1.00 GB"},
		{12840, "12.54 KB"},
		{3323985, "3.17 MB"},
		{1524713472, "1.42 GB"},
		{2199023255552, "2.00 TB"},
	}

	for _, tt := range tests {
		result := FormatSize(tt.bytes)
		if result != tt.expected {
			t.Errorf("FormatSize(%d): expected %q, got %q", tt.bytes, tt.expected, result)
		}
	}
}

func BenchmarkFormatSize(b *testing.B) {
	for i := 0; i < b.N; i++ {
		FormatSize(3323985) // 3.17 MB
	}
}
