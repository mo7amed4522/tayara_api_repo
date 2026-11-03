package cache

import (
	"fmt"
	"reflect"
)

// CostEstimator estimates memory cost of values for Ristretto's MaxCost.
type CostEstimator interface {
	EstimateCost(v interface{}) int64
}

// DefaultCostEstimator provides memory estimates for common types.
type DefaultCostEstimator struct{}

// EstimateCost returns an approximate memory cost in bytes.
func (d *DefaultCostEstimator) EstimateCost(v interface{}) int64 {
	if v == nil {
		return 8
	}

	// Use reflection for accurate estimation
	rv := reflect.ValueOf(v)
	if !rv.IsValid() {
		return 8
	}

	switch rv.Kind() {
	case reflect.String:
		// Go strings: header (16 bytes) + actual bytes
		return int64(rv.Len()) + 16

	case reflect.Slice:
		// Slice header (24 bytes) + underlying array size
		return int64(rv.Len()) * int64(rv.Type().Elem().Size()) + 24

	case reflect.Array:
		// Array: fixed size
		return int64(rv.Len()) * int64(rv.Type().Elem().Size())

	case reflect.Map:
		// Map header (24 bytes) + approximate entry overhead
		return int64(rv.Len()*8 + 100)

	case reflect.Struct:
		// Sum of struct fields
		total := int64(0)
		for i := 0; i < rv.NumField(); i++ {
			total += d.EstimateCost(rv.Field(i).Interface())
		}
		return total

	case reflect.Ptr:
		// Pointer: 8 bytes (assume points to allocated data)
		return 8

	case reflect.Interface:
		return 8

	case reflect.Chan, reflect.Func, reflect.UnsafePointer:
		return 8

	default:
		// Basic types
		switch rv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return 8
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return 8
		case reflect.Bool:
			return 1
		case reflect.Float32, reflect.Float64:
			return 8
		case reflect.Complex64, reflect.Complex128:
			return 16
		default:
			return 16
		}
	}
}

// CostOfBytes estimates cost based on byte array
func CostOfBytes(data []byte) int64 {
	if data == nil {
		return 8
	}
	return int64(len(data)) + 24
}

// CostOfString estimates cost of string
func CostOfString(s string) int64 {
	return int64(len(s)) + 16
}

// EstimateResponseSize estimates size of typical API responses
func EstimateResponseSize(resp interface{}) int64 {
	// Approximate: 100 bytes per string, 200 per object
	total := float64(0)
	val := reflect.ValueOf(resp)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() == reflect.Struct {
		for i := 0; i < val.NumField(); i++ {
			f := val.Field(i)
			switch f.Kind() {
			case reflect.Float64:
				total += f.Float()
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				total += float64(f.Int())
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				total += float64(f.Uint())
			default:
				total += 16 // placeholder for other types
			}
		}
	}
	return int64(total)
}

// FormatSize formats bytes to human readable
func FormatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
