package seq

// Filter 返回一个 [Seq[T]]，仅保留 pred 为真的元素。
//
// Filter 是惰性的：在终止操作驱动迭代前不会调用 pred；当下游 yield 返回
// false 时，迭代立即向上游传播停止。nil 接收器产出空流（SPEC §5.5）。
func (s Seq[T]) Filter(pred func(T) bool) Seq[T] {
	return func(yield func(T) bool) {
		if s == nil {
			return
		}
		for v := range s {
			if pred(v) {
				if !yield(v) {
					return
				}
			}
		}
	}
}

// FilterNot 返回一个 [Seq[T]]，仅保留 pred 为假的元素。它是 [Seq.Filter] 的
// 补集，用于按否定条件保留元素（例如「不等于某值」「不属于某集合」）。
//
// FilterNot 是惰性的；nil 接收器产出空流（SPEC §5.5）。
func (s Seq[T]) FilterNot(pred func(T) bool) Seq[T] {
	return func(yield func(T) bool) {
		if s == nil {
			return
		}
		for v := range s {
			if !pred(v) {
				if !yield(v) {
					return
				}
			}
		}
	}
}

// DedupeBy 返回一个 [Seq[T]]，丢弃与前一个产出元素 key 相同的元素：仅比较
// 相邻元素，因此非相邻的相同值会被保留。这是「去相邻重复」，区别于全流去重
// （后者需缓冲整流，见 SPEC §4.4 的 Uniq）。
//
// comparable 约束施加在方法自己的类型参数 K 上，而非接收器的 T 上，因此可
// 作为方法存在（SPEC §4.2、设计决策 D4）。DedupeBy 是惰性的：只需记住前一
// 个 key，无需缓冲；nil 接收器产出空流（SPEC §5.5）。
func (s Seq[T]) DedupeBy[K comparable](key func(T) K) Seq[T] {
	return func(yield func(T) bool) {
		if s == nil {
			return
		}
		first := true
		var prev K
		for v := range s {
			k := key(v)
			if first || k != prev {
				if !yield(v) {
					return
				}
				prev = k
				first = false
			}
		}
	}
}

// Dedupe 返回一个 [Seq[T]]，丢弃与前一个产出元素相等的元素：仅比较相邻元素，
// 非相邻的相同值会被保留。
//
// 因方法无法对其接收器的类型参数 T 施加 comparable 约束（设计决策 D4 的硬性
// 语言限制），Dedupe 以顶层泛型函数提供。它在功能上等价于 [Seq.DedupeBy] 配
// 恒等键，但直接比较元素本身，避免恒等闭包在热路径上的间接调用开销（SPEC §8.2
// 闭包内联）。Dedupe 是惰性的；传入 nil 产出空流（SPEC §5.5）。
func Dedupe[T comparable](s Seq[T]) Seq[T] {
	return func(yield func(T) bool) {
		if s == nil {
			return
		}
		first := true
		var prev T
		for v := range s {
			if first || v != prev {
				if !yield(v) {
					return
				}
				prev = v
				first = false
			}
		}
	}
}

// Compact 返回一个 [Seq[T]]，去除零值元素（等于 T 的零值，即 v == *new(T)）。
// 它对齐 lo.Compact 的语义（设计决策 D6）；去相邻重复请用 [Dedupe]。
//
// 因需 T comparable，Compact 以顶层泛型函数提供（设计决策 D4）。Compact 是
// 惰性的，无需缓冲；传入 nil 产出空流（SPEC §5.5）。
func Compact[T comparable](s Seq[T]) Seq[T] {
	return func(yield func(T) bool) {
		if s == nil {
			return
		}
		var zero T
		for v := range s {
			if v != zero {
				if !yield(v) {
					return
				}
			}
		}
	}
}
