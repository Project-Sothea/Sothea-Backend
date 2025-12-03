package entities

import "time"

// PtrTo is a generic method for creating a pointer to a primitive type, used when manually creating a models object
func PtrTo[T any](v T) *T {
	return &v
}

// SafeDeref is a generic method used in String(), to safely dereference a pointer that might be nil
// Returns the null version of a type, instead of nil, which cannot be printed
func SafeDeref[T any](p *T) T {
	if p == nil {
		var v T
		return v
	}
	return *p
}

// SafeDerefTime safely dereferences a *time.Time, returning a zero time if nil
func SafeDerefTime(p *time.Time) time.Time {
	if p == nil {
		return time.Time{}
	}
	return *p
}

// SafeDerefBool safely dereferences a *bool, returning "Nil" if nil, otherwise "true" or "false"
func SafeDerefBool(p *bool) string {
	if p == nil {
		return "Nil"
	}
	if *p {
		return "true"
	}
	return "false"
}
