package seq

import (
	"reflect"
	"testing"
)

func TestFirst(t *testing.T) {
	t.Run("non-empty", func(t *testing.T) {
		v, ok := Of(5, 6, 7).First()
		if !ok || v != 5 {
			t.Fatalf("want 5,true; got %v,%v", v, ok)
		}
	})
	t.Run("single", func(t *testing.T) {
		v, ok := Of(42).First()
		if !ok || v != 42 {
			t.Fatalf("want 42,true; got %v,%v", v, ok)
		}
	})
	t.Run("empty", func(t *testing.T) {
		v, ok := Empty[int]().First()
		if ok || v != 0 {
			t.Fatalf("want 0,false; got %v,%v", v, ok)
		}
	})
	t.Run("nil", func(t *testing.T) {
		var s Seq[int]
		v, ok := s.First()
		if ok || v != 0 {
			t.Fatalf("want 0,false; got %v,%v", v, ok)
		}
	})
}

func TestLast(t *testing.T) {
	t.Run("non-empty", func(t *testing.T) {
		v, ok := Of(5, 6, 7).Last()
		if !ok || v != 7 {
			t.Fatalf("want 7,true; got %v,%v", v, ok)
		}
	})
	t.Run("single", func(t *testing.T) {
		v, ok := Of(42).Last()
		if !ok || v != 42 {
			t.Fatalf("want 42,true; got %v,%v", v, ok)
		}
	})
	t.Run("empty", func(t *testing.T) {
		v, ok := Empty[int]().Last()
		if ok || v != 0 {
			t.Fatalf("want 0,false; got %v,%v", v, ok)
		}
	})
	t.Run("nil", func(t *testing.T) {
		var s Seq[int]
		v, ok := s.Last()
		if ok || v != 0 {
			t.Fatalf("want 0,false; got %v,%v", v, ok)
		}
	})
}

func TestAt(t *testing.T) {
	t.Run("in bounds", func(t *testing.T) {
		v, ok := Of(10, 20, 30).At(1)
		if !ok || v != 20 {
			t.Fatalf("want 20,true; got %v,%v", v, ok)
		}
	})
	t.Run("index 0", func(t *testing.T) {
		v, ok := Of(10, 20, 30).At(0)
		if !ok || v != 10 {
			t.Fatalf("want 10,true; got %v,%v", v, ok)
		}
	})
	t.Run("last index", func(t *testing.T) {
		v, ok := Of(10, 20, 30).At(2)
		if !ok || v != 30 {
			t.Fatalf("want 30,true; got %v,%v", v, ok)
		}
	})
	t.Run("negative index", func(t *testing.T) {
		v, ok := Of(10, 20, 30).At(-1)
		if ok || v != 0 {
			t.Fatalf("want 0,false; got %v,%v", v, ok)
		}
	})
	t.Run("out of bounds", func(t *testing.T) {
		v, ok := Of(10, 20, 30).At(5)
		if ok || v != 0 {
			t.Fatalf("want 0,false; got %v,%v", v, ok)
		}
	})
	t.Run("empty", func(t *testing.T) {
		v, ok := Empty[int]().At(0)
		if ok || v != 0 {
			t.Fatalf("want 0,false; got %v,%v", v, ok)
		}
	})
	t.Run("nil", func(t *testing.T) {
		var s Seq[int]
		v, ok := s.At(0)
		if ok || v != 0 {
			t.Fatalf("want 0,false; got %v,%v", v, ok)
		}
	})
}

func TestPartition(t *testing.T) {
	even := func(n int) bool { return n%2 == 0 }
	t.Run("split", func(t *testing.T) {
		yes, no := Of(1, 2, 3, 4, 5, 6).Partition(even)
		if got := collect(yes); !reflect.DeepEqual(got, []int{2, 4, 6}) {
			t.Fatalf("yes: want [2 4 6], got %v", got)
		}
		if got := collect(no); !reflect.DeepEqual(got, []int{1, 3, 5}) {
			t.Fatalf("no: want [1 3 5], got %v", got)
		}
	})
	t.Run("all match in yes", func(t *testing.T) {
		yes, no := Of(2, 4, 6).Partition(even)
		if got := collect(yes); !reflect.DeepEqual(got, []int{2, 4, 6}) {
			t.Fatalf("yes: want [2 4 6], got %v", got)
		}
		if got := collect(no); got != nil {
			t.Fatalf("no: want nil, got %v", got)
		}
	})
	t.Run("all in no", func(t *testing.T) {
		yes, no := Of(1, 3, 5).Partition(even)
		if got := collect(yes); got != nil {
			t.Fatalf("yes: want nil, got %v", got)
		}
		if got := collect(no); !reflect.DeepEqual(got, []int{1, 3, 5}) {
			t.Fatalf("no: want [1 3 5], got %v", got)
		}
	})
	t.Run("empty", func(t *testing.T) {
		yes, no := Empty[int]().Partition(even)
		if got := collect(yes); got != nil {
			t.Fatalf("yes: want nil, got %v", got)
		}
		if got := collect(no); got != nil {
			t.Fatalf("no: want nil, got %v", got)
		}
	})
	t.Run("nil", func(t *testing.T) {
		var s Seq[int]
		yes, no := s.Partition(even)
		if got := collect(yes); got != nil {
			t.Fatalf("yes: want nil, got %v", got)
		}
		if got := collect(no); got != nil {
			t.Fatalf("no: want nil, got %v", got)
		}
	})
	t.Run("two streams reentrant", func(t *testing.T) {
		// 两条流各自独立迭代上游（要求上游可重入）；同一流可被多次消费。
		yes, no := Of(1, 2, 3).Partition(even)
		y1 := collect(yes)
		y2 := collect(yes)
		n1 := collect(no)
		if !reflect.DeepEqual(y1, []int{2}) || !reflect.DeepEqual(y2, []int{2}) {
			t.Fatalf("yes consumed twice: y1=%v y2=%v", y1, y2)
		}
		if !reflect.DeepEqual(n1, []int{1, 3}) {
			t.Fatalf("no: want [1 3], got %v", n1)
		}
	})
}
