package seq

// Map returns a [Seq[U]] that applies f to every element of s. It is the core
// generic-method operation: the method carries its own type parameter U so the
// element type can change along the chain (for example int → string).
//
// Map is lazy: f is not invoked until a downstream terminator drives the
// iteration, and it stops as soon as the downstream consumer stops. A nil
// receiver yields an empty stream (SPEC §5.5).
func (s Seq[T]) Map[U any](f func(T) U) Seq[U] {
	return func(yield func(U) bool) {
		if s == nil {
			return
		}
		for v := range s {
			if !yield(f(v)) {
				return
			}
		}
	}
}

// FilterMap returns a [Seq[U]] that applies f to every element of s and keeps
// only the results for which f returns ok=true. It is a combined map and filter
// in a single pass, useful when the predicate and the projection are naturally
// computed together (for example parsing valid items out of a stream).
//
// FilterMap is lazy; a nil receiver yields an empty stream (SPEC §5.5).
func (s Seq[T]) FilterMap[U any](f func(T) (U, bool)) Seq[U] {
	return func(yield func(U) bool) {
		if s == nil {
			return
		}
		for v := range s {
			u, ok := f(v)
			if ok && !yield(u) {
				return
			}
		}
	}
}

// FlatMap returns a [Seq[U]] that applies f to every element of s and yields
// each element of the sub-stream f returns, in order. It is the monadic bind
// (flatten after map) and is the natural way to expand one element into many.
// If f returns nil for an element, that element contributes nothing (nil is
// treated as an empty sub-stream).
//
// FlatMap is lazy; when a downstream consumer stops, both the current
// sub-stream and the outer stream stop immediately. A nil receiver yields an
// empty stream (SPEC §5.5).
func (s Seq[T]) FlatMap[U any](f func(T) Seq[U]) Seq[U] {
	return func(yield func(U) bool) {
		if s == nil {
			return
		}
		for v := range s {
			sub := f(v)
			if sub == nil {
				continue
			}
			for u := range sub {
				if !yield(u) {
					return
				}
			}
		}
	}
}

// Tap returns a [Seq[T]] that invokes f for every element that flows through,
// without changing the stream. It is intended for side effects such as logging
// or counting; f is called once per consumed element, lazily, and is not called
// for elements the downstream consumer never reaches.
//
// A nil receiver yields an empty stream and never invokes f (SPEC §5.5).
func (s Seq[T]) Tap(f func(T)) Seq[T] {
	return func(yield func(T) bool) {
		if s == nil {
			return
		}
		for v := range s {
			f(v)
			if !yield(v) {
				return
			}
		}
	}
}

// Flatten collapses a stream of streams into a single [Seq[U]] by concatenating
// each sub-stream in order. It is the top-level counterpart of [Seq.FlatMap]
// for the common case where the sub-streams already exist. A nil sub-stream
// contributes no elements.
//
// Flatten is lazy; when a downstream consumer stops, the current sub-stream and
// the outer stream stop immediately. A nil outer stream yields an empty stream
// (SPEC §5.5).
func Flatten[U any](s Seq[Seq[U]]) Seq[U] {
	return func(yield func(U) bool) {
		if s == nil {
			return
		}
		for sub := range s {
			if sub == nil {
				continue
			}
			for u := range sub {
				if !yield(u) {
					return
				}
			}
		}
	}
}
