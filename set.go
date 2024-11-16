package core

import "sync"

// Set implements a simple set using generics.
type Set[K, H comparable, T any] struct {
	// Hash computes the bucket identifier for the given K type
	Hash func(k K) (H, error)
	// ItemKey computes the K value from the T instance
	ItemKey func(v T) (K, error)
	// ItemMatch confirms the T instance matches the K value
	ItemMatch func(k K, v T) bool

	mu      sync.RWMutex
	buckets map[H]*List[T]
}

func (set *Set[K, H, T]) lazyInit() error {
	if set == nil {
		return ErrNilReceiver
	}

	// RO
	set.mu.RLock()
	ready := set.isInitialized()
	set.mu.RUnlock()
	if ready {
		return nil
	}

	// RW
	set.mu.Lock()
	defer set.mu.Unlock()

	set.unsafeInit()
	return nil
}

func (set *Set[K, H, T]) isInitialized() bool {
	switch {
	case set == nil, set.buckets == nil:
		return false
	default:
		return true
	}
}

func (set *Set[K, H, T]) unsafeInit() {
	if !set.isInitialized() {
		set.unsafeReset()
	}
}

// Reset removes all entires from the [Set].
func (set *Set[K, H, T]) Reset() error {
	if set == nil {
		return ErrNilReceiver
	}

	// RW
	set.mu.Lock()
	defer set.mu.Unlock()

	set.unsafeReset()
	return nil
}

func (set *Set[K, H, T]) unsafeReset() {
	set.buckets = make(map[H]*List[T])
}

// Push adds entries to the set unless it already exist.
// It returns the value with matching key stored in the Set so it
// can be treated as a global reference.
func (set *Set[K, H, T]) Push(value T) (T, error) {
	var zero T

	if err := set.lazyInit(); err != nil {
		return zero, err
	}

	key, err := set.doItemKey(value)
	if err != nil {
		return zero, err
	}

	hash, err := set.doHash(key)
	if err != nil {
		return zero, err
	}

	// RW
	set.mu.Lock()
	defer set.mu.Unlock()

	l, ok := set.buckets[hash]
	if !ok {
		// new
		set.buckets[hash] = set.newList(value)
		return value, nil
	}

	if v, err := set.unsafeGet(key, l); err == nil {
		// found
		return v, ErrExists
	}

	// new
	l.PushBack(value)
	return value, nil
}

func (*Set[K, H, T]) newList(values ...T) *List[T] {
	l := new(List[T])
	for _, v := range values {
		l.PushBack(v)
	}
	return l
}

// Get returns the item matching the key
func (set *Set[K, H, T]) Get(key K) (T, error) {
	var zero T

	if err := set.lazyInit(); err != nil {
		return zero, err
	}

	hash, err := set.doHash(key)
	if err != nil {
		// invalid key
		return zero, err
	}

	// RO
	set.mu.RLock()
	defer set.mu.RUnlock()

	l, ok := set.buckets[hash]
	if !ok {
		return zero, ErrNotExists
	}

	return set.unsafeGet(key, l)
}

func (set *Set[K, H, T]) unsafeGet(key K, l *List[T]) (T, error) {
	var zero T

	match, err := set.newMatchKey(key)
	if err != nil {
		return zero, err
	}

	value, found := l.FirstMatchFn(match)
	if !found {
		return zero, ErrNotExists
	}

	return value, nil
}

func (set *Set[K, H, T]) newMatchKey(key K) (func(T) bool, error) {
	fn1 := set.ItemMatch
	if fn1 == nil {
		return nil, Wrap(ErrNotImplemented, "ItemMatch")
	}

	fn2 := func(v T) bool {
		return fn1(key, v)
	}

	return fn2, nil
}

// Pop removes and return the item matching the given key from the
// Set.
func (set *Set[K, H, T]) Pop(key K) (T, error) {
	var zero T

	if err := set.lazyInit(); err != nil {
		return zero, err
	}

	hash, err := set.doHash(key)
	if err != nil {
		return zero, err
	}

	match, err := set.newMatchKey(key)
	if err != nil {
		return zero, err
	}

	// RW
	set.mu.Lock()
	defer set.mu.Unlock()

	l, ok := set.buckets[hash]
	if !ok {
		return zero, ErrNotExists
	}

	value, ok := l.PopFirstMatchFn(match)
	if !ok {
		return zero, ErrNotExists
	}
	return value, nil
}

func (set *Set[K, H, T]) doHash(key K) (H, error) {
	var zero H

	if fn := set.Hash; fn != nil {
		return fn(key)
	}

	return zero, Wrap(ErrNotImplemented, "Hash")
}

func (set *Set[K, H, T]) doItemKey(value T) (K, error) {
	var zero K

	if fn := set.ItemKey; fn != nil {
		return fn(value)
	}

	return zero, Wrap(ErrNotImplemented, "ItemKey")
}
