package template

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

//The following func is copied from spf13/hugo and do some refactor.

var timeType = reflect.TypeOf((*time.Time)(nil)).Elem()

// DoArithmetic performs arithmetic operations (+,-,*,/) using reflection to
// determine the type of the two terms.
// This func will auto convert string and uint to int64/float64, then apply  operations,
// return float64, or int64
func DoArithmetic(a, b interface{}, op rune) (interface{}, error) {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	var ai, bi int64
	var af, bf float64

	var err error
	if av.Kind() == reflect.String {
		av, err = stringToNumber(av)
		if err != nil {
			return nil, err
		}
	}
	if bv.Kind() == reflect.String {
		bv, err = stringToNumber(bv)
		if err != nil {
			return nil, err
		}
	}

	switch av.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		ai = av.Int()
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			bi = bv.Int()
		case reflect.Float32, reflect.Float64:
			af = float64(ai) // may overflow
			ai = 0
			bf = bv.Float()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			bi = int64(bv.Uint()) // may overflow
		default:
			return nil, errors.New("Can't apply the operator to the values")
		}
	case reflect.Float32, reflect.Float64:
		af = av.Float()
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			bf = float64(bv.Int()) // may overflow
		case reflect.Float32, reflect.Float64:
			bf = bv.Float()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			bf = float64(bv.Uint()) // may overflow
		default:
			return nil, errors.New("Can't apply the operator to the values")
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		ai = int64(av.Uint()) // may overflow
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			bi = bv.Int()
		case reflect.Float32, reflect.Float64:
			af = float64(ai) // may overflow
			ai = 0
			bf = bv.Float()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			bi = int64(bv.Uint()) // may overflow
		default:
			return nil, errors.New("Can't apply the operator to the values")
		}
	default:
		return nil, errors.New("Can't apply the operator to the values")
	}

	switch op {
	case '+':
		if af != 0 || bf != 0 {
			return af + bf, nil
		} else if ai != 0 || bi != 0 {
			return ai + bi, nil
		}
		return 0, nil
	case '-':
		if af != 0 || bf != 0 {
			return af - bf, nil
		} else if ai != 0 || bi != 0 {
			return ai - bi, nil
		}
		return 0, nil
	case '*':
		if af != 0 || bf != 0 {
			return af * bf, nil
		}
		if ai != 0 || bi != 0 {
			return ai * bi, nil
		}
		return 0, nil
	case '/':
		if bf != 0 {
			return af / bf, nil
		} else if bi != 0 {
			return ai / bi, nil
		}
		return nil, errors.New("Can't divide the value by 0")
	default:
		return nil, errors.New("There is no such an operation")
	}
}

func stringToNumber(value reflect.Value) (reflect.Value, error) {
	var result reflect.Value
	str := value.String()
	if isFloat(str) {
		vf, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("Can't apply the operator to the value [%s] ,err [%s] ", str, err.Error())
		}
		result = reflect.ValueOf(vf)
	} else {
		vi, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("Can't apply the operator to the value [%s] ,err [%s] ", str, err.Error())
		}
		result = reflect.ValueOf(vi)
	}
	return result, nil
}

func isFloat(value string) bool {
	return strings.Index(value, ".") >= 0
}

// eq returns the boolean truth of arg1 == arg2.
func eq(x, y interface{}) bool {
	normalize := func(v interface{}) interface{} {
		vv := reflect.ValueOf(v)
		nv, err := stringToNumber(vv)
		if err == nil {
			vv = nv
		}
		switch vv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return float64(vv.Int()) //may overflow
		case reflect.Float32, reflect.Float64:
			return vv.Float()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return float64(vv.Uint()) //may overflow
		default:
			return v
		}
	}
	x = normalize(x)
	y = normalize(y)
	return reflect.DeepEqual(x, y)
}

// ne returns the boolean truth of arg1 != arg2.
func ne(x, y interface{}) bool {
	return !eq(x, y)
}

// ge returns the boolean truth of arg1 >= arg2.
func ge(a, b interface{}) bool {
	left, right := compareGetFloat(a, b)
	return left >= right
}

// gt returns the boolean truth of arg1 > arg2.
func gt(a, b interface{}) bool {
	left, right := compareGetFloat(a, b)
	return left > right
}

// le returns the boolean truth of arg1 <= arg2.
func le(a, b interface{}) bool {
	left, right := compareGetFloat(a, b)
	return left <= right
}

// lt returns the boolean truth of arg1 < arg2.
func lt(a, b interface{}) bool {
	left, right := compareGetFloat(a, b)
	return left < right
}

func compareGetFloat(a interface{}, b interface{}) (float64, float64) {
	var left, right float64
	var leftStr, rightStr *string
	av := reflect.ValueOf(a)

	switch av.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice:
		left = float64(av.Len())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		left = float64(av.Int())
	case reflect.Float32, reflect.Float64:
		left = av.Float()
	case reflect.String:
		var err error
		left, err = strconv.ParseFloat(av.String(), 64)
		if err != nil {
			str := av.String()
			leftStr = &str
		}
	case reflect.Struct:
		switch av.Type() {
		case timeType:
			left = float64(toTimeUnix(av))
		}
	}

	bv := reflect.ValueOf(b)

	switch bv.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice:
		right = float64(bv.Len())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		right = float64(bv.Int())
	case reflect.Float32, reflect.Float64:
		right = bv.Float()
	case reflect.String:
		var err error
		right, err = strconv.ParseFloat(bv.String(), 64)
		if err != nil {
			str := bv.String()
			rightStr = &str
		}
	case reflect.Struct:
		switch bv.Type() {
		case timeType:
			right = float64(toTimeUnix(bv))
		}
	}

	switch {
	case leftStr == nil || rightStr == nil:
	case *leftStr < *rightStr:
		return 0, 1
	case *leftStr > *rightStr:
		return 1, 0
	default:
		return 0, 0
	}

	return left, right
}

// mod returns a % b.
func mod(a, b interface{}) (int64, error) {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)
	var err error
	if av.Kind() == reflect.String {
		av, err = stringToNumber(av)
		if err != nil {
			return 0, err
		}
	}
	if bv.Kind() == reflect.String {
		bv, err = stringToNumber(bv)
		if err != nil {
			return 0, err
		}
	}

	var ai, bi int64

	switch av.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		ai = av.Int()
	default:
		return 0, errors.New("Modulo operator can't be used with non integer value")
	}

	switch bv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		bi = bv.Int()
	default:
		return 0, errors.New("Modulo operator can't be used with non integer value")
	}

	if bi == 0 {
		return 0, errors.New("The number can't be divided by zero at modulo operation")
	}

	return ai % bi, nil
}

func toTimeUnix(v reflect.Value) int64 {
	if v.Kind() == reflect.Interface {
		return toTimeUnix(v.Elem())
	}
	if v.Type() != timeType {
		panic("coding error: argument must be time.Time type reflect Value")
	}
	return v.MethodByName("Unix").Call([]reflect.Value{})[0].Int()
}
