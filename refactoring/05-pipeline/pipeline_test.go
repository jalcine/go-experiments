package pipeline

import "testing"

func TestProcessAll(t *testing.T) {
	in := []int{1, 2, 3, 4, 5}
	want := []int{1, 4, 9, 16, 25}

	got := ProcessAll(in)
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("result[%d] = %d, want %d", i, got[i], want[i])
		}
	}
}

func TestProcessAllEmpty(t *testing.T) {
	if got := ProcessAll(nil); len(got) != 0 {
		t.Errorf("ProcessAll(nil) returned %d results, want 0", len(got))
	}
}

func TestProcessAllPreservesOrder(t *testing.T) {
	in := make([]int, 100)
	for i := range in {
		in[i] = i
	}
	got := ProcessAll(in)
	for i := range in {
		if got[i] != i*i {
			t.Errorf("result[%d] = %d, want %d", i, got[i], i*i)
		}
	}
}
