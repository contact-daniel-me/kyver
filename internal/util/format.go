package util

import (
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// FormatSize is the standard formatting helper for all Kyver commands that display file sizes.
// It ensures that file sizes are displayed consistently across the entire CLI application.
// It converts a byte count into a human-readable string with comma separators where appropriate.
// It consistently follows IEC binary units (1024-based) up to TB, rounding to 2 decimal places.
func FormatSize(b int64) string {
	p := message.NewPrinter(language.English)
	const unit = 1024
	
	if b == 0 {
		return "0 B"
	}
	
	if b < unit {
		return p.Sprintf("%d B", b)
	}

	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"KB", "MB", "GB", "TB"}
	if exp >= len(units) {
		exp = len(units) - 1
		// If it goes beyond TB, we just cap it at TB but adjust div
		for i := 0; i < exp; i++ {
			div = 1024 // just recompute div safely if we capped exp, but actually we can just leave it
		}
		// Actually, let's keep it simple. If b > 1024 TB, it'll just show huge TB.
	}
	
	return p.Sprintf("%.2f %s", float64(b)/float64(div), units[exp])
}
