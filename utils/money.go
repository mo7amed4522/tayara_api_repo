package utils

import (
	"strconv"
)

// Minor-unit money helpers. All monetary values are int64 in minor units (AED fils).
// Percentages are int64 basis points, where 10000 == 100%. No floats are ever used for money.

const (
	BasisPointsPerPercent = 10000
)

// MoneyPercentOf returns amount * basisPoints / 10000, rounded half-up.
// e.g. MoneyPercentOf(100_000, 1000) == 10_000  (10% of 1000.00 AED).
func MoneyPercentOf(amount int64, basisPoints int64) int64 {
	if basisPoints <= 0 {
		return 0
	}
	// half-up rounding: add half the divisor before dividing
	return (amount*basisPoints + BasisPointsPerPercent/2) / BasisPointsPerPercent
}

// AgentShare computes the agent's take: gross - strip fee - platform commission.
func AgentShare(gross, stripFee, commission int64) int64 {
	return gross - stripFee - commission
}

// MoneyToDecimal renders minor units as a decimal string without floats, e.g. 123456 -> "1234.56".
func MoneyToDecimal(minor int64) string {
	neg := minor < 0
	if neg {
		minor = -minor
	}
	whole := minor / 100
	frac := minor % 100
	s := strconv.FormatInt(whole, 10)
	if frac < 10 {
		s += ".0"
	} else {
		s += "."
	}
	s += strconv.FormatInt(frac, 10)
	if neg {
		s = "-" + s
	}
	return s
}
