package bt

import "fmt"

// Key is a typed blackboard key. The type parameter documents the value type
// expected at that key; it is not enforced by the store itself.
//
//	const Health bt.Key[int] = "health"
//	bt.SetKey(env, Health, 100)
//	h, ok := bt.GetKey(env, Health)
//
// String keys remain fully supported via Env.Set/Get and GetAs. Key helpers
// are an additive, compile-time documentation layer on the same blackboard.
type Key[T any] string

// String returns the underlying blackboard key string.
func (k Key[T]) String() string {
	return string(k)
}

// SetKey stores value under key on env's blackboard.
func SetKey[T any](env Env, key Key[T], value T) {
	if env == nil {
		return
	}
	env.Set(string(key), value)
}

// GetKey reads key from env's blackboard and type-asserts to T.
// ok is false if the key is missing or the value is not of type T.
func GetKey[T any](env Env, key Key[T]) (T, bool) {
	return GetAs[T](env, string(key))
}

// MustGetKey is like GetKey but panics if the key is missing or has the wrong type.
func MustGetKey[T any](env Env, key Key[T]) T {
	t, ok := GetKey(env, key)
	if !ok {
		panic(fmt.Sprintf("bt: MustGetKey[%T](%q): missing or wrong type", t, key))
	}
	return t
}

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
