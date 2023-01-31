package sync

import "sync/atomic"

// AtomicBool is an atomic Boolean, which make it safe to be used from multiple go routines.
// If you embed in a struct use a pointer to avoid copies.
type AtomicBool int32

// New creates an AtomicBool with default set to false.
func New() *AtomicBool {
	return new(AtomicBool)
}

// Set bool to true.
func (ab *AtomicBool) Set() {
	atomic.StoreInt32((*int32)(ab), 1)
}

// UnSet bool to false.
func (ab *AtomicBool) UnSet() {
	atomic.StoreInt32((*int32)(ab), 0)
}

// IsSet returns whether the Boolean is true.
func (ab *AtomicBool) IsSet() bool {
	return atomic.LoadInt32((*int32)(ab)) == 1
}
