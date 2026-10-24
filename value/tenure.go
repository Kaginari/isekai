package value

import "time"

// TenureDays is the tenure signal's reading: whole days from a body's first birth to now (zero
// for a birth in the future — a clock that moved back never reads negative).
func TenureDays(born, now time.Time) float64 {
	if !now.After(born) {
		return 0
	}
	return float64(int(now.Sub(born).Hours() / 24))
}
