package bt

import "fmt"

// GetAs reads key from env's blackboard and type-asserts to T.
// ok is false if the key is missing or the value is not of type T.
func GetAs[T any](env Env, key string) (T, bool) {
	var zero T
	if env == nil {
		return zero, false
	}
	v, ok := env.Get(key)
	if !ok {
		return zero, false
	}
	t, ok := v.(T)
	if !ok {
		return zero, false
	}
	return t, true
}

// MustGet is like GetAs but panics if the key is missing or has the wrong type.
// Prefer GetAs in production paths; MustGet is for tests and trusted setup.
func MustGet[T any](env Env, key string) T {
	t, ok := GetAs[T](env, key)
	if !ok {
		panic(fmt.Sprintf("bt: MustGet[%T](%q): missing or wrong type", t, key))
	}
	return t
}
