package seq

import (
	"testing"
)

func TestFindBy(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		got := Of(1, 2, 3, 4).FindBy(func(n int) bool { return n > 2 }, -1)
		if got != 3 {
			t.Fatalf("want 3, got %v", got)
		}
	})
	t.Run("not found returns def", func(t *testing.T) {
		if got := Of(1, 2, 3).FindBy(func(n int) bool { return n > 10 }, -1); got != -1 {
			t.Fatalf("want -1, got %v", got)
		}
	})
	t.Run("empty returns def", func(t *testing.T) {
		if got := Empty[int]().FindBy(func(n int) bool { return n > 0 }, 42); got != 42 {
			t.Fatalf("want 42, got %v", got)
		}
	})
	t.Run("nil returns def", func(t *testing.T) {
		var s Seq[int]
		if got := s.FindBy(func(int) bool { return true }, 7); got != 7 {
			t.Fatalf("want 7, got %v", got)
		}
	})
	t.Run("short circuit", func(t *testing.T) {
		calls := 0
		got := Generate(func(i int) int { calls++; return i }, 100).
			FindBy(func(n int) bool { return n == 5 }, -1)
		if got != 5 {
			t.Fatalf("want 5, got %v", got)
		}
		// 命中 i=5 → 上游调用 6 次（0..5）
		if calls != 6 {
			t.Fatalf("want 6 upstream calls (short circuit at 5), got %d", calls)
		}
	})
}

func TestAny(t *testing.T) {
	gt5 := func(n int) bool { return n > 5 }
	t.Run("hit", func(t *testing.T) {
		if !Of(1, 2, 7, 3).Any(gt5) {
			t.Fatal("want true")
		}
	})
	t.Run("miss", func(t *testing.T) {
		if Of(1, 2, 3).Any(gt5) {
			t.Fatal("want false")
		}
	})
	t.Run("empty", func(t *testing.T) {
		if Empty[int]().Any(gt5) {
			t.Fatal("want false")
		}
	})
	t.Run("nil", func(t *testing.T) {
		var s Seq[int]
		if s.Any(gt5) {
			t.Fatal("want false")
		}
	})
}

func TestAnyShortCircuit(t *testing.T) {
	calls := 0
	hit := Generate(func(i int) int { calls++; return i }, 100).Any(func(n int) bool { return n >= 3 })
	if !hit {
		t.Fatal("want true")
	}
	// 命中 i=3 → 上游调用 4 次（0,1,2,3）
	if calls != 4 {
		t.Fatalf("want 4 upstream calls (short circuit at 3), got %d", calls)
	}
}

func TestAll(t *testing.T) {
	lt5 := func(n int) bool { return n < 5 }
	t.Run("all satisfy", func(t *testing.T) {
		if !Of(1, 2, 3, 4).All(lt5) {
			t.Fatal("want true")
		}
	})
	t.Run("one fails", func(t *testing.T) {
		if Of(1, 2, 6, 3).All(lt5) {
			t.Fatal("want false")
		}
	})
	t.Run("empty is true", func(t *testing.T) {
		if !Empty[int]().All(lt5) {
			t.Fatal("want true (vacuous)")
		}
	})
	t.Run("nil is true", func(t *testing.T) {
		var s Seq[int]
		if !s.All(lt5) {
			t.Fatal("want true")
		}
	})
}

func TestAllShortCircuit(t *testing.T) {
	calls := 0
	ok := Generate(func(i int) int { calls++; return i }, 100).All(func(n int) bool { return n < 3 })
	if ok {
		t.Fatal("want false")
	}
	// i=3 不满足 → 上游调用 4 次（0,1,2,3）
	if calls != 4 {
		t.Fatalf("want 4 upstream calls (short circuit at 3), got %d", calls)
	}
}

func TestNone(t *testing.T) {
	gt5 := func(n int) bool { return n > 5 }
	t.Run("none satisfy", func(t *testing.T) {
		if !Of(1, 2, 3).None(gt5) {
			t.Fatal("want true")
		}
	})
	t.Run("one satisfies", func(t *testing.T) {
		if Of(1, 6, 2).None(gt5) {
			t.Fatal("want false")
		}
	})
	t.Run("empty is true", func(t *testing.T) {
		if !Empty[int]().None(gt5) {
			t.Fatal("want true")
		}
	})
	t.Run("nil is true", func(t *testing.T) {
		var s Seq[int]
		if !s.None(gt5) {
			t.Fatal("want true")
		}
	})
}

func TestContainsBy(t *testing.T) {
	t.Run("hit", func(t *testing.T) {
		if !Of(1, 2, 3).ContainsBy(func(n int) bool { return n == 2 }) {
			t.Fatal("want true")
		}
	})
	t.Run("miss", func(t *testing.T) {
		if Of(1, 2, 3).ContainsBy(func(n int) bool { return n == 9 }) {
			t.Fatal("want false")
		}
	})
	t.Run("empty", func(t *testing.T) {
		if Empty[int]().ContainsBy(func(int) bool { return true }) {
			t.Fatal("want false")
		}
	})
	t.Run("nil", func(t *testing.T) {
		var s Seq[int]
		if s.ContainsBy(func(int) bool { return true }) {
			t.Fatal("want false")
		}
	})
}

func TestContains(t *testing.T) {
	t.Run("hit", func(t *testing.T) {
		if !Contains(Of(1, 2, 3), 2) {
			t.Fatal("want true")
		}
	})
	t.Run("miss", func(t *testing.T) {
		if Contains(Of(1, 2, 3), 9) {
			t.Fatal("want false")
		}
	})
	t.Run("empty", func(t *testing.T) {
		if Contains(Empty[int](), 1) {
			t.Fatal("want false")
		}
	})
	t.Run("nil", func(t *testing.T) {
		if Contains[int](nil, 1) {
			t.Fatal("want false")
		}
	})
	t.Run("strings", func(t *testing.T) {
		if !Contains(Of("a", "b", "c"), "b") {
			t.Fatal("want true")
		}
	})
}

func TestContainsShortCircuit(t *testing.T) {
	calls := 0
	src := Generate(func(i int) int { calls++; return i }, 100)
	if !Contains(src, 5) {
		t.Fatal("want true")
	}
	// 命中 i=5 → 上游调用 6 次（0..5）
	if calls != 6 {
		t.Fatalf("want 6 upstream calls (short circuit at 5), got %d", calls)
	}
}

func TestIndexOf(t *testing.T) {
	t.Run("hit first", func(t *testing.T) {
		if got := IndexOf(Of(1, 2, 3, 2), 2); got != 1 {
			t.Fatalf("want 1, got %v", got)
		}
	})
	t.Run("hit last", func(t *testing.T) {
		if got := IndexOf(Of(1, 2, 3), 3); got != 2 {
			t.Fatalf("want 2, got %v", got)
		}
	})
	t.Run("miss returns -1", func(t *testing.T) {
		if got := IndexOf(Of(1, 2, 3), 9); got != -1 {
			t.Fatalf("want -1, got %v", got)
		}
	})
	t.Run("empty returns -1", func(t *testing.T) {
		if got := IndexOf(Empty[int](), 1); got != -1 {
			t.Fatalf("want -1, got %v", got)
		}
	})
	t.Run("nil returns -1", func(t *testing.T) {
		if got := IndexOf[int](nil, 1); got != -1 {
			t.Fatalf("want -1, got %v", got)
		}
	})
}

func TestFind(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		if got := Find(Of(1, 2, 3), 2, -1); got != 2 {
			t.Fatalf("want 2, got %v", got)
		}
	})
	t.Run("not found returns def", func(t *testing.T) {
		if got := Find(Of(1, 2, 3), 9, -1); got != -1 {
			t.Fatalf("want -1, got %v", got)
		}
	})
	t.Run("empty returns def", func(t *testing.T) {
		if got := Find(Empty[int](), 1, 42); got != 42 {
			t.Fatalf("want 42, got %v", got)
		}
	})
	t.Run("nil returns def", func(t *testing.T) {
		if got := Find[int](nil, 1, 42); got != 42 {
			t.Fatalf("want 42, got %v", got)
		}
	})
}

// TestContainsByNoneEquivalent 锁定 ContainsBy == Any、None == !Any 的语义等价。
func TestContainsByNoneEquivalent(t *testing.T) {
	s := Of(1, 2, 3, 4, 5)
	p := func(n int) bool { return n%2 == 0 }
	if s.ContainsBy(p) != s.Any(p) {
		t.Fatal("ContainsBy should equal Any")
	}
	if s.None(p) == s.Any(p) {
		t.Fatal("None should be !Any")
	}
	// 全不命中谓词：None 须为 true、ContainsBy 须为 false（覆盖空命中集边界）
	never := func(int) bool { return false }
	if !s.None(never) {
		t.Fatal("None(never) should be true")
	}
	if s.ContainsBy(never) {
		t.Fatal("ContainsBy(never) should be false")
	}
}

func TestIndexOfShortCircuit(t *testing.T) {
	calls := 0
	src := Generate(func(i int) int { calls++; return i }, 100)
	if got := IndexOf(src, 5); got != 5 {
		t.Fatalf("want index 5, got %v", got)
	}
	// 命中 i=5 → 上游调用 6 次（0..5）
	if calls != 6 {
		t.Fatalf("want 6 upstream calls (short circuit at 5), got %d", calls)
	}
}

func TestFindShortCircuit(t *testing.T) {
	calls := 0
	src := Generate(func(i int) int { calls++; return i }, 100)
	if got := Find(src, 5, -1); got != 5 {
		t.Fatalf("want 5, got %v", got)
	}
	// 命中 i=5 → 上游调用 6 次（0..5）
	if calls != 6 {
		t.Fatalf("want 6 upstream calls (short circuit at 5), got %d", calls)
	}
}
