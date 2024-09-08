package util

func Args(params ...interface{}) []interface{} { return params }

func Tern[T bool, U any](isTrue T, ifValue U, elseValue U) U {
	if isTrue {
		return ifValue
	} else {
		return elseValue
	}
}

func Cast(condition bool, trueFun, falseFun func()) {
	if condition {
		if trueFun != nil {
			trueFun()
		}
	} else {
		if falseFun != nil {
			falseFun()
		}
	}
}

func DefaultVal[T any](vars []T) T {
	var zero T
	if len(vars) > 0 {
		return vars[0]
	}
	return zero
}
