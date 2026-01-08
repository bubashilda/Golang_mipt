//go:build !solution

package testequal

// AssertEqual checks that expected and actual are equal.
//
// Marks caller function as having failed but continues execution.
//
// Returns true iff arguments are equal.
import (
	"bytes"
)

func AssertEqual(t T, expected, actual interface{}, msgAndArgs ...interface{}) bool {
	t.Helper()

	if equal(expected, actual) {
		return true
	}

	if len(msgAndArgs) > 0 {
		if format, ok := msgAndArgs[0].(string); ok {
			t.Errorf(format, msgAndArgs[1:]...)
		} else {
			t.Errorf("%v != %v", expected, actual)
		}
	} else {
		t.Errorf("%v != %v", expected, actual)
	}
	return false
}

func equal(expected, actual interface{}) bool {
	if expected == nil || actual == nil {
		return expected == actual
	}

	switch exp := expected.(type) {
	case int:
		if act, ok := actual.(int); ok {
			return exp == act
		}
	case int8:
		if act, ok := actual.(int8); ok {
			return exp == act
		}
	case int16:
		if act, ok := actual.(int16); ok {
			return exp == act
		}
	case int32:
		if act, ok := actual.(int32); ok {
			return exp == act
		}
	case int64:
		if act, ok := actual.(int64); ok {
			return exp == act
		}

	case uint:
		if act, ok := actual.(uint); ok {
			return exp == act
		}
	case uint8:
		if act, ok := actual.(uint8); ok {
			return exp == act
		}
	case uint16:
		if act, ok := actual.(uint16); ok {
			return exp == act
		}
	case uint32:
		if act, ok := actual.(uint32); ok {
			return exp == act
		}
	case uint64:
		if act, ok := actual.(uint64); ok {
			return exp == act
		}

	case string:
		if act, ok := actual.(string); ok {
			return exp == act
		}

	case []int:
		if act, ok := actual.([]int); ok {
			if exp == nil {
				return act == nil
			}
			if act == nil {
				return exp == nil
			}
			return sliceEqualInt(exp, act)
		}

	case []byte:
		if act, ok := actual.([]byte); ok {
			if exp == nil {
				return act == nil
			}
			if act == nil {
				return exp == nil
			}
			return bytes.Equal(exp, act)
		}

	case map[string]string:
		if act, ok := actual.(map[string]string); ok {
			if exp == nil {
				return act == nil
			}
			if act == nil {
				return exp == nil
			}
			return mapEqualStringString(exp, act)
		}

	default:
		return false
	}

	return false
}

func sliceEqualInt(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func mapEqualStringString(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

// AssertNotEqual checks that expected and actual are not equal.
//
// Marks caller function as having failed but continues execution.
//
// Returns true iff arguments are not equal.
func AssertNotEqual(t T, expected, actual interface{}, msgAndArgs ...interface{}) bool {
	t.Helper()

	if !equal(expected, actual) {
		return true
	}

	if len(msgAndArgs) > 0 {
		if format, ok := msgAndArgs[0].(string); ok {
			t.Errorf(format, msgAndArgs[1:]...)
		} else {
			t.Errorf("%v != %v", expected, actual)
		}
	} else {
		t.Errorf("%v != %v", expected, actual)
	}

	return false
}

// RequireEqual does the same as AssertEqual but fails caller test immediately.
func RequireEqual(t T, expected, actual interface{}, msgAndArgs ...interface{}) {
	t.Helper()

	if equal(expected, actual) {
		return
	}

	if len(msgAndArgs) > 0 {
		if format, ok := msgAndArgs[0].(string); ok {
			t.Errorf(format, msgAndArgs[1:]...)
		} else {
			t.Errorf("%v != %v", expected, actual)
		}
	} else {
		t.Errorf("%v != %v", expected, actual)
	}

	t.FailNow()
}

// RequireNotEqual does the same as AssertNotEqual but fails caller test immediately.
func RequireNotEqual(t T, expected, actual interface{}, msgAndArgs ...interface{}) {
	t.Helper()

	if !equal(expected, actual) {
		return
	}

	if len(msgAndArgs) > 0 {
		if format, ok := msgAndArgs[0].(string); ok {
			t.Errorf(format, msgAndArgs[1:]...)
		} else {
			t.Errorf("%v != %v", expected, actual)
		}
	} else {
		t.Errorf("%v != %v", expected, actual)
	}

	t.FailNow()
}
