package version

import (
	"reflect"
	"strings"

	"github.com/elliotchance/pie/v2"
)

var Properties = []string{"settings", "advanced", "styles"}

type Condition struct {
	X          any           `json:"x,omitempty"`
	Y          any           `json:"y,omitempty"`
	Comparison CompareResult `json:"type,omitempty"`
	And        *Condition    `json:"and,omitempty"`
	Or         *Condition    `json:"or,omitempty"`
}

// CompareResult is an enum representing the result of the comparison.
type CompareResult string

const (
	LT  CompareResult = "LT"  // <
	LTE               = "LTE" // <=
	GT                = "GT"  // >
	GTE               = "GTE" // >=
	EQ                = "EQ"  // ==
	NEQ               = "NEQ" // !=
)

func EvaluateCondition(data map[string]interface{}, condition *Condition) bool {
	if condition == nil {
		return true
	}

	result := CompareInterface(getObjectValue(data, condition.X), getObjectValue(data, condition.Y), condition.Comparison)

	if condition.And != nil {
		if !result {
			return result
		}
		subResult := EvaluateCondition(data, condition.And)
		result = result && subResult
	}

	if condition.Or != nil {
		if result {
			return result
		}
		subResult := EvaluateCondition(data, condition.Or)
		result = result || subResult
	}

	return result
}

func getObjectValue(object map[string]interface{}, field any) interface{} {
	fieldStr, ok := field.(string)
	if !ok {
		return field
	}
	fields := splitFieldPath(fieldStr)
	if pie.Contains(Properties, fields[0]) {
		current := object
		for i := range fields {
			value, ok := current[fields[i]]
			if !ok {
				return nil
			}
			current, ok = value.(map[string]interface{})
			if !ok {
				return value
			}
		}
		return current
	}
	return field
}

func splitFieldPath(fieldPath string) []string {
	return strings.Split(fieldPath, ".")
}

// CompareInterface performs a comparison between two interface{} values with the specified option.
func CompareInterface(a, b interface{}, option CompareResult) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	v1 := reflect.ValueOf(a)
	v2 := reflect.ValueOf(b)
	if v1.Type() != v2.Type() {
		return false
	}

	switch option {
	case LT:
		return compare(v1, v2) < 0
	case LTE:
		return compare(v1, v2) <= 0
	case GT:
		return compare(v1, v2) > 0
	case GTE:
		return compare(v1, v2) >= 0
	case EQ:
		return compare(v1, v2) == 0
	case NEQ:
		return compare(v1, v2) != 0
	}
	return false
}

// Perform comparison of comparable data types here.
// This is where you can check the data type and make the appropriate comparisons.
func compare(x, y reflect.Value) int {
	switch x.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return int(x.Int() - y.Int())
	case reflect.Float32, reflect.Float64:
		return compareFloat(x.Float(), y.Float())
	case reflect.Bool:
		return compareBool(x.Bool(), y.Bool())
	case reflect.String:
		return strings.Compare(x.String(), y.String())
	}
	return 0
}

func compareBool(x, y bool) int {
	if x == y {
		return 0
	} else if x {
		return 1
	}
	return -1
}

func compareFloat(x, y float64) int {
	if x < y {
		return -1
	} else if x > y {
		return 1
	}
	return 0
}
