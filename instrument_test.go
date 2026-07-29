package bt_test

import (
	"context"
	"strings"
	"testing"

	"github.com/ratlabs-io/bt-go"
)

func TestInstrumentDeepObservation(t *testing.T) {
	env := bt.NewEnv(context.Background())
	leafA := bt.NewAction(func(env bt.Env) bt.RunStatus { return bt.Success })
	leafB := bt.NewAction(func(env bt.Env) bt.RunStatus { return bt.Success })
	root := bt.NewSequence(leafA, leafB)

	var seen []bt.RunStatus
	var nodes int
	tree := bt.Instrument(root, func(node bt.Behavior, status bt.RunStatus) {
		nodes++
		seen = append(seen, status)
	})

	if tree.Tick(env) != bt.Success {
		t.Fatal("expected Success")
	}
	// Observing(A), Observing(B), Observing(Sequence) — at least 3 notifications.
	if nodes < 3 {
		t.Fatalf("expected deep observations >= 3, got %d", nodes)
	}
	for _, st := range seen {
		if st != bt.Success {
			t.Fatalf("unexpected status %v", st)
		}
	}

	// Original tree untouched (still ticks).
	if root.Tick(env) != bt.Success {
		t.Fatal("original root should still work")
	}
}

func TestInstrumentRecorderVisualize(t *testing.T) {
	env := bt.NewEnv(context.Background())
	tree := bt.NewSelector(
		bt.NewNamed("Fail", bt.NewAction(func(env bt.Env) bt.RunStatus { return bt.Failure })),
		bt.NewNamed("Win", bt.NewAction(func(env bt.Env) bt.RunStatus { return bt.Success })),
	)
	instr, rec := bt.InstrumentRecorder(tree, nil)
	if instr.Tick(env) != bt.Success {
		t.Fatal("expected Success")
	}
	out := rec.Visualize(instr)
	// Observing wrappers collapse in TreeVisualizer; real nodes keep statuses.
	if strings.Contains(out, "Observing") {
		t.Fatalf("Observing should be transparent in dumps, got:\n%s", out)
	}
	for _, want := range []string{"Win [Success]", "Fail [Failure]", "Selector [Success]"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q in dump:\n%s", want, out)
		}
	}
	if len(rec.GetStatusMap()) == 0 {
		t.Fatal("expected recorded statuses")
	}
}

func TestInstrumentNil(t *testing.T) {
	if bt.Instrument(nil, nil) != nil {
		t.Fatal("nil root")
	}
}
