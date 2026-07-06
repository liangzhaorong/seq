package seq

import "cmp"

// Reduce 用 merge 从左到右归约上游：以首个元素为累加器初值，依次与后续元素
// merge。返回归约结果与 true；空流（或 nil 接收器）返回 T 的零值与 false
// （SPEC §5.6）。单元素流返回该元素与 true。
func (s Seq[T]) Reduce(merge func(a, b T) T) (T, bool) {
	var zero T
	if s == nil {
		return zero, false
	}
	first := true
	var acc T
	for v := range s {
		if first {
			acc = v
			first = false
		} else {
			acc = merge(acc, v)
		}
	}
	if first {
		return zero, false
	}
	return acc, true
}

// Fold 以 seed 为累加器初值，用 f 从左到右折叠上游；累加器类型 U 可与 T 不同。
// 空流（或 nil 接收器）返回 seed。Fold 总是返回（无需 ok，seed 即默认值）。
func (s Seq[T]) Fold[U any](seed U, f func(U, T) U) U {
	if s == nil {
		return seed
	}
	acc := seed
	for v := range s {
		acc = f(acc, v)
	}
	return acc
}

// ReduceRight 用 merge 从右到左归约上游：等价于 merge(e0, merge(e1, … merge(e_{n-2}, e_{n-1})…))。
// 返回归约结果与 true；空流（或 nil 接收器）返回 T 的零值与 false。
//
// 由于「从右」需先看到末尾，ReduceRight 必须缓冲整个流（O(n) 内存，SPEC §5.3）。
func (s Seq[T]) ReduceRight(merge func(a, b T) T) (T, bool) {
	var zero T
	if s == nil {
		return zero, false
	}
	var buf []T
	for v := range s {
		buf = append(buf, v)
	}
	if len(buf) == 0 {
		return zero, false
	}
	acc := buf[len(buf)-1]
	for i := len(buf) - 2; i >= 0; i-- {
		acc = merge(buf[i], acc)
	}
	return acc, true
}

// Count 返回上游的元素个数。空流或 nil 接收器返回 0。
func (s Seq[T]) Count() int {
	if s == nil {
		return 0
	}
	n := 0
	for range s {
		n++
	}
	return n
}

// ForEach 对上游每个元素调用 f。不短路：消费全部元素（即使 f 无返回值）。
// nil 接收器不调用 f。
func (s Seq[T]) ForEach(f func(T)) {
	if s == nil {
		return
	}
	for v := range s {
		f(v)
	}
}

// ForEachWhile 对上游元素逐个调用 f，直到 f 首次返回 false（此时立即停止上游，
// 不再处理后续元素）。若 f 对所有元素都返回 true，则消费全部。nil 接收器不调用 f。
func (s Seq[T]) ForEachWhile(f func(T) bool) {
	if s == nil {
		return
	}
	for v := range s {
		if !f(v) {
			return
		}
	}
}

// MinBy 返回按 less 排序最小的元素（less(a,b) 为真表示 a<b）与 true；空流
// （或 nil 接收器）返回 T 的零值与 false（SPEC §5.6）。
func (s Seq[T]) MinBy(less func(a, b T) bool) (T, bool) {
	var zero T
	if s == nil {
		return zero, false
	}
	first := true
	var best T
	for v := range s {
		if first {
			best = v
			first = false
		} else if less(v, best) {
			best = v
		}
	}
	if first {
		return zero, false
	}
	return best, true
}

// MaxBy 返回按 less 排序最大的元素（less(a,b) 为真表示 a<b，故 less(best,v)
// 为真时 v 更大）与 true；空流（或 nil 接收器）返回 T 的零值与 false。
func (s Seq[T]) MaxBy(less func(a, b T) bool) (T, bool) {
	var zero T
	if s == nil {
		return zero, false
	}
	first := true
	var best T
	for v := range s {
		if first {
			best = v
			first = false
		} else if less(best, v) {
			best = v
		}
	}
	if first {
		return zero, false
	}
	return best, true
}

// Sum 返回上游所有数值元素之和。空流或 nil 输入返回 T 的零值（0）。
// T 须满足 [Number] 约束（SPEC §3.3/§5.2）。
func Sum[T Number](s Seq[T]) T {
	var acc T
	if s == nil {
		return acc
	}
	for v := range s {
		acc += v
	}
	return acc
}

// Product 返回上游所有数值元素的乘积，以 1（乘法单位元）为初值。空流或 nil
// 输入返回 1。T 须满足 [Number] 约束（SPEC §3.3/§5.2）。
func Product[T Number](s Seq[T]) T {
	var acc T = 1
	if s == nil {
		return acc
	}
	for v := range s {
		acc *= v
	}
	return acc
}

// Min 返回上游中最小的有序元素与 true；空流（或 nil 输入）返回 T 的零值与
// false（SPEC §5.6）。T 须满足 [Ordered] 约束。
func Min[T Ordered](s Seq[T]) (T, bool) {
	var zero T
	if s == nil {
		return zero, false
	}
	first := true
	var best T
	for v := range s {
		if first {
			best = v
			first = false
		} else if cmp.Compare(v, best) < 0 {
			best = v
		}
	}
	if first {
		return zero, false
	}
	return best, true
}

// Max 返回上游中最大的有序元素与 true；空流（或 nil 输入）返回 T 的零值与
// false（SPEC §5.6）。T 须满足 [Ordered] 约束。
func Max[T Ordered](s Seq[T]) (T, bool) {
	var zero T
	if s == nil {
		return zero, false
	}
	first := true
	var best T
	for v := range s {
		if first {
			best = v
			first = false
		} else if cmp.Compare(v, best) > 0 {
			best = v
		}
	}
	if first {
		return zero, false
	}
	return best, true
}

// Mean 返回上游数值元素的算术平均值（float64）。空流或 nil 输入返回 0
// （SPEC §5.6）。T 须满足 [Number] 约束；元素在累加前转为 float64，因此极大
// 整数可能损失精度（固有的 float64 表示限制）。
func Mean[T Number](s Seq[T]) float64 {
	if s == nil {
		return 0
	}
	sum := 0.0
	n := 0
	for v := range s {
		sum += float64(v)
		n++
	}
	if n == 0 {
		return 0
	}
	return sum / float64(n)
}
