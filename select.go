package seq

// First 返回上游的第一个元素与 true；若上游为空或为 nil，返回 T 的零值与
// false。找到首个元素后立即停止迭代（短路）。
func (s Seq[T]) First() (T, bool) {
	var zero T
	if s == nil {
		return zero, false
	}
	for v := range s {
		return v, true
	}
	return zero, false
}

// Last 返回上游的最后一个元素与 true；若上游为空或为 nil，返回 T 的零值与
// false。Last 必须迭代到底才能确定末尾元素（无法短路）。
func (s Seq[T]) Last() (T, bool) {
	var zero T
	if s == nil {
		return zero, false
	}
	found := false
	var last T
	for v := range s {
		last = v
		found = true
	}
	if !found {
		return zero, false
	}
	return last, true
}

// At 返回上游第 i 个（0 基）元素与 true；若 i 越界（i<0 或 i>=长度）或上游为
// nil，返回 T 的零值与 false。找到目标后立即停止迭代（短路）。
func (s Seq[T]) At(i int) (T, bool) {
	var zero T
	if s == nil || i < 0 {
		return zero, false
	}
	idx := 0
	for v := range s {
		if idx == i {
			return v, true
		}
		idx++
	}
	return zero, false
}

// Partition 按谓词 pred 把上游一分为二，返回两条惰性流：第一条保留 pred 为真
// 的元素，第二条保留 pred 为假的元素。
//
// 两条流各自独立迭代上游一次，因此要求上游可被重复消费（[Of]、[FromSlice]、
// [Range] 等构造器天然可重入；channel-backed 或一次性上游不适用）。pred 在每
// 条流中各调用一次。nil 接收器返回两条空流（SPEC §5.5）。
func (s Seq[T]) Partition(pred func(T) bool) (Seq[T], Seq[T]) {
	yes := func(yield func(T) bool) {
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
	no := func(yield func(T) bool) {
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
	return yes, no
}
