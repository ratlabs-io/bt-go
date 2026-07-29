package bt_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/ratlabs-io/bt-go"
)

func TestGetAsAndMustGet(t *testing.T) {
	env := bt.NewEnv(context.Background())
	env.Set("health", 42)
	env.Set("name", "hero")

	h, ok := bt.GetAs[int](env, "health")
	if !ok || h != 42 {
		t.Fatalf("GetAs int = %v, %v", h, ok)
	}
	if _, ok := bt.GetAs[string](env, "health"); ok {
		t.Fatal("wrong type should fail")
	}
	if _, ok := bt.GetAs[int](env, "missing"); ok {
		t.Fatal("missing key should fail")
	}
	if bt.MustGet[string](env, "name") != "hero" {
		t.Fatal("MustGet string")
	}
}

func TestTypedKeys(t *testing.T) {
	const Health bt.Key[int] = "health"
	const Name bt.Key[string] = "name"
	env := bt.NewEnv(context.Background())

	bt.SetKey(env, Health, 99)
	bt.SetKey(env, Name, "hero")

	h, ok := bt.GetKey(env, Health)
	if !ok || h != 99 {
		t.Fatalf("GetKey = %v, %v", h, ok)
	}
	if bt.MustGetKey(env, Name) != "hero" {
		t.Fatal("MustGetKey")
	}
	// Same blackboard as string API.
	if v, ok := bt.GetAs[int](env, "health"); !ok || v != 99 {
		t.Fatal("Key and string APIs share the store")
	}
	if Health.String() != "health" {
		t.Fatal("Key.String")
	}
}

func TestMustGetPanics(t *testing.T) {
	env := bt.NewEnv(context.Background())
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	_ = bt.MustGet[int](env, "nope")
}

func TestNamedVisualizer(t *testing.T) {
	leaf := bt.NewAction(func(env bt.Env) bt.RunStatus { return bt.Success })
	named := bt.NewNamed("Patrol", leaf)
	out := bt.NewTreeVisualizer(named).Visualize()
	if !strings.Contains(out, "Patrol") {
		t.Fatalf("expected Patrol in %q", out)
	}
	env := bt.NewEnv(context.Background())
	if named.Tick(env) != bt.Success {
		t.Fatal("Named should pass through")
	}
}

func TestObserving(t *testing.T) {
	env := bt.NewEnv(context.Background())
	var saw bt.RunStatus
	var calls int
	child := bt.NewAction(func(env bt.Env) bt.RunStatus { return bt.Success })
	obs := bt.NewObserving(child, func(node bt.Behavior, status bt.RunStatus) {
		calls++
		saw = status
		if node != child {
			t.Errorf("observer should see child")
		}
	})
	if obs.Tick(env) != bt.Success {
		t.Fatal("expected Success")
	}
	if calls != 1 || saw != bt.Success {
		t.Fatalf("observer calls=%d status=%v", calls, saw)
	}
}

func TestAbortHookOnPreempt(t *testing.T) {
	env := bt.NewEnv(context.Background())
	var aborted bool

	low := bt.NewAbortHook(
		bt.NewAction(func(env bt.Env) bt.RunStatus { return bt.Running }),
		func(env bt.Env) { aborted = true },
	)
	highReady := false
	high := bt.NewAction(func(env bt.Env) bt.RunStatus {
		if highReady {
			return bt.Success
		}
		return bt.Failure
	})

	sel := bt.NewSelector(high, low)
	if sel.Tick(env) != bt.Running {
		t.Fatal("expected Running on low")
	}
	if aborted {
		t.Fatal("should not abort while still selected")
	}
	highReady = true
	if sel.Tick(env) != bt.Success {
		t.Fatal("expected high to win")
	}
	if !aborted {
		t.Fatal("low branch should have been Halted/aborted on preempt")
	}
}

func TestRepeaterReset(t *testing.T) {
	env := bt.NewEnv(context.Background())
	var n int
	child := bt.NewAction(func(env bt.Env) bt.RunStatus {
		n++
		return bt.Success
	})
	r := bt.NewRepeater(child, 3)
	_ = r.Tick(env)
	_ = r.Tick(env)
	if n != 2 {
		t.Fatalf("ticks=%d", n)
	}
	r.Reset()
	if r.Tick(env) != bt.Running {
		t.Fatal("after Reset expect Running")
	}
	if n != 3 {
		t.Fatalf("after Reset counter should restart, n=%d", n)
	}
}

func TestRunnerHaltsOnCancel(t *testing.T) {
	parent, cancel := context.WithCancel(context.Background())
	env := bt.NewEnv(parent)

	var halted bool
	hook := bt.NewAbortHook(
		bt.NewAction(func(env bt.Env) bt.RunStatus { return bt.Running }),
		func(env bt.Env) { halted = true },
	)
	runner := bt.NewTreeRunner(hook, bt.WithTickRate(time.Millisecond))

	done := make(chan struct{})
	go func() {
		runner.Run(env)
		close(done)
	}()

	time.Sleep(20 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("runner did not stop")
	}

	if !halted {
		t.Fatal("runner should Halt tree on cancel so AbortHook fires")
	}
}
