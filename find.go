package seq

// FindBy 返回首个满足 pred 的元素；若不存在（或接收器为 nil）返回 def。命中后
// 立即停止迭代（短路）。FindBy 不约束 T，故为方法（设计决策 D4）。
func (s Seq[T]) FindBy(pred func(T) bool, def T) T {
	if s == nil {
		return def
	}
	for v := range s {
		if pred(v) {
			return v
		}
	}
	return def
}

// Any 报告上游是否存在满足 pred 的元素。一旦找到首个满足的元素即返回 true
// （短路）；上游耗尽仍未命中则返回 false。空流或 nil 接收器返回 false（SPEC §5.6）。
func (s Seq[T]) Any(pred func(T) bool) bool {
	if s == nil {
		return false
	}
	for v := range s {
		if pred(v) {
			return true
		}
	}
	return false
}

// All 报告上游是否所有元素都满足 pred。一旦遇到首个不满足的元素即返回 false
// （短路）；上游耗尽全部满足则返回 true。空流或 nil 接收器返回 true（SPEC §5.6：
// 「空真」语义）。
func (s Seq[T]) All(pred func(T) bool) bool {
	if s == nil {
		return true
	}
	for v := range s {
		if !pred(v) {
			return false
		}
	}
	return true
}

// ContainsBy 报告是否存在元素满足 pred，语义等价于 [Seq.Any]（命中即短路）。
// 空流或 nil 接收器返回 false。
func (s Seq[T]) ContainsBy(pred func(T) bool) bool {
	return s.Any(pred)
}

// None 报告是否所有元素都不满足 pred，语义等价于 !Any（遇到首个满足的即返回
// false，短路）。空流或 nil 接收器返回 true。
func (s Seq[T]) None(pred func(T) bool) bool {
	return !s.Any(pred)
}

// Contains 报告上游是否包含值 v。一旦遇到首个等于 v 的元素即返回 true（短路）；
// 上游耗尽仍未命中则返回 false。因需 T comparable，Contains 以顶层泛型函数提供
// （设计决策 D4）。nil 输入返回 false。
func Contains[T comparable](s Seq[T], v T) bool {
	if s == nil {
		return false
	}
	for x := range s {
		if x == v {
			return true
		}
	}
	return false
}

// IndexOf 返回首个等于 v 的元素的索引（0 基）；未找到（或输入为 nil）返回 -1。
// 命中即停止迭代（短路）。因需 T comparable，IndexOf 以顶层泛型函数提供（D4）。
func IndexOf[T comparable](s Seq[T], v T) int {
	if s == nil {
		return -1
	}
	idx := 0
	for x := range s {
		if x == v {
			return idx
		}
		idx++
	}
	return -1
}

// Find 返回首个等于 v 的元素；若不存在（或输入为 nil）返回 def。命中即停止迭代
// （短路）。因需 T comparable，Find 以顶层泛型函数提供（D4）；它是以值查找的
// [Seq.FindBy] 的对应物（后者按谓词查找，不约束 T）。
func Find[T comparable](s Seq[T], v T, def T) T {
	if s == nil {
		return def
	}
	for x := range s {
		if x == v {
			return x
		}
	}
	return def
}
