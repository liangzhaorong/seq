package seq

import (
	"reflect"
	"testing"
)

func TestReduce(t *testing.T) {
	add := func(a, b int) int { return a + b }
	t.Run("multiple", func(t *testing.T) {
		v, ok := Of(1, 2, 3, 4).Reduce(add)
		if !ok || v != 10 {
			t.Fatalf("want 10,true; got %v,%v", v, ok)
		}
	})
	t.Run("single", func(t *testing.T) {
		v, ok := Of(7).Reduce(add)
		if !ok || v != 7 {
			t.Fatalf("want 7,true; got %v,%v", v, ok)
		}
	})
	t.Run("empty ok=false", func(t *testing.T) {
		v, ok := Empty[int]().Reduce(add)
		if ok || v != 0 {
			t.Fatalf("want 0,false; got %v,%v", v, ok)
		}
	})
	t.Run("nil ok=false", func(t *testing.T) {
		var s Seq[int]
		v, ok := s.Reduce(add)
		if ok || v != 0 {
			t.Fatalf("want 0,false; got %v,%v", v, ok)
		}
	})
}

func TestFold(t *testing.T) {
	t.Run("sum with seed", func(t *testing.T) {
		got := Of(1, 2, 3, 4).Fold(10, func(acc, n int) int { return acc + n })
		if got != 20 {
			t.Fatalf("want 20, got %v", got)
		}
	})
	t.Run("empty returns seed", func(t *testing.T) {
		if got := Empty[int]().Fold(99, func(acc, n int) int { return acc + n }); got != 99 {
			t.Fatalf("want 99, got %v", got)
		}
	})
	t.Run("type change to string", func(t *testing.T) {
		// int → string：累加器类型 U 可与 T 不同（1→'a', 2→'b', 3→'c'）
		got := Of(1, 2, 3).Fold("", func(acc string, n int) string {
			return acc + string(rune('a'+n-1))
		})
		if got != "abc" {
			t.Fatalf("want abc, got %q", got)
		}
	})
	t.Run("nil returns seed", func(t *testing.T) {
		var s Seq[int]
		if got := s.Fold(42, func(a, n int) int { return a + n }); got != 42 {
			t.Fatalf("want 42, got %v", got)
		}
	})
}

func TestReduceRight(t *testing.T) {
	sub := func(a, b int) int { return a - b }
	t.Run("right order", func(t *testing.T) {
		// [10,3,2]: ReduceRight = 10-(3-2)=9；而 Reduce = (10-3)-2=5
		v, ok := Of(10, 3, 2).ReduceRight(sub)
		if !ok || v != 9 {
			t.Fatalf("want 9,true; got %v,%v", v, ok)
		}
	})
	t.Run("single", func(t *testing.T) {
		v, ok := Of(7).ReduceRight(sub)
		if !ok || v != 7 {
			t.Fatalf("want 7,true; got %v,%v", v, ok)
		}
	})
	t.Run("empty ok=false", func(t *testing.T) {
		v, ok := Empty[int]().ReduceRight(sub)
		if ok || v != 0 {
			t.Fatalf("want 0,false; got %v,%v", v, ok)
		}
	})
	t.Run("string concat", func(t *testing.T) {
		// 拼接可结合：ReduceRight 与 Reduce 同结果，验证基本正确性
		v, ok := Of("a", "b", "c").ReduceRight(func(a, b string) string { return a + b })
		if !ok || v != "abc" {
			t.Fatalf("want abc,true; got %q,%v", v, ok)
		}
	})
	t.Run("nil ok=false", func(t *testing.T) {
		var s Seq[int]
		v, ok := s.ReduceRight(func(a, b int) int { return a + b })
		if ok || v != 0 {
			t.Fatalf("want 0,false; got %v,%v", v, ok)
		}
	})
}

func TestCount(t *testing.T) {
	t.Run("multiple", func(t *testing.T) {
		if got := Of(1, 2, 3, 4, 5).Count(); got != 5 {
			t.Fatalf("want 5, got %v", got)
		}
	})
	t.Run("empty", func(t *testing.T) {
		if got := Empty[int]().Count(); got != 0 {
			t.Fatalf("want 0, got %v", got)
		}
	})
	t.Run("nil", func(t *testing.T) {
		var s Seq[int]
		if got := s.Count(); got != 0 {
			t.Fatalf("want 0, got %v", got)
		}
	})
}

func TestForEach(t *testing.T) {
	t.Run("collects all", func(t *testing.T) {
		var got []int
		Of(1, 2, 3).ForEach(func(n int) { got = append(got, n) })
		want := []int{1, 2, 3}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
	t.Run("empty no calls", func(t *testing.T) {
		calls := 0
		Empty[int]().ForEach(func(int) { calls++ })
		if calls != 0 {
			t.Fatalf("want 0 calls, got %d", calls)
		}
	})
	t.Run("nil no calls", func(t *testing.T) {
		calls := 0
		var s Seq[int]
		s.ForEach(func(int) { calls++ })
		if calls != 0 {
			t.Fatalf("want 0 calls, got %d", calls)
		}
	})
}

func TestForEachWhile(t *testing.T) {
	t.Run("stops when f false", func(t *testing.T) {
		var got []int
		Of(1, 2, 3, 4, 5).ForEachWhile(func(n int) bool {
			got = append(got, n)
			return n < 3
		})
		// 处理 1(继续)、2(继续)、3(返回 false 停止，但 3 已加入)
		want := []int{1, 2, 3}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
	t.Run("consumes all when always true", func(t *testing.T) {
		calls := 0
		Of(1, 2, 3).ForEachWhile(func(int) bool { calls++; return true })
		if calls != 3 {
			t.Fatalf("want 3 calls, got %d", calls)
		}
	})
	t.Run("empty", func(t *testing.T) {
		calls := 0
		Empty[int]().ForEachWhile(func(int) bool { calls++; return true })
		if calls != 0 {
			t.Fatalf("want 0 calls, got %d", calls)
		}
	})
	t.Run("nil no calls", func(t *testing.T) {
		calls := 0
		var s Seq[int]
		s.ForEachWhile(func(int) bool { calls++; return true })
		if calls != 0 {
			t.Fatalf("want 0 calls, got %d", calls)
		}
	})
}

func TestMinByMaxBy(t *testing.T) {
	less := func(a, b int) bool { return a < b }
	t.Run("MinBy multiple", func(t *testing.T) {
		v, ok := Of(3, 1, 4, 1, 5).MinBy(less)
		if !ok || v != 1 {
			t.Fatalf("want 1,true; got %v,%v", v, ok)
		}
	})
	t.Run("MaxBy multiple", func(t *testing.T) {
		v, ok := Of(3, 1, 4, 1, 5).MaxBy(less)
		if !ok || v != 5 {
			t.Fatalf("want 5,true; got %v,%v", v, ok)
		}
	})
	t.Run("single", func(t *testing.T) {
		v, ok := Of(42).MaxBy(less)
		if !ok || v != 42 {
			t.Fatalf("want 42,true; got %v,%v", v, ok)
		}
	})
	t.Run("empty ok=false", func(t *testing.T) {
		if _, ok := Empty[int]().MinBy(less); ok {
			t.Fatal("MinBy: want false")
		}
		if _, ok := Empty[int]().MaxBy(less); ok {
			t.Fatal("MaxBy: want false")
		}
	})
	t.Run("nil ok=false", func(t *testing.T) {
		var s Seq[int]
		if _, ok := s.MinBy(less); ok {
			t.Fatal("MinBy: want false")
		}
		if _, ok := s.MaxBy(less); ok {
			t.Fatal("MaxBy: want false")
		}
	})
}

func TestSum(t *testing.T) {
	t.Run("ints", func(t *testing.T) {
		if got := Sum(Of(1, 2, 3, 4)); got != 10 {
			t.Fatalf("want 10, got %v", got)
		}
	})
	t.Run("empty is zero", func(t *testing.T) {
		if got := Sum(Empty[int]()); got != 0 {
			t.Fatalf("want 0, got %v", got)
		}
	})
	t.Run("nil is zero", func(t *testing.T) {
		if got := Sum[int](nil); got != 0 {
			t.Fatalf("want 0, got %v", got)
		}
	})
	t.Run("floats accumulate", func(t *testing.T) {
		if got := Sum(Of(1.5, 2.5, 3.0)); got != 7.0 {
			t.Fatalf("want 7.0, got %v", got)
		}
	})
}

func TestProduct(t *testing.T) {
	t.Run("ints", func(t *testing.T) {
		if got := Product(Of(1, 2, 3, 4)); got != 24 {
			t.Fatalf("want 24, got %v", got)
		}
	})
	t.Run("empty is one", func(t *testing.T) {
		if got := Product(Empty[int]()); got != 1 {
			t.Fatalf("want 1, got %v", got)
		}
	})
	t.Run("nil is one", func(t *testing.T) {
		if got := Product[int](nil); got != 1 {
			t.Fatalf("want 1, got %v", got)
		}
	})
	t.Run("floats", func(t *testing.T) {
		if got := Product(Of(1.5, 2.0, 2.5)); got != 7.5 {
			t.Fatalf("want 7.5, got %v", got)
		}
	})
}

func TestMinMax(t *testing.T) {
	t.Run("Min int", func(t *testing.T) {
		v, ok := Min(Of(3, 1, 4, 1, 5))
		if !ok || v != 1 {
			t.Fatalf("want 1,true; got %v,%v", v, ok)
		}
	})
	t.Run("Max int", func(t *testing.T) {
		v, ok := Max(Of(3, 1, 4, 1, 5))
		if !ok || v != 5 {
			t.Fatalf("want 5,true; got %v,%v", v, ok)
		}
	})
	t.Run("Min string", func(t *testing.T) {
		v, ok := Min(Of("banana", "apple", "cherry"))
		if !ok || v != "apple" {
			t.Fatalf("want apple,true; got %v,%v", v, ok)
		}
	})
	t.Run("empty ok=false", func(t *testing.T) {
		if _, ok := Min(Empty[int]()); ok {
			t.Fatal("Min: want false")
		}
		if _, ok := Max(Empty[int]()); ok {
			t.Fatal("Max: want false")
		}
	})
	t.Run("nil ok=false", func(t *testing.T) {
		if _, ok := Min[int](nil); ok {
			t.Fatal("Min: want false")
		}
		if _, ok := Max[int](nil); ok {
			t.Fatal("Max: want false")
		}
	})
	t.Run("negatives", func(t *testing.T) {
		v, ok := Min(Of(-5, 0, 3, -2))
		if !ok || v != -5 {
			t.Fatalf("Min negatives: want -5, got %v,%v", v, ok)
		}
		v, ok = Max(Of(-5, 0, 3, -2))
		if !ok || v != 3 {
			t.Fatalf("Max negatives: want 3, got %v,%v", v, ok)
		}
	})
}

func TestMean(t *testing.T) {
	t.Run("ints", func(t *testing.T) {
		if got := Mean(Of(1, 2, 3, 4)); got != 2.5 {
			t.Fatalf("want 2.5, got %v", got)
		}
	})
	t.Run("single", func(t *testing.T) {
		if got := Mean(Of(42)); got != 42.0 {
			t.Fatalf("want 42.0, got %v", got)
		}
	})
	t.Run("floats accumulate", func(t *testing.T) {
		if got := Mean(Of(1.5, 2.5, 3.0)); got != 7.0/3 {
			t.Fatalf("want 7/3, got %v", got)
		}
	})
	t.Run("empty is zero", func(t *testing.T) {
		if got := Mean(Empty[int]()); got != 0 {
			t.Fatalf("want 0, got %v", got)
		}
	})
	t.Run("nil is zero", func(t *testing.T) {
		if got := Mean[int](nil); got != 0 {
			t.Fatalf("want 0, got %v", got)
		}
	})
	t.Run("negatives", func(t *testing.T) {
		if got := Mean(Of(-1, -2, -3)); got != -2.0 {
			t.Fatalf("want -2.0, got %v", got)
		}
	})
}
