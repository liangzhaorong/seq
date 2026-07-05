package seq

import (
	"reflect"
	"testing"
	"time"
)

// collect is a test helper that drives a [Seq[T]] to completion, mirroring the
// future Collect method. It is used until issue #7 introduces Collect.
func collect[T any](s Seq[T]) []T {
	var out []T
	for v := range s {
		out = append(out, v)
	}
	return out
}

// takeN drives a [Seq[T]] and stops after consuming the first n elements,
// asserting that early termination propagates upstream.
func takeN[T any](s Seq[T], n int) []T {
	var out []T
	for v := range s {
		out = append(out, v)
		if len(out) == n {
			break
		}
	}
	return out
}

// collect2 is the Seq2 analogue of collect.
func collect2[K any, V any](s Seq2[K, V]) ([]K, []V) {
	var ks []K
	var vs []V
	for k, v := range s {
		ks = append(ks, k)
		vs = append(vs, v)
	}
	return ks, vs
}

func TestOf(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		want []int
	}{
		{"empty", []int{}, nil},
		{"single", []int{42}, []int{42}},
		{"multiple", []int{1, 2, 3}, []int{1, 2, 3}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := collect(Of(tc.in...))
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("want %v, got %v", tc.want, got)
			}
		})
	}
}

func TestFromSlice(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		want []string
	}{
		{"nil slice", nil, nil},
		{"empty slice", []string{}, nil},
		{"non-empty", []string{"a", "b", "c"}, []string{"a", "b", "c"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := collect(FromSlice(tc.in))
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("want %v, got %v", tc.want, got)
			}
		})
	}
}

func TestEmpty(t *testing.T) {
	got := collect(Empty[int]())
	if got != nil {
		t.Fatalf("want nil, got %v", got)
	}
}

func TestRange(t *testing.T) {
	tests := []struct {
		name        string
		start, stop int
		step        int
		want        []int
	}{
		{"ascending unit step", 0, 5, 1, []int{0, 1, 2, 3, 4}},
		{"positive step", 0, 10, 2, []int{0, 2, 4, 6, 8}},
		{"single element", 7, 8, 1, []int{7}},
		{"empty interval start==stop", 3, 3, 1, nil},
		{"empty interval start>stop", 5, 0, 1, nil},
		{"descending", 5, 0, -1, []int{5, 4, 3, 2, 1}},
		{"descending step -2", 10, 0, -2, []int{10, 8, 6, 4, 2}},
		{"from negative", -3, 3, 1, []int{-3, -2, -1, 0, 1, 2}},
		{"empty: neg step start<stop", 0, 5, -1, nil},
		{"empty: neg step start==stop", 3, 3, -1, nil},
		{"empty: pos step start>stop", 5, 0, 1, nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := collect(Range(tc.start, tc.stop, tc.step))
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("want %v, got %v", tc.want, got)
			}
		})
	}
}

func TestRangeZeroStepPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for zero step")
		}
	}()
	_ = Range(0, 10, 0)
}

func TestRepeat(t *testing.T) {
	tests := []struct {
		name string
		n    int
		want []string
	}{
		{"zero", 0, nil},
		{"negative", -3, nil},
		{"one", 1, []string{"x"}},
		{"three", 3, []string{"x", "x", "x"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := collect(Repeat("x", tc.n))
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("want %v, got %v", tc.want, got)
			}
		})
	}
}

func TestGenerate(t *testing.T) {
	tests := []struct {
		name string
		n    int
		want []int
	}{
		{"zero", 0, nil},
		{"negative", -2, nil},
		{"five squares", 5, []int{0, 1, 4, 9, 16}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := collect(Generate(func(i int) int { return i * i }, tc.n))
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("want %v, got %v", tc.want, got)
			}
		})
	}
}

// TestGenerateLazy asserts the laziness invariant (SPEC §9.1): the generator
// function is not invoked at all until the stream is consumed.
func TestGenerateLazy(t *testing.T) {
	calls := 0
	s := Generate(func(i int) int { calls++; return i }, 100)
	if calls != 0 {
		t.Fatalf("before iteration: want 0 calls, got %d", calls)
	}
	_ = collect(s)
	if calls != 100 {
		t.Fatalf("after full collect: want 100 calls, got %d", calls)
	}
}

// TestGenerateEarlyTermination asserts the early-termination invariant (SPEC
// §5.1, FR-5): breaking out of the iteration stops the upstream generator.
func TestGenerateEarlyTermination(t *testing.T) {
	calls := 0
	s := Generate(func(i int) int { calls++; return i }, 100)
	count := 0
	for range s {
		count++
		if count == 5 {
			break
		}
	}
	if calls != 5 {
		t.Fatalf("want 5 upstream calls after early break, got %d", calls)
	}
}

func TestCycle(t *testing.T) {
	t.Run("zero times", func(t *testing.T) {
		if got := collect(Cycle(Of(1, 2), 0)); got != nil {
			t.Fatalf("want nil, got %v", got)
		}
	})
	t.Run("three times", func(t *testing.T) {
		got := collect(Cycle(Of(1, 2), 3))
		want := []int{1, 2, 1, 2, 1, 2}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
	t.Run("nil source", func(t *testing.T) {
		if got := collect(Cycle[int](nil, 3)); got != nil {
			t.Fatalf("want nil, got %v", got)
		}
	})
}

// TestCycleInfinite drives an infinite stream (times < 0) and asserts that a
// short-circuiting break terminates it, yielding the repeated cycle prefix.
func TestCycleInfinite(t *testing.T) {
	s := Cycle(Of(1, 2, 3), -1)
	var got []int
	for v := range s {
		got = append(got, v)
		if len(got) == 7 {
			break
		}
	}
	want := []int{1, 2, 3, 1, 2, 3, 1}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestFromMap(t *testing.T) {
	t.Run("non-empty", func(t *testing.T) {
		m := map[string]int{"a": 1, "b": 2, "c": 3}
		got := map[string]int{}
		for k, v := range FromMap(m) {
			got[k] = v
		}
		if !reflect.DeepEqual(got, m) {
			t.Fatalf("want %v, got %v", m, got)
		}
	})
	t.Run("empty", func(t *testing.T) {
		count := 0
		for range FromMap(map[string]int{}) {
			count++
		}
		if count != 0 {
			t.Fatalf("want 0 pairs, got %d", count)
		}
	})
	t.Run("collect2 helper", func(t *testing.T) {
		ks, vs := collect2(FromMap(map[int]string{1: "a", 2: "b"}))
		if len(ks) != 2 || len(vs) != 2 {
			t.Fatalf("want 2 pairs, got keys=%v values=%v", ks, vs)
		}
	})
}

func TestUnboxFromSeqRoundtrip(t *testing.T) {
	s := Of(1, 2, 3)
	roundtripped := FromSeq(s.Unbox())
	got := collect(roundtripped)
	want := []int{1, 2, 3}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

// TestUnboxConsumedBySlices confirms Seq.Unbox yields an iter.Seq consumable by
// the standard library range-over-func machinery.
func TestUnboxConsumedBySlices(t *testing.T) {
	s := Of(3, 1, 2)
	sum := 0
	for v := range s.Unbox() {
		sum += v
	}
	if sum != 6 {
		t.Fatalf("want sum 6, got %d", sum)
	}
}

// TestEarlyTermination covers the short-circuit return path of each
// constructor: breaking out of the range must stop the upstream producer.
func TestEarlyTermination(t *testing.T) {
	t.Run("FromSlice", func(t *testing.T) {
		got := takeN(FromSlice([]int{1, 2, 3, 4, 5}), 3)
		want := []int{1, 2, 3}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
	t.Run("Range ascending", func(t *testing.T) {
		got := takeN(Range(0, 100, 1), 4)
		want := []int{0, 1, 2, 3}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
	t.Run("Range descending", func(t *testing.T) {
		got := takeN(Range(100, 0, -1), 3)
		want := []int{100, 99, 98}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
	t.Run("Repeat", func(t *testing.T) {
		got := takeN(Repeat("x", 1000), 3)
		want := []string{"x", "x", "x"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
	t.Run("FromMap", func(t *testing.T) {
		m := map[int]int{1: 1, 2: 2, 3: 3, 4: 4, 5: 5}
		count := 0
		for range FromMap(m) {
			count++
			if count == 2 {
				break
			}
		}
		if count != 2 {
			t.Fatalf("want 2 consumed before break, got %d", count)
		}
	})
	t.Run("Of", func(t *testing.T) {
		got := takeN(Of(10, 20, 30, 40), 2)
		want := []int{10, 20}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
}

// TestCycleInfiniteEmptySource locks the empty-source guard: an infinite cycle
// over a source that yields nothing terminates after the first empty pass
// instead of spinning forever.
func TestCycleInfiniteEmptySource(t *testing.T) {
	done := make(chan struct{})
	var count int
	go func() {
		defer close(done)
		for range Cycle(Empty[int](), -1) {
			count++
		}
	}()
	select {
	case <-done:
		if count != 0 {
			t.Fatalf("want 0 yielded, got %d", count)
		}
	case <-time.After(time.Second):
		t.Fatal("Cycle(Empty, -1) hung: empty-source guard regressed")
	}
}

func TestSeq2UnboxFromSeq2Roundtrip(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2}
	roundtripped := FromSeq2(FromMap(m).Unbox())
	got := map[string]int{}
	for k, v := range roundtripped {
		got[k] = v
	}
	if !reflect.DeepEqual(got, m) {
		t.Fatalf("want %v, got %v", m, got)
	}
}

func TestConstraints(t *testing.T) {
	// These instantiations compile only if Number and Ordered accept the given
	// types; they serve as a compile-time check that the constraints are well
	// formed and cover the intended numeric and ordered types.
	if got := numIdentity(42); got != 42 { // int ∈ Number
		t.Fatalf("numIdentity[int]: want 42, got %v", got)
	}
	if got := numIdentity[uint](7); got != 7 { // uint ∈ Number
		t.Fatalf("numIdentity[uint]: want 7, got %v", got)
	}
	if got := numIdentity(3.0); got != 3.0 { // float64 ∈ Number
		t.Fatalf("numIdentity[float64]: want 3.0, got %v", got)
	}
	if got := orderedIdentity(9); got != 9 { // int ∈ Ordered
		t.Fatalf("orderedIdentity[int]: want 9, got %v", got)
	}
	if got := orderedIdentity("s"); got != "s" { // string ∈ Ordered
		t.Fatalf("orderedIdentity[string]: want s, got %v", got)
	}
}

func numIdentity[T Number](v T) T      { return v }
func orderedIdentity[T Ordered](v T) T { return v }
