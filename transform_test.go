package seq

import (
	"reflect"
	"strings"
	"testing"
)

func TestMap(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		got := collect(Empty[int]().Map(func(n int) int { return n + 1 }))
		if got != nil {
			t.Fatalf("want nil, got %v", got)
		}
	})
	t.Run("single", func(t *testing.T) {
		got := collect(Of(5).Map(func(n int) int { return n * n }))
		want := []int{25}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
	t.Run("multiple", func(t *testing.T) {
		got := collect(Of(1, 2, 3, 4).Map(func(n int) int { return n + 10 }))
		want := []int{11, 12, 13, 14}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
	t.Run("type change", func(t *testing.T) {
		got := collect(Of(1, 2, 3).Map(func(n int) string {
			switch n {
			case 1:
				return "a"
			case 2:
				return "b"
			}
			return "c"
		}))
		want := []string{"a", "b", "c"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
}

func TestFilterMap(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		got := collect(Empty[int]().FilterMap(func(n int) (int, bool) { return n, true }))
		if got != nil {
			t.Fatalf("want nil, got %v", got)
		}
	})
	t.Run("keep evens squared", func(t *testing.T) {
		// keep even numbers and square them, drop odds
		got := collect(Of(1, 2, 3, 4, 5, 6).FilterMap(func(n int) (int, bool) {
			if n%2 == 0 {
				return n * n, true
			}
			return 0, false
		}))
		want := []int{4, 16, 36}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
	t.Run("type change", func(t *testing.T) {
		got := collect(Of("1", "x", "2", "y", "3").FilterMap(func(s string) (int, bool) {
			if len(s) == 1 && s[0] >= '0' && s[0] <= '9' {
				return int(s[0] - '0'), true
			}
			return 0, false
		}))
		want := []int{1, 2, 3}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
	t.Run("all filtered", func(t *testing.T) {
		got := collect(Of(1, 3, 5).FilterMap(func(int) (int, bool) { return 0, false }))
		if got != nil {
			t.Fatalf("want nil, got %v", got)
		}
	})
}

func TestFlatMap(t *testing.T) {
	t.Run("empty outer", func(t *testing.T) {
		got := collect(Empty[int]().FlatMap(func(n int) Seq[int] { return Of(n, n) }))
		if got != nil {
			t.Fatalf("want nil, got %v", got)
		}
	})
	t.Run("expand each", func(t *testing.T) {
		// each int n -> [n, n*10]
		got := collect(Of(1, 2, 3).FlatMap(func(n int) Seq[int] { return Of(n, n*10) }))
		want := []int{1, 10, 2, 20, 3, 30}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
	t.Run("empty sub-streams", func(t *testing.T) {
		// every f returns an empty stream
		got := collect(Of(1, 2, 3).FlatMap(func(int) Seq[int] { return Empty[int]() }))
		if got != nil {
			t.Fatalf("want nil, got %v", got)
		}
	})
	t.Run("nested via Range", func(t *testing.T) {
		// each i -> [0, 1, ..., i-1]  (i.e., Range(0, i, 1))
		got := collect(Of(1, 2, 3).FlatMap(func(i int) Seq[int] { return Range(0, i, 1) }))
		want := []int{0, 0, 1, 0, 1, 2}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
	t.Run("type change", func(t *testing.T) {
		got := collect(Of(1, 2).FlatMap(func(n int) Seq[string] {
			return Of("a" + strings.Repeat("x", n))
		}))
		want := []string{"ax", "axx"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
}

func TestFlatten(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		got := collect(Flatten(Empty[Seq[int]]()))
		if got != nil {
			t.Fatalf("want nil, got %v", got)
		}
	})
	t.Run("concat", func(t *testing.T) {
		got := collect(Flatten(Of(Of(1, 2), Of(3), Of(4, 5, 6))))
		want := []int{1, 2, 3, 4, 5, 6}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
	t.Run("with empty sub-stream", func(t *testing.T) {
		got := collect(Flatten(Of(Of(1, 2), Empty[int](), Of(3))))
		want := []int{1, 2, 3}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
}

func TestTap(t *testing.T) {
	t.Run("values unchanged", func(t *testing.T) {
		got := collect(Of(1, 2, 3).Tap(func(int) {}))
		want := []int{1, 2, 3}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
	t.Run("side effect per element", func(t *testing.T) {
		seen := []int{}
		got := collect(Of(1, 2, 3).Tap(func(n int) { seen = append(seen, n) }))
		if !reflect.DeepEqual(got, []int{1, 2, 3}) {
			t.Fatalf("values changed: %v", got)
		}
		if !reflect.DeepEqual(seen, []int{1, 2, 3}) {
			t.Fatalf("tap saw %v, want %v", seen, []int{1, 2, 3})
		}
	})
	t.Run("empty", func(t *testing.T) {
		calls := 0
		collect(Empty[int]().Tap(func(int) { calls++ }))
		if calls != 0 {
			t.Fatalf("want 0 tap calls, got %d", calls)
		}
	})
}

// TestMapLazy asserts the laziness invariant (SPEC §5.1): the mapping function
// is not invoked until the stream is consumed by a terminator.
func TestMapLazy(t *testing.T) {
	calls := 0
	s := Of(1, 2, 3, 4, 5).Map(func(n int) int { calls++; return n * 2 })
	if calls != 0 {
		t.Fatalf("before iteration: want 0 calls, got %d", calls)
	}
	got := collect(s)
	if calls != 5 {
		t.Fatalf("after collect: want 5 calls, got %d", calls)
	}
	want := []int{2, 4, 6, 8, 10}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

// TestMapEarlyTermination asserts the early-termination invariant (SPEC §5.1,
// FR-5): breaking out of the iteration stops the upstream producer and the
// mapping function immediately.
func TestMapEarlyTermination(t *testing.T) {
	srcCalls := 0
	mapCalls := 0
	s := Generate(func(i int) int { srcCalls++; return i }, 100).
		Map(func(n int) int { mapCalls++; return n })
	count := 0
	for range s {
		count++
		if count == 5 {
			break
		}
	}
	if srcCalls != 5 {
		t.Fatalf("upstream: want 5 calls, got %d", srcCalls)
	}
	if mapCalls != 5 {
		t.Fatalf("map fn: want 5 calls, got %d", mapCalls)
	}
}

func TestFlatMapEarlyTermination(t *testing.T) {
	srcCalls := 0
	s := Generate(func(i int) int { srcCalls++; return i }, 100).
		FlatMap(func(i int) Seq[int] { return Of(i, i+100) })
	count := 0
	for range s {
		count++
		if count == 4 {
			break
		}
	}
	// 4 elements come from src 0 (→0,100) and 1 (→1,101); the break fires
	// after the 4th element so src 2 is never pulled.
	if srcCalls != 2 {
		t.Fatalf("upstream: want 2 src calls after early break, got %d", srcCalls)
	}
}

func TestChaining(t *testing.T) {
	// end-to-end: Range → FilterMap → Map → FlatMap → collect (SPEC §9.2 style)
	got := collect(
		Range(1, 11, 1). // 1..10
					FilterMap(func(n int) (int, bool) { return n, n%2 == 0 }). // evens 2,4,6,8,10
					Map(func(n int) int { return n / 2 }).                     // 1,2,3,4,5
					FlatMap(func(n int) Seq[int] { return Of(n, -n) }),        // each ±
	)
	want := []int{1, -1, 2, -2, 3, -3, 4, -4, 5, -5}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

// TestNilReceiverSafe locks SPEC §5.5: calling any operator on a nil Seq
// yields an empty stream and never panics.
func TestNilReceiverSafe(t *testing.T) {
	var s Seq[int] // nil
	if got := collect(s.Map(func(n int) int { return n })); got != nil {
		t.Errorf("Map nil: want nil, got %v", got)
	}
	if got := collect(s.FilterMap(func(int) (int, bool) { return 0, true })); got != nil {
		t.Errorf("FilterMap nil: want nil, got %v", got)
	}
	if got := collect(s.FlatMap(func(int) Seq[int] { return Of(1) })); got != nil {
		t.Errorf("FlatMap nil: want nil, got %v", got)
	}
	if got := collect(s.Tap(func(int) {})); got != nil {
		t.Errorf("Tap nil: want nil, got %v", got)
	}
	if got := collect(Flatten(Seq[Seq[int]](nil))); got != nil {
		t.Errorf("Flatten nil: want nil, got %v", got)
	}
}

func TestFlatMapNilSubstream(t *testing.T) {
	// f returns nil for the middle element; nil must behave as an empty
	// sub-stream, not panic.
	got := collect(Of(1, 2, 3).FlatMap(func(n int) Seq[int] {
		if n == 2 {
			return nil
		}
		return Of(n, n*10)
	}))
	want := []int{1, 10, 3, 30}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestFlattenNilSubstream(t *testing.T) {
	got := collect(Flatten(Of(Of(1, 2), Seq[int](nil), Of(3))))
	want := []int{1, 2, 3}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

// TestTapEarlyTermination asserts Tap's core contract: f is invoked only for
// elements the downstream actually reaches.
func TestTapEarlyTermination(t *testing.T) {
	calls := 0
	s := Generate(func(i int) int { return i }, 100).Tap(func(int) { calls++ })
	count := 0
	for range s {
		count++
		if count == 3 {
			break
		}
	}
	if calls != 3 {
		t.Fatalf("want 3 tap calls after early break, got %d", calls)
	}
}

// TestFilterMapLazyAndEarlyTermination covers FilterMap's skip path: f is called
// once per pulled element (including skipped ones) and stops once the
// downstream consumer stops.
func TestFilterMapLazyAndEarlyTermination(t *testing.T) {
	calls := 0
	s := Of(1, 2, 3, 4, 5, 6, 7, 8, 9, 10).FilterMap(func(n int) (int, bool) {
		calls++
		return n, n%2 == 0
	})
	if calls != 0 {
		t.Fatalf("lazy: want 0 calls before iteration, got %d", calls)
	}
	count := 0
	for range s {
		count++
		if count == 1 { // first even (2): pulls 1 (skip) then 2 (keep), stops
			break
		}
	}
	// one skipped (1) + one kept (2) = 2 calls; element 3 onward never pulled
	if calls != 2 {
		t.Fatalf("want 2 calls (1 skip + 1 keep), got %d", calls)
	}
}

func TestFlattenEarlyTermination(t *testing.T) {
	srcCalls := 0
	outer := Generate(func(i int) Seq[int] { srcCalls++; return Of(i*10, i*10+1) }, 100)
	count := 0
	for range Flatten(outer) {
		count++
		if count == 3 {
			break
		}
	}
	// 3 elements: sub 0 (0,1) fully + sub 1 (10,…) first element, then break.
	if srcCalls != 2 {
		t.Fatalf("want 2 outer pulls after early break, got %d", srcCalls)
	}
}
