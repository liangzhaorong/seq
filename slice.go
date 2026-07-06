package seq

// Take 返回一个 [Seq[T]]，仅产出上游的前 n 个元素。n<=0 产出空流。
//
// Take 是惰性的：产出第 n 个元素后立即停止上游迭代；下游提前停止时同样立即
// 向上游传播。无需缓冲（SPEC §4.2）。nil 接收器产出空流（SPEC §5.5）。
func (s Seq[T]) Take(n int) Seq[T] {
	return func(yield func(T) bool) {
		if s == nil || n <= 0 {
			return
		}
		count := 0
		for v := range s {
			if !yield(v) {
				return
			}
			count++
			if count == n {
				return
			}
		}
	}
}

// TakeWhile 返回一个 [Seq[T]]，逐个产出元素，直到 pred 首次为假（此后停止，
// 不再消费上游）。若 pred 对所有元素恒真，则产出全部元素。
//
// TakeWhile 是惰性的；nil 接收器产出空流（SPEC §5.5）。
func (s Seq[T]) TakeWhile(pred func(T) bool) Seq[T] {
	return func(yield func(T) bool) {
		if s == nil {
			return
		}
		for v := range s {
			if !pred(v) {
				return
			}
			if !yield(v) {
				return
			}
		}
	}
}

// TakeRight 返回一个 [Seq[T]]，仅产出上游的最后 n 个元素。n<0 视为 0（SPEC §5.5）。
//
// 由于「最后」必须等到流结束才能确定，TakeRight 需缓冲至多 n 个元素（环形
// 缓冲，O(n) 内存，SPEC §4.2/§5.3）。n 过大将因 make([]T, n) 分配失败而 panic，
// 调用方应确保 n 合理。nil 接收器产出空流（SPEC §5.5）。
func (s Seq[T]) TakeRight(n int) Seq[T] {
	return func(yield func(T) bool) {
		if s == nil || n <= 0 {
			return
		}
		ring := make([]T, n)
		filled := 0
		for v := range s {
			ring[filled%n] = v
			filled++
		}
		start, count := 0, filled
		if filled > n {
			start, count = filled%n, n
		}
		for i := range count {
			if !yield(ring[(start+i)%n]) {
				return
			}
		}
	}
}

// Drop 返回一个 [Seq[T]]，丢弃上游的前 n 个元素后产出剩余元素。n<=0 产出全部。
//
// Drop 是惰性的，无需缓冲（SPEC §4.2）。nil 接收器产出空流（SPEC §5.5）。
func (s Seq[T]) Drop(n int) Seq[T] {
	return func(yield func(T) bool) {
		if s == nil {
			return
		}
		count := 0
		for v := range s {
			if count < n {
				count++
				continue
			}
			if !yield(v) {
				return
			}
		}
	}
}

// DropWhile 返回一个 [Seq[T]]，丢弃上游开头连续满足 pred 的元素；一旦遇到 pred
// 为假的元素，该元素及其后所有元素全部产出（不再对后续元素调用 pred）。
//
// DropWhile 是惰性的；nil 接收器产出空流（SPEC §5.5）。
func (s Seq[T]) DropWhile(pred func(T) bool) Seq[T] {
	return func(yield func(T) bool) {
		if s == nil {
			return
		}
		dropping := true
		for v := range s {
			if dropping && pred(v) {
				continue
			}
			dropping = false
			if !yield(v) {
				return
			}
		}
	}
}

// DropRight 返回一个 [Seq[T]]，丢弃上游的最后 n 个元素。n<0 视为 0（SPEC §5.5）。
//
// 由于「最后 n 个」必须等到流结束才能确定，DropRight 需缓冲至多 n 个元素
// （环形缓冲，O(n) 内存，SPEC §4.2/§5.3）。n 过大将因 make([]T, n) 分配失败而
// panic，调用方应确保 n 合理。nil 接收器产出空流（SPEC §5.5）。
func (s Seq[T]) DropRight(n int) Seq[T] {
	return func(yield func(T) bool) {
		if s == nil {
			return
		}
		if n <= 0 {
			for v := range s {
				if !yield(v) {
					return
				}
			}
			return
		}
		ring := make([]T, n)
		filled := 0
		for v := range s {
			if filled >= n {
				// 缓冲已满，ring[filled%n] 是 n 步前写入的最旧元素，产出它
				if !yield(ring[filled%n]) {
					return
				}
			}
			ring[filled%n] = v
			filled++
		}
	}
}

// Slice 返回一个 [Seq[T]]，产出索引落在 [start, end) 区间内的元素（左闭右开，
// 语义同 Go 切片）：start 是首个保留元素的索引，end 是首个丢弃元素的索引。
//
// 边界裁剪（godoc 注明语义）：负的 start 归一化为 0；负的 end 归一化为 0；
// 若 start>=end 产出空流；end 超过流长度则取到末尾。Slice 在功能上等价于
// Drop(start).Take(end-start)，但采用单趟索引扫描实现，避免两层闭包的间接
// 调用。Slice 是惰性的、单趟扫描、无需缓冲。nil 接收器产出空流（SPEC §5.5）。
func (s Seq[T]) Slice(start, end int) Seq[T] {
	if start < 0 {
		start = 0
	}
	if end < 0 {
		end = 0
	}
	return func(yield func(T) bool) {
		if s == nil || start >= end {
			return
		}
		idx := 0
		for v := range s {
			if idx >= end {
				return
			}
			if idx >= start {
				if !yield(v) {
					return
				}
			}
			idx++
		}
	}
}
