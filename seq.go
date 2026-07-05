// Package seq provides Scala-style lazy, chainable collection operations on top
// of Go's range-over-func iterators ([iter.Seq[T]] / [iter.Seq2[K, V]]).
//
// The package wraps iter.Seq[T] in the named type [Seq[T]] so that methods with
// their own type parameters — a Go 1.27 feature — can be defined on it. This
// enables true chained type transformations such as:
//
//	seq.Of(1, 2, 3).
//	    Filter(func(n int) bool { return n%2 == 1 }).
//	    Map(func(n int) string { return fmt.Sprintf("#%d", n) }).
//	    Take(2).
//	    Collect() // → ["#1", "#3"]
//
// All intermediate operations are lazy: they perform no work until a terminating
// operation (Collect, Reduce, ForEach, ...) drives the iteration. When a
// downstream yield function returns false, iteration stops immediately upstream.
//
// Per design decision D4 (see tasks/spec-lazy-seq.md), operations that constrain
// the element type T (comparable / [Ordered] / [Number]) are provided as
// top-level generic functions rather than methods, because Go 1.27 generic
// methods cannot add constraints to the receiver's type parameter.
package seq

import "iter"

// Seq is the named wrapper around [iter.Seq[T]]. It is defined as a named type
// (not an alias) so that methods with their own type parameters can be declared
// on it. The underlying type is identical to iter.Seq[T], so conversion between
// the two is a zero-cost compile-time reinterpretation (see [Seq.Unbox] and
// [FromSeq]).
type Seq[T any] func(yield func(T) bool)

// Seq2 is the named wrapper around [iter.Seq2[K, V]], the iterator that yields
// key/value pairs. It is the return type of operations such as GroupBy, Zip and
// [FromMap], and supports its own chain of KV operations.
type Seq2[K any, V any] func(yield func(K, V) bool)

// Unbox returns the wrapped value as a standard library [iter.Seq[T]], so that
// Seq[T] can be consumed by packages such as [slices] and [maps]. The
// conversion is zero-cost because Seq[T] and iter.Seq[T] share the same
// underlying type.
func (s Seq[T]) Unbox() iter.Seq[T] { return iter.Seq[T](s) }

// Unbox returns the wrapped value as a standard library [iter.Seq2[K, V]]. It is
// the [Seq2] analogue of [Seq.Unbox], letting Seq2 feed standard library KV
// helpers such as [maps.Collect].
func (s Seq2[K, V]) Unbox() iter.Seq2[K, V] { return iter.Seq2[K, V](s) }

// FromSeq wraps a standard library [iter.Seq[T]] as a [Seq[T]], admitting it
// into the chainable world. The conversion is zero-cost.
func FromSeq[T any](s iter.Seq[T]) Seq[T] { return Seq[T](s) }

// FromSeq2 wraps a standard library [iter.Seq2[K, V]] as a [Seq2[K, V]]. It is
// the [Seq2] analogue of [FromSeq]; the conversion is zero-cost.
func FromSeq2[K any, V any](s iter.Seq2[K, V]) Seq2[K, V] { return Seq2[K, V](s) }
