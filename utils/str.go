package utils

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func StringToInt(str string) int {
	i, err := strconv.Atoi(str)
	if err != nil {
		return 0
	}
	return i
}

func StringToInt64(str string) int64 {
	i, err := strconv.Atoi(str)
	if err != nil {
		return 0
	}
	return int64(i)
}

// Function to convert []string to []int, skipping invalid strings
func StrToIntSlice(strSlice []string) []int {
	var intSlice []int

	for _, str := range strSlice {
		intValue, err := strconv.Atoi(str)
		if err != nil {
			continue
		}
		intSlice = append(intSlice, intValue)
	}

	return intSlice
}

// RandString automatically generates a 12-digit random string
func RandString(n int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

// FormatAmount formats the amount to two decimal places if necessary
func FormatAmount(amount float64) string {
	// Convert float64 to string with 2 decimal places
	amountStr := fmt.Sprintf("%.2f", amount)

	// Check if the amount has exactly two decimal places
	if strings.Contains(amountStr, ".") {
		// Split the amount into integer and decimal parts
		parts := strings.Split(amountStr, ".")
		if len(parts) == 2 {
			// If decimal part is zero, ensure .00 is shown
			if parts[1] == "00" {
				return parts[0] + ".00"
			}
			// If the decimal part is less than 10 (e.g., 100.2), ensure it has two digits
			if len(parts[1]) == 1 {
				parts[1] = parts[1] + "0"
			}
			return parts[0] + "." + parts[1]
		}
	}

	return amountStr
}

// Check if the float amount is exactly two decimal places.
// This function checks if the amount is exactly two decimal places by converting it to a string and then checking the length of the decimal part.
// parameter:
//
//	amount - The floating-point amount that needs to be checked.
//
// Return value:
//
//	If the amount is exactly two decimal places, return true; otherwise, return false.
func IsValidAmount(amount string) bool {
	// Use regular expressions to match the format: number, followed by a decimal point, then two digits.
	regex := `^\d+\.\d{2}$`
	matched, _ := regexp.MatchString(regex, amount)
	return matched
}

func GetStartTime(startTime string) (t time.Time, err error) {
	//t, err = time.Parse(time.DateOnly, startTime)
	t, err = time.Parse("2006-01-02 15:04", startTime)
	if err != nil {
		return
	}
	t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, time.Local)
	return
}

func GetEndTime(endTime string) (t time.Time, err error) {
	//t, err = time.Parse(time.DateOnly, endTime)
	t, err = time.Parse("2006-01-02 15:04", endTime)
	if err != nil {
		return
	}
	t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 59, 0, time.Local)
	return
}
