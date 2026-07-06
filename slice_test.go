package seq

import (
	"reflect"
	"testing"
)

func TestTake(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		if got := collect(Empty[int]().Take(3)); got != nil {
			t.Fatalf("want nil, got %v", got)
		}
	})
	t.Run("n=0", func(t *testing.T) {
		if got := collect(Of(1, 2, 3).Take(0)); got != nil {
			t.Fatalf("want nil, got %v", got)
		}
	})
	t.Run("negative n", func(t *testing.T) {
		if got := collect(Of(1, 2, 3).Take(-2)); got != nil {
			t.Fatalf("want nil, got %v", got)
		}
	})
	t.Run("n<len", func(t *testing.T) {
		got := collect(Of(1, 2, 3, 4, 5).Take(3))
		want := []int{1, 2, 3}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
	t.Run("n=len", func(t *testing.T) {
		got := collect(Of(1, 2, 3).Take(3))
		want := []int{1, 2, 3}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
	t.Run("n>len", func(t *testing.T) {
		got := collect(Of(1, 2, 3).Take(10))
		want := []int{1, 2, 3}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
}

func TestTakeLazy(t *testing.T) {
	calls := 0
	s := Generate(func(i int) int { calls++; return i }, 100).Take(5)
	if calls != 0 {
		t.Fatalf("lazy: want 0 calls before iteration, got %d", calls)
	}
	collect(s)
	if calls != 5 {
		t.Fatalf("after collect: want 5 calls, got %d", calls)
	}
}

func TestTakeEarlyTermination(t *testing.T) {
	calls := 0
	s := Generate(func(i int) int { calls++; return i }, 100).Take(20)
	got := takeN(s, 3)
	want := []int{0, 1, 2}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("want %v, got %v", want, got)
	}
	if calls != 3 {
		t.Fatalf("want 3 upstream calls after early break, got %d", calls)
	}
}

func TestTakeWhile(t *testing.T) {
	lt4 := func(n int) bool { return n < 4 }
	t.Run("partial", func(t *testing.T) {
		// 取开头连续 <4 的：1,2,3；遇到 4 停止（不再看后面的 1）
		got := collect(Of(1, 2, 3, 4, 5, 1).TakeWhile(lt4))
		want := []int{1, 2, 3}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
	t.Run("all match", func(t *testing.T) {
		got := collect(Of(1, 2, 3).TakeWhile(func(int) bool { return true }))
		want := []int{1, 2, 3}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
	t.Run("first false", func(t *testing.T) {
		if got := collect(Of(5, 1, 2).TakeWhile(lt4)); got != nil {
			t.Fatalf("want nil, got %v", got)
		}
	})
	t.Run("empty", func(t *testing.T) {
		if got := collect(Empty[int]().TakeWhile(lt4)); got != nil {
			t.Fatalf("want nil, got %v", got)
		}
	})
}

func TestTakeRight(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		if got := collect(Empty[int]().TakeRight(3)); got != nil {
			t.Fatalf("want nil, got %v", got)
		}
	})
	t.Run("n=0", func(t *testing.T) {
		if got := collect(Of(1, 2, 3).TakeRight(0)); got != nil {
			t.Fatalf("want nil, got %v", got)
		}
	})
	t.Run("negative n", func(t *testing.T) {
		if got := collect(Of(1, 2, 3).TakeRight(-2)); got != nil {
			t.Fatalf("want nil, got %v", got)
		}
	})
	t.Run("n<len", func(t *testing.T) {
		got := collect(Of(1, 2, 3, 4, 5).TakeRight(3))
		want := []int{3, 4, 5}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
	t.Run("n=len", func(t *testing.T) {
		got := collect(Of(1, 2, 3).TakeRight(3))
		want := []int{1, 2, 3}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
	t.Run("n>len", func(t *testing.T) {
		got := collect(Of(1, 2, 3).TakeRight(10))
		want := []int{1, 2, 3}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
	t.Run("n=1", func(t *testing.T) {
		got := collect(Of(1, 2, 3).TakeRight(1))
		want := []int{3}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
	t.Run("len is multiple of n", func(t *testing.T) {
		// 流长 6 恰为 n=3 的倍数：环形缓冲回绕到 start=0 边界
		got := collect(Of(1, 2, 3, 4, 5, 6).TakeRight(3))
		want := []int{4, 5, 6}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
}

func TestDrop(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		if got := collect(Empty[int]().Drop(3)); got != nil {
			t.Fatalf("want nil, got %v", got)
		}
	})
	t.Run("n=0 keeps all", func(t *testing.T) {
		got := collect(Of(1, 2, 3).Drop(0))
		want := []int{1, 2, 3}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
	t.Run("negative n keeps all", func(t *testing.T) {
		got := collect(Of(1, 2, 3).Drop(-2))
		want := []int{1, 2, 3}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
	t.Run("n<len", func(t *testing.T) {
		got := collect(Of(1, 2, 3, 4, 5).Drop(2))
		want := []int{3, 4, 5}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
	t.Run("n=len", func(t *testing.T) {
		if got := collect(Of(1, 2, 3).Drop(3)); got != nil {
			t.Fatalf("want nil, got %v", got)
		}
	})
	t.Run("n>len", func(t *testing.T) {
		if got := collect(Of(1, 2, 3).Drop(10)); got != nil {
			t.Fatalf("want nil, got %v", got)
		}
	})
}

func TestDropWhile(t *testing.T) {
	lt4 := func(n int) bool { return n < 4 }
	t.Run("partial", func(t *testing.T) {
		// 丢开头连续 <4 的（1,2,3）；从 4 起全留，含后面再次出现的 1
		got := collect(Of(1, 2, 3, 4, 5, 1).DropWhile(lt4))
		want := []int{4, 5, 1}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
	t.Run("all match drops all", func(t *testing.T) {
		if got := collect(Of(1, 2, 3).DropWhile(func(int) bool { return true })); got != nil {
			t.Fatalf("want nil, got %v", got)
		}
	})
	t.Run("first false keeps all", func(t *testing.T) {
		got := collect(Of(5, 1, 2).DropWhile(lt4))
		want := []int{5, 1, 2}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
	t.Run("empty", func(t *testing.T) {
		if got := collect(Empty[int]().DropWhile(lt4)); got != nil {
			t.Fatalf("want nil, got %v", got)
		}
	})
}

func TestDropRight(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		if got := collect(Empty[int]().DropRight(3)); got != nil {
			t.Fatalf("want nil, got %v", got)
		}
	})
	t.Run("n=0 keeps all", func(t *testing.T) {
		got := collect(Of(1, 2, 3).DropRight(0))
		want := []int{1, 2, 3}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
	t.Run("negative n keeps all", func(t *testing.T) {
		got := collect(Of(1, 2, 3).DropRight(-2))
		want := []int{1, 2, 3}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
	t.Run("n<len", func(t *testing.T) {
		got := collect(Of(1, 2, 3, 4, 5).DropRight(2))
		want := []int{1, 2, 3}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
	t.Run("n=len", func(t *testing.T) {
		if got := collect(Of(1, 2, 3).DropRight(3)); got != nil {
			t.Fatalf("want nil, got %v", got)
		}
	})
	t.Run("n>len", func(t *testing.T) {
		if got := collect(Of(1, 2, 3).DropRight(10)); got != nil {
			t.Fatalf("want nil, got %v", got)
		}
	})
	t.Run("n=1", func(t *testing.T) {
		got := collect(Of(1, 2, 3).DropRight(1))
		want := []int{1, 2}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
	t.Run("len is multiple of n", func(t *testing.T) {
		// 流长 6 恰为 n=2 的倍数：环形缓冲回绕边界
		got := collect(Of(1, 2, 3, 4, 5, 6).DropRight(2))
		want := []int{1, 2, 3, 4}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
}

func TestSlice(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		got := collect(Of(1, 2, 3, 4, 5).Slice(1, 4))
		want := []int{2, 3, 4}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
	t.Run("negative start clamps to 0", func(t *testing.T) {
		got := collect(Of(1, 2, 3).Slice(-1, 2))
		want := []int{1, 2}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
	t.Run("negative end clamps to 0 empty", func(t *testing.T) {
		if got := collect(Of(1, 2, 3).Slice(1, -1)); got != nil {
			t.Fatalf("want nil, got %v", got)
		}
	})
	t.Run("start==end empty", func(t *testing.T) {
		if got := collect(Of(1, 2, 3).Slice(3, 3)); got != nil {
			t.Fatalf("want nil, got %v", got)
		}
	})
	t.Run("start>end empty", func(t *testing.T) {
		if got := collect(Of(1, 2, 3).Slice(4, 2)); got != nil {
			t.Fatalf("want nil, got %v", got)
		}
	})
	t.Run("end beyond length", func(t *testing.T) {
		got := collect(Of(1, 2, 3).Slice(1, 100))
		want := []int{2, 3}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
	t.Run("start beyond length", func(t *testing.T) {
		if got := collect(Of(1, 2, 3).Slice(10, 20)); got != nil {
			t.Fatalf("want nil, got %v", got)
		}
	})
	t.Run("empty", func(t *testing.T) {
		if got := collect(Empty[int]().Slice(0, 3)); got != nil {
			t.Fatalf("want nil, got %v", got)
		}
	})
}

func TestSliceNilReceiver(t *testing.T) {
	var s Seq[int]
	if got := collect(s.Slice(0, 3)); got != nil {
		t.Errorf("Slice nil: want nil, got %v", got)
	}
}

// TestTakeWhileLazy 锁定惰性 + 停止契约：构造期 pred 调用 0 次；pred 首次为假
// 后立即停止消费上游。
func TestTakeWhileLazy(t *testing.T) {
	calls := 0
	pred := func(n int) bool { calls++; return n < 3 }
	s := Generate(func(i int) int { return i }, 100).TakeWhile(pred)
	if calls != 0 {
		t.Fatalf("lazy: want 0 pred calls before iteration, got %d", calls)
	}
	got := collect(s)
	want := []int{0, 1, 2}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("want %v, got %v", want, got)
	}
	// pred 对 0,1,2 为真（产出），对 3 为假（停止）→ 共 4 次
	if calls != 4 {
		t.Fatalf("after collect: want 4 pred calls, got %d", calls)
	}
}

// TestDropWhileLazy 锁定惰性 + 一次性停止契约：dropping 转 false 后不再调用 pred。
func TestDropWhileLazy(t *testing.T) {
	calls := 0
	pred := func(n int) bool { calls++; return n < 3 }
	s := Generate(func(i int) int { return i }, 100).DropWhile(pred)
	if calls != 0 {
		t.Fatalf("lazy: want 0 pred calls before iteration, got %d", calls)
	}
	got := collect(s)
	// drop 0,1,2（pred 真）；pred(3) 假 → 转 dropping=false，3..99 全留（共 97 个）
	if len(got) != 97 || got[0] != 3 || got[len(got)-1] != 99 {
		t.Fatalf("want 97 elems [3..99], got len=%d first=%v last=%v", len(got), got[0], got[len(got)-1])
	}
	// pred 仅在 dropping 阶段调用：0,1,2,3 = 4 次，之后不再调用
	if calls != 4 {
		t.Fatalf("after collect: want 4 pred calls, got %d", calls)
	}
}
