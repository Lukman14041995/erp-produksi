package pdf

import (
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

func groupThousands(s string) string {
	n := len(s)
	if n <= 3 {
		return s
	}
	var b strings.Builder
	rem := n % 3
	if rem > 0 {
		b.WriteString(s[:rem])
		if n > rem {
			b.WriteByte('.')
		}
	}
	for i := rem; i < n; i += 3 {
		b.WriteString(s[i : i+3])
		if i+3 < n {
			b.WriteByte('.')
		}
	}
	return b.String()
}

// money formats a decimal as Indonesian Rupiah, e.g. "Rp 1.234.567".
func money(v decimal.Decimal) string {
	neg := v.IsNegative()
	intPart := v.Abs().Round(0).StringFixed(0)
	grouped := groupThousands(intPart)
	if neg {
		return "-Rp " + grouped
	}
	return "Rp " + grouped
}

// qty formats a decimal quantity, trimming trailing zeros (e.g. "12.5", "40").
func qty(v decimal.Decimal) string {
	s := v.Round(2).StringFixed(2)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	if s == "" || s == "-" {
		return "0"
	}
	return s
}

func dateStr(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Format("02 Jan 2006")
}

func dateStrPtr(t *time.Time) string {
	if t == nil {
		return "-"
	}
	return dateStr(*t)
}

func orDash(s string) string {
	if strings.TrimSpace(s) == "" {
		return "-"
	}
	return s
}
