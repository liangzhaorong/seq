package seq

// Of returns a [Seq[T]] that yields the given values, in order. It is the
// simplest constructor for building a stream from literals.
func Of[T any](vs ...T) Seq[T] {
	return func(yield func(T) bool) {
		for _, v := range vs {
			if !yield(v) {
				return
			}
		}
	}
}

// FromSlice returns a [Seq[T]] backed directly by the given slice. It does not
// copy the slice; callers must not mutate the slice while the stream is being
// consumed. The stream yields the slice elements in order.
func FromSlice[T any](s []T) Seq[T] {
	return func(yield func(T) bool) {
		for _, v := range s {
			if !yield(v) {
				return
			}
		}
	}
}

// Empty returns a [Seq[T]] that yields no elements.
func Empty[T any]() Seq[T] {
	return func(_ func(T) bool) {}
}

// Range returns a [Seq[int]] over the half-open numeric interval [start, stop).
// Step gives the increment (which may be negative to count downward); a zero
// step panics. When step is positive the sequence stops once i >= stop; when
// negative it stops once i <= stop. An empty interval yields nothing.
//
// Range operates on fixed-width int: bounds within step of math.MinInt/math.MaxInt
// can overflow during iteration (inherent to int arithmetic), so callers should
// keep start/stop well inside the int range.
func Range(start, stop, step int) Seq[int] {
	if step == 0 {
		panic("seq: Range step must be non-zero")
	}
	return func(yield func(int) bool) {
		if step > 0 {
			for i := start; i < stop; i += step {
				if !yield(i) {
					return
				}
			}
			return
		}
		for i := start; i > stop; i += step {
			if !yield(i) {
				return
			}
		}
	}
}

// Repeat returns a [Seq[T]] that yields v exactly n times. A negative n is
// treated as zero, yielding an empty stream.
func Repeat[T any](v T, n int) Seq[T] {
	n = max(0, n)
	return func(yield func(T) bool) {
		for range n {
			if !yield(v) {
				return
			}
		}
	}
}

// Generate returns a [Seq[T]] of length n, where the element at each index i is
// produced by f(i). A negative n is treated as zero, yielding an empty stream.
// f is invoked lazily as the stream is consumed, once per index, and only up to
// the point where a downstream consumer stops iterating.
func Generate[T any](f func(index int) T, n int) Seq[T] {
	n = max(0, n)
	return func(yield func(T) bool) {
		for i := range n {
			if !yield(f(i)) {
				return
			}
		}
	}
}

// Cycle returns a [Seq[T]] that repeats the elements of s. When times is
// non-negative the elements of s are emitted times times in succession; when
// times is negative the stream is infinite and must be combined with a
// short-circuiting terminator such as Take, Find or Any — collecting an
// infinite stream blocks forever or exhausts memory.
//
// A nil or empty source yields nothing. In the infinite case (times < 0) an
// empty source terminates after the first empty pass rather than spinning, so
// the short-circuit guarantee still holds.
func Cycle[T any](s Seq[T], times int) Seq[T] {
	return func(yield func(T) bool) {
		if s == nil {
			return
		}
		if times < 0 {
			for {
				yieldedAny := false
				for v := range s {
					yieldedAny = true
					if !yield(v) {
						return
					}
				}
				if !yieldedAny {
					return
				}
			}
		}
		for range times {
			for v := range s {
				if !yield(v) {
					return
				}
			}
		}
	}
}

// FromMap returns a [Seq2[K, V]] over the key/value pairs of m. The iteration
// order is non-deterministic, matching Go's map iteration; callers that need a
// stable order should sort the resulting stream.
func FromMap[K comparable, V any](m map[K]V) Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for k, v := range m {
			if !yield(k, v) {
				return
			}
		}
	}
}
