package seq

import (
	"reflect"
	"testing"
)

func TestFilter(t *testing.T) {
	even := func(n int) bool { return n%2 == 0 }

	t.Run("empty", func(t *testing.T) {
		if got := collect(Empty[int]().Filter(even)); got != nil {
			t.Fatalf("want nil, got %v", got)
		}
	})
	t.Run("all kept", func(t *testing.T) {
		got := collect(Of(2, 4, 6).Filter(even))
		want := []int{2, 4, 6}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
	t.Run("all filtered", func(t *testing.T) {
		if got := collect(Of(1, 3, 5).Filter(even)); got != nil {
			t.Fatalf("want nil, got %v", got)
		}
	})
	t.Run("partial", func(t *testing.T) {
		got := collect(Of(1, 2, 3, 4, 5, 6).Filter(even))
		want := []int{2, 4, 6}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
	t.Run("non-comparable type", func(t *testing.T) {
		// 含切片字段的结构体不可比较；Filter 只需 T any（D4：方法不约束 T），
		// 故必须能处理不可比类型。
		type item struct{ tags []string }
		in := []item{{[]string{"a"}}, {}, {[]string{"b"}}, {}}
		got := collect(FromSlice(in).Filter(func(x item) bool { return len(x.tags) > 0 }))
		want := []item{{[]string{"a"}}, {[]string{"b"}}}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
}

func TestFilterNot(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		if got := collect(Empty[int]().FilterNot(func(n int) bool { return n < 0 })); got != nil {
			t.Fatalf("want nil, got %v", got)
		}
	})
	t.Run("keeps complement", func(t *testing.T) {
		got := collect(Of(1, 2, 3, 4, 5, 6).FilterNot(func(n int) bool { return n%2 == 0 }))
		want := []int{1, 3, 5}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
	t.Run("all kept (pred always false)", func(t *testing.T) {
		// 谓词恒假 → 取反后恒真，全部保留；锁定 !pred 不会回归成 pred。
		got := collect(Of(1, 2, 3).FilterNot(func(_ int) bool { return false }))
		want := []int{1, 2, 3}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
	t.Run("equivalent to negated Filter", func(t *testing.T) {
		in := Of(1, 2, 3, 4, 5)
		p := func(n int) bool { return n > 2 }
		notP := collect(in.FilterNot(p))
		negFilter := collect(in.Filter(func(n int) bool { return !p(n) }))
		if !reflect.DeepEqual(notP, negFilter) {
			t.Fatalf("FilterNot(p) = %v, Filter(!p) = %v", notP, negFilter)
		}
	})
}

func TestDedupeBy(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		if got := collect(Empty[int]().DedupeBy(func(n int) int { return n })); got != nil {
			t.Fatalf("want nil, got %v", got)
		}
	})
	t.Run("adjacent by key", func(t *testing.T) {
		// 按奇偶性去相邻重复：连续同奇偶只留首个。
		// 1(奇),3(奇→丢),2(偶),4(偶→丢),5(奇) → [1,2,5]
		got := collect(Of(1, 3, 2, 4, 5).DedupeBy(func(n int) int { return n % 2 }))
		want := []int{1, 2, 5}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
	t.Run("non-adjacent same key kept", func(t *testing.T) {
		// 奇偶交替：每个键都与前一个不同，全部保留。
		got := collect(Of(1, 2, 3, 4).DedupeBy(func(n int) int { return n % 2 }))
		want := []int{1, 2, 3, 4}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
	t.Run("all same key collapses to one", func(t *testing.T) {
		// 所有元素同键 → 只留首个；锁定 first 标志位在「键恒为零值」时仍产出首元素。
		got := collect(Of(1, 2, 3, 4).DedupeBy(func(_ int) int { return 0 }))
		want := []int{1}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
	t.Run("non-comparable T with comparable key", func(t *testing.T) {
		// T 含切片（不可比），但键 K 是 comparable（字符串），方法可行（D4）。
		type item struct {
			id   string
			tags []string
		}
		in := []item{
			{"a", []string{"x"}},
			{"a", []string{"y"}}, // 相邻同 id → 丢
			{"b", []string{"z"}},
		}
		got := collect(FromSlice(in).DedupeBy(func(x item) string { return x.id }))
		want := []item{{"a", []string{"x"}}, {"b", []string{"z"}}}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
}

func TestDedupe(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		if got := collect(Dedupe(Empty[int]())); got != nil {
			t.Fatalf("want nil, got %v", got)
		}
	})
	t.Run("adjacent only", func(t *testing.T) {
		// 仅去相邻：1,1→1; 2,2,2→2; 3; 1（与前一个 3 不同，保留）。
		got := collect(Dedupe(Of(1, 1, 2, 2, 2, 3, 1)))
		want := []int{1, 2, 3, 1}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
	t.Run("all identical collapses to one", func(t *testing.T) {
		got := collect(Dedupe(Of(7, 7, 7, 7)))
		want := []int{7}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
	t.Run("strings", func(t *testing.T) {
		got := collect(Dedupe(Of("a", "a", "b", "b", "a")))
		want := []string{"a", "b", "a"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
	t.Run("nil input", func(t *testing.T) {
		if got := collect(Dedupe[int](nil)); got != nil {
			t.Fatalf("want nil, got %v", got)
		}
	})
}

func TestCompact(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		if got := collect(Compact(Empty[int]())); got != nil {
			t.Fatalf("want nil, got %v", got)
		}
	})
	t.Run("ints drop zero", func(t *testing.T) {
		got := collect(Compact(Of(0, 1, 0, 2, 0, 3)))
		want := []int{1, 2, 3}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
	t.Run("strings drop empty", func(t *testing.T) {
		got := collect(Compact(Of("", "a", "", "b", "")))
		want := []string{"a", "b"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
	t.Run("bools drop false", func(t *testing.T) {
		// bool 零值为 false；锁定按零值（而非 truthy）判定。
		got := collect(Compact(Of(false, true, false, true, false)))
		want := []bool{true, true}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
	t.Run("all zero yields nil", func(t *testing.T) {
		if got := collect(Compact(Of(0, 0, 0))); got != nil {
			t.Fatalf("want nil, got %v", got)
		}
	})
	t.Run("none zero all kept", func(t *testing.T) {
		got := collect(Compact(Of(1, 2, 3)))
		want := []int{1, 2, 3}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
	t.Run("pointers drop nil", func(t *testing.T) {
		x, y := 5, 6
		got := collect(Compact[*int](Of[*int](nil, &x, nil, &y, nil)))
		want := []*int{&x, &y}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
	t.Run("nil input", func(t *testing.T) {
		if got := collect(Compact[int](nil)); got != nil {
			t.Fatalf("want nil, got %v", got)
		}
	})
}

// TestFilterLazy 锁定惰性不变量（SPEC §5.1）：在终止操作驱动迭代前，谓词
// 调用次数为 0。
func TestFilterLazy(t *testing.T) {
	calls := 0
	s := Of(1, 2, 3, 4, 5).Filter(func(n int) bool { calls++; return n%2 == 0 })
	if calls != 0 {
		t.Fatalf("lazy: want 0 calls before iteration, got %d", calls)
	}
	collect(s)
	if calls != 5 {
		t.Fatalf("after collect: want 5 calls, got %d", calls)
	}
}

// TestFilterEarlyTermination 锁定提前终止不变量（SPEC §5.1, FR-5）：下游 break
// 后谓词立即停止调用。
func TestFilterEarlyTermination(t *testing.T) {
	calls := 0
	s := Of(1, 2, 3, 4, 5, 6, 7, 8).Filter(func(_ int) bool { calls++; return true })
	got := takeN(s, 3)
	want := []int{1, 2, 3}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("want %v, got %v", want, got)
	}
	if calls != 3 {
		t.Fatalf("want 3 predicate calls after early break, got %d", calls)
	}
}

func TestDedupeLazy(t *testing.T) {
	calls := 0
	s := Of(1, 1, 2, 2, 3).DedupeBy(func(n int) int { calls++; return n })
	if calls != 0 {
		t.Fatalf("lazy: want 0 calls before iteration, got %d", calls)
	}
	collect(s)
	if calls != 5 {
		t.Fatalf("after collect: want 5 calls, got %d", calls)
	}
}

// TestDedupeEarlyTermination 断言去相邻重复的提前终止：产出前 3 个唯一键
// (0,1,2) 后立即停止拉取上游。
func TestDedupeEarlyTermination(t *testing.T) {
	srcCalls := 0
	// i/2 的键序列：0,0,1,1,2,2,… → 去相邻后 0,1,2,…
	s := Dedupe(Generate(func(i int) int { srcCalls++; return i / 2 }, 100))
	got := takeN(s, 3)
	want := []int{0, 1, 2}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("want %v, got %v", want, got)
	}
	// 产出 0(拉0)、1(拉1丢,拉2)、2(拉3丢,拉4) → 共拉 0,1,2,3,4 = 5 次。
	if srcCalls != 5 {
		t.Fatalf("want 5 upstream calls after early break, got %d", srcCalls)
	}
}

// TestFilterNilReceiver 锁定 SPEC §5.5：对 nil Seq 调用过滤/去重操作产出空流。
func TestFilterNilReceiver(t *testing.T) {
	var s Seq[int] // nil
	if got := collect(s.Filter(func(int) bool { return true })); got != nil {
		t.Errorf("Filter nil: want nil, got %v", got)
	}
	if got := collect(s.FilterNot(func(int) bool { return false })); got != nil {
		t.Errorf("FilterNot nil: want nil, got %v", got)
	}
	if got := collect(s.DedupeBy(func(n int) int { return n })); got != nil {
		t.Errorf("DedupeBy nil: want nil, got %v", got)
	}
}

func TestFilterChaining(t *testing.T) {
	// Range → Filter(偶数) → FilterNot(去掉>6) 端到端。
	got := collect(
		Range(1, 11, 1). // 1..10
					Filter(func(n int) bool { return n%2 == 0 }). // 2,4,6,8,10
					FilterNot(func(n int) bool { return n > 6 }), // 去掉 8,10 → 2,4,6
	)
	want := []int{2, 4, 6}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("want %v, got %v", want, got)
	}
}
