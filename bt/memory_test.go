package bt_test

import (
	"context"
	"strings"
	"testing"

	"github.com/ratlabs-io/bt-go/bt"
)

func TestMemorySequenceSkipsCompletedPrefix(t *testing.T) {
	var firstCalls, secondCalls int
	first := bt.NewAction(func(env bt.Env) bt.RunStatus {
		firstCalls++
		return bt.Success
	})
	second := bt.NewAction(func(env bt.Env) bt.RunStatus {
		secondCalls++
		if secondCalls < 2 {
			return bt.Running
		}
		return bt.Success
	})

	seq := bt.NewMemorySequence(first, second)
	env := bt.NewEnv(context.Background())

	if seq.Tick(env) != bt.Running {
		t.Fatal("expected Running on first tick")
	}
	if firstCalls != 1 {
		t.Fatalf("first child calls = %d, want 1", firstCalls)
	}
	if seq.RunningIndex() != 1 {
		t.Fatalf("RunningIndex = %d, want 1", seq.RunningIndex())
	}

	if seq.Tick(env) != bt.Success {
		t.Fatal("expected Success on second tick")
	}
	if firstCalls != 1 {
		t.Fatalf("memory sequence must not re-tick first child, got %d calls", firstCalls)
	}
	if secondCalls != 2 {
		t.Fatalf("second child calls = %d, want 2", secondCalls)
	}
	if seq.RunningIndex() != 0 {
		t.Fatalf("RunningIndex should reset after Success, got %d", seq.RunningIndex())
	}
}

func TestMemorySequenceFailureResets(t *testing.T) {
	var firstCalls int
	first := bt.NewAction(func(env bt.Env) bt.RunStatus {
		firstCalls++
		return bt.Success
	})
	failOnce := 0
	second := bt.NewAction(func(env bt.Env) bt.RunStatus {
		failOnce++
		if failOnce == 1 {
			return bt.Running
		}
		return bt.Failure
	})

	seq := bt.NewMemorySequence(first, second)
	env := bt.NewEnv(context.Background())

	if seq.Tick(env) != bt.Running {
		t.Fatal("expected Running")
	}
	if seq.Tick(env) != bt.Failure {
		t.Fatal("expected Failure")
	}
	if seq.RunningIndex() != 0 {
		t.Fatal("Failure should reset running index")
	}

	// After failure, sequence restarts from the first child.
	failOnce = 0
	secondOK := bt.NewAction(func(env bt.Env) bt.RunStatus {
		return bt.Success
	})
	// Replace second via new sequence to verify restart path cleanly.
	seq = bt.NewMemorySequence(first, secondOK)
	firstCalls = 0
	if seq.Tick(env) != bt.Success {
		t.Fatal("expected Success after rebuild")
	}
	if firstCalls != 1 {
		t.Fatalf("expected first child once, got %d", firstCalls)
	}
}

func TestMemorySequenceReset(t *testing.T) {
	var firstCalls int
	first := bt.NewAction(func(env bt.Env) bt.RunStatus {
		firstCalls++
		return bt.Success
	})
	second := bt.NewAction(func(env bt.Env) bt.RunStatus {
		return bt.Running
	})
	seq := bt.NewMemorySequence(first, second)
	env := bt.NewEnv(context.Background())

	_ = seq.Tick(env)
	if seq.RunningIndex() != 1 {
		t.Fatal("expected to be parked on second child")
	}
	seq.Reset()
	if seq.RunningIndex() != 0 {
		t.Fatal("Reset should clear index")
	}
	_ = seq.Tick(env)
	if firstCalls != 2 {
		t.Fatalf("after Reset, first child should run again, calls=%d", firstCalls)
	}
}

func TestMemorySequenceEmptyAndNil(t *testing.T) {
	env := bt.NewEnv(context.Background())
	if bt.NewMemorySequence().Tick(env) != bt.Success {
		t.Error("empty memory sequence should Success")
	}
	if bt.NewMemorySequence(nil).Tick(env) != bt.Failure {
		t.Error("nil child should Failure")
	}
}

func TestMemorySelectorSticksWithoutPreempt(t *testing.T) {
	// Contrast with reactive Selector: high priority does NOT preempt memory.
	highReady := false
	var highCalls, lowCalls int
	high := bt.NewAction(func(env bt.Env) bt.RunStatus {
		highCalls++
		if highReady {
			return bt.Success
		}
		return bt.Failure
	})
	low := bt.NewAction(func(env bt.Env) bt.RunStatus {
		lowCalls++
		if lowCalls < 3 {
			return bt.Running
		}
		return bt.Success
	})

	sel := bt.NewMemorySelector(high, low)
	env := bt.NewEnv(context.Background())

	if sel.Tick(env) != bt.Running {
		t.Fatal("expected Running (low)")
	}
	if sel.RunningIndex() != 1 {
		t.Fatalf("RunningIndex = %d, want 1", sel.RunningIndex())
	}

	// High becomes ready, but memory selector must stay on low.
	highReady = true
	if sel.Tick(env) != bt.Running {
		t.Fatal("expected still Running on low")
	}
	if highCalls != 1 {
		t.Fatalf("high must not be re-checked while low is remembered, highCalls=%d", highCalls)
	}
	if lowCalls != 2 {
		t.Fatalf("lowCalls=%d, want 2", lowCalls)
	}

	if sel.Tick(env) != bt.Success {
		t.Fatal("expected low to eventually Success")
	}
	if sel.RunningIndex() != -1 {
		t.Fatal("memory should clear after Success")
	}
}

func TestMemorySelectorFailureContinues(t *testing.T) {
	var midCalls, lastCalls int
	mid := bt.NewAction(func(env bt.Env) bt.RunStatus {
		midCalls++
		if midCalls == 1 {
			return bt.Running
		}
		return bt.Failure
	})
	last := bt.NewAction(func(env bt.Env) bt.RunStatus {
		lastCalls++
		return bt.Success
	})

	// first always fails so we land on mid.
	first := bt.NewAction(func(env bt.Env) bt.RunStatus {
		return bt.Failure
	})
	sel := bt.NewMemorySelector(first, mid, last)
	env := bt.NewEnv(context.Background())

	if sel.Tick(env) != bt.Running {
		t.Fatal("expected Running on mid")
	}
	// mid fails → try last without re-running first? Actually after mid fails we continue loop
	// and first is not re-tried in the same tick; next tick after Success/Failure of whole node.
	if sel.Tick(env) != bt.Success {
		t.Fatal("expected Success from last after mid fails")
	}
	if lastCalls != 1 {
		t.Fatalf("lastCalls=%d, want 1", lastCalls)
	}
}

func TestMemorySelectorAllFail(t *testing.T) {
	sel := bt.NewMemorySelector(
		bt.NewAction(func(env bt.Env) bt.RunStatus { return bt.Failure }),
		bt.NewAction(func(env bt.Env) bt.RunStatus { return bt.Failure }),
	)
	env := bt.NewEnv(context.Background())
	if sel.Tick(env) != bt.Failure {
		t.Fatal("expected Failure")
	}
	if sel.RunningIndex() != -1 {
		t.Fatal("index should be idle after total Failure")
	}
}

func TestMemorySelectorReset(t *testing.T) {
	var highCalls int
	high := bt.NewAction(func(env bt.Env) bt.RunStatus {
		highCalls++
		return bt.Failure
	})
	low := bt.NewAction(func(env bt.Env) bt.RunStatus {
		return bt.Running
	})
	sel := bt.NewMemorySelector(high, low)
	env := bt.NewEnv(context.Background())

	_ = sel.Tick(env)
	if sel.RunningIndex() != 1 {
		t.Fatal("expected parked on low")
	}
	sel.Reset()
	if sel.RunningIndex() != -1 {
		t.Fatal("Reset should idle")
	}
	_ = sel.Tick(env)
	if highCalls != 2 {
		t.Fatalf("after Reset high should run again, calls=%d", highCalls)
	}
}

func TestMemoryVsReactiveSequence(t *testing.T) {
	// Side-by-side: same children, different call counts.
	makeChildren := func(firstCalls, secondCalls *int) (bt.Behavior, bt.Behavior) {
		first := bt.NewAction(func(env bt.Env) bt.RunStatus {
			*firstCalls++
			return bt.Success
		})
		second := bt.NewAction(func(env bt.Env) bt.RunStatus {
			*secondCalls++
			if *secondCalls < 2 {
				return bt.Running
			}
			return bt.Success
		})
		return first, second
	}

	var rf, rs int
	f1, s1 := makeChildren(&rf, &rs)
	reactive := bt.NewSequence(f1, s1)

	var mf, ms int
	f2, s2 := makeChildren(&mf, &ms)
	memory := bt.NewMemorySequence(f2, s2)

	env := bt.NewEnv(context.Background())
	_ = reactive.Tick(env)
	_ = reactive.Tick(env)
	_ = memory.Tick(env)
	_ = memory.Tick(env)

	if rf != 2 {
		t.Errorf("reactive first calls = %d, want 2", rf)
	}
	if mf != 1 {
		t.Errorf("memory first calls = %d, want 1", mf)
	}
}

func TestMemoryNodesVisualize(t *testing.T) {
	tree := bt.NewSequence(
		bt.NewMemorySequence(bt.NewAction(func(env bt.Env) bt.RunStatus { return bt.Success })),
		bt.NewMemorySelector(bt.NewAction(func(env bt.Env) bt.RunStatus { return bt.Success })),
	)
	out := bt.NewTreeVisualizer(tree).Visualize()
	if !strings.Contains(out, "MemorySequence") || !strings.Contains(out, "MemorySelector") {
		t.Fatalf("visualize missing memory node names:\n%s", out)
	}
}
