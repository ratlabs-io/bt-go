package bt_test

import (
	"context"
	"testing"

	"github.com/ratlabs-io/bt-go"
)

func abortCounter(t *testing.T) (bt.Behavior, *int) {
	t.Helper()
	var n int
	hook := bt.NewAbortHook(
		bt.NewAction(func(env bt.Env) bt.RunStatus { return bt.Running }),
		func(env bt.Env) { n++ },
	)
	return hook, &n
}

func TestSequenceHaltOnEarlierFailure(t *testing.T) {
	env := bt.NewEnv(context.Background())
	combat, halted := abortCounter(t)
	ok := true
	seq := bt.NewSequence(
		bt.NewCondition(func(env bt.Env) bool { return ok }),
		combat,
	)
	if seq.Tick(env) != bt.Running {
		t.Fatal("expected Running")
	}
	if *halted != 0 {
		t.Fatal("should not abort while condition holds")
	}
	ok = false
	if seq.Tick(env) != bt.Failure {
		t.Fatal("expected Failure when condition fails")
	}
	if *halted != 1 {
		t.Fatalf("combat should Halt on earlier sibling Failure, halted=%d", *halted)
	}
}

func TestSequenceNoFalseHaltOnProgress(t *testing.T) {
	// Child B succeeds after Running, C becomes Running — B must not OnAbort.
	env := bt.NewEnv(context.Background())
	var bHalted int
	bTicks := 0
	b := bt.NewAbortHook(
		bt.NewAction(func(env bt.Env) bt.RunStatus {
			bTicks++
			if bTicks == 1 {
				return bt.Running
			}
			return bt.Success
		}),
		func(env bt.Env) { bHalted++ },
	)
	c := bt.NewAction(func(env bt.Env) bt.RunStatus { return bt.Running })
	seq := bt.NewSequence(
		bt.NewAction(func(env bt.Env) bt.RunStatus { return bt.Success }),
		b,
		c,
	)
	if seq.Tick(env) != bt.Running {
		t.Fatal("tick1 Running")
	}
	if seq.Tick(env) != bt.Running {
		t.Fatal("tick2 Running")
	}
	if bHalted != 0 {
		t.Fatalf("progress Success→later Running must not AbortHook B, halted=%d", bHalted)
	}
}

func TestSequenceHaltPropagates(t *testing.T) {
	env := bt.NewEnv(context.Background())
	child, halted := abortCounter(t)
	seq := bt.NewSequence(child)
	_ = seq.Tick(env)
	bt.Halt(env, seq)
	if *halted != 1 {
		t.Fatalf("Sequence.Halt should abort Running child, halted=%d", *halted)
	}
}

func TestMemorySequenceIdleHaltIsNoop(t *testing.T) {
	env := bt.NewEnv(context.Background())
	var halted int
	child := bt.NewAbortHook(
		bt.NewAction(func(env bt.Env) bt.RunStatus { return bt.Success }),
		func(env bt.Env) { halted++ },
	)
	ms := bt.NewMemorySequence(child)
	if ms.Tick(env) != bt.Success {
		t.Fatal("expected Success")
	}
	if ms.RunningIndex() != -1 {
		t.Fatalf("idle index want -1, got %d", ms.RunningIndex())
	}
	bt.Halt(env, ms)
	if halted != 0 {
		t.Fatalf("idle MemorySequence.Halt must not OnAbort, halted=%d", halted)
	}
}

func TestMemorySequenceHaltWhileRunning(t *testing.T) {
	env := bt.NewEnv(context.Background())
	child, halted := abortCounter(t)
	ms := bt.NewMemorySequence(
		bt.NewAction(func(env bt.Env) bt.RunStatus { return bt.Success }),
		child,
	)
	if ms.Tick(env) != bt.Running {
		t.Fatal("expected Running")
	}
	bt.Halt(env, ms)
	if *halted != 1 {
		t.Fatalf("halted=%d", *halted)
	}
	if ms.RunningIndex() != -1 {
		t.Fatal("Halt should idle memory")
	}
}

func TestBinarySelectorHaltOnBranchFlip(t *testing.T) {
	env := bt.NewEnv(context.Background())
	ifTrue, halted := abortCounter(t)
	useTrue := true
	bin := bt.NewBinarySelector(
		bt.NewCondition(func(env bt.Env) bool { return useTrue }),
		ifTrue,
		bt.NewAction(func(env bt.Env) bt.RunStatus { return bt.Success }),
	)
	if bin.Tick(env) != bt.Running {
		t.Fatal("expected Running on IfTrue")
	}
	useTrue = false
	if bin.Tick(env) != bt.Success {
		t.Fatal("expected IfFalse Success")
	}
	if *halted != 1 {
		t.Fatalf("IfTrue should Halt on branch flip, halted=%d", *halted)
	}
}

func TestBinarySelectorHaltMethod(t *testing.T) {
	env := bt.NewEnv(context.Background())
	ifTrue, halted := abortCounter(t)
	bin := bt.NewBinarySelector(
		bt.NewCondition(func(env bt.Env) bool { return true }),
		ifTrue,
		bt.NewAction(func(env bt.Env) bt.RunStatus { return bt.Success }),
	)
	_ = bin.Tick(env)
	bt.Halt(env, bin)
	if *halted != 1 {
		t.Fatalf("Halt(bin) should reach IfTrue, halted=%d", *halted)
	}
}

func TestSwitchHaltOnCaseChange(t *testing.T) {
	env := bt.NewEnv(context.Background())
	caseA, halted := abortCounter(t)
	key := "a"
	sw := bt.NewSwitch(
		func(env bt.Env) string { return key },
		map[string]bt.Behavior{
			"a": caseA,
			"b": bt.NewAction(func(env bt.Env) bt.RunStatus { return bt.Success }),
		},
		nil,
	)
	if sw.Tick(env) != bt.Running {
		t.Fatal("expected Running on a")
	}
	key = "b"
	if sw.Tick(env) != bt.Success {
		t.Fatal("expected Success on b")
	}
	if *halted != 1 {
		t.Fatalf("case a should Halt on key change, halted=%d", *halted)
	}
}

func TestSwitchHaltMethod(t *testing.T) {
	env := bt.NewEnv(context.Background())
	caseA, halted := abortCounter(t)
	sw := bt.NewSwitch(
		func(env bt.Env) string { return "a" },
		map[string]bt.Behavior{"a": caseA},
		nil,
	)
	_ = sw.Tick(env)
	bt.Halt(env, sw)
	if *halted != 1 {
		t.Fatalf("Halt(switch) halted=%d", *halted)
	}
}

func TestConditionalHaltOnConditionFail(t *testing.T) {
	env := bt.NewEnv(context.Background())
	act, halted := abortCounter(t)
	ready := true
	c := bt.NewConditional(
		bt.NewCondition(func(env bt.Env) bool { return ready }),
		act,
	)
	if c.Tick(env) != bt.Running {
		t.Fatal("expected Running")
	}
	ready = false
	if c.Tick(env) != bt.Failure {
		t.Fatal("expected Failure")
	}
	if *halted != 1 {
		t.Fatalf("Action should Halt when condition fails, halted=%d", *halted)
	}
}

func TestConditionalHaltMethod(t *testing.T) {
	env := bt.NewEnv(context.Background())
	act, halted := abortCounter(t)
	c := bt.NewConditional(
		bt.NewCondition(func(env bt.Env) bool { return true }),
		act,
	)
	_ = c.Tick(env)
	bt.Halt(env, c)
	if *halted != 1 {
		t.Fatalf("Halt(conditional) halted=%d", *halted)
	}
}

func TestParallelHaltsResidualRunningOnSuccess(t *testing.T) {
	env := bt.NewEnv(context.Background())
	running, halted := abortCounter(t)
	p := bt.NewParallel(bt.RequireOne,
		bt.NewAction(func(env bt.Env) bt.RunStatus { return bt.Success }),
		running,
	)
	if p.Tick(env) != bt.Success {
		t.Fatal("RequireOne should Success")
	}
	if *halted != 1 {
		t.Fatalf("residual Running child should Halt on policy Success, halted=%d", *halted)
	}
}

func TestParallelNoHaltWhileStillRunning(t *testing.T) {
	env := bt.NewEnv(context.Background())
	var halted int
	a := bt.NewAbortHook(
		bt.NewAction(func(env bt.Env) bt.RunStatus { return bt.Running }),
		func(env bt.Env) { halted++ },
	)
	b := bt.NewAction(func(env bt.Env) bt.RunStatus { return bt.Running })
	p := bt.NewParallel(bt.RequireAll, a, b)
	if p.Tick(env) != bt.Running {
		t.Fatal("expected Running")
	}
	if halted != 0 {
		t.Fatalf("must not Halt while policy still Running, halted=%d", halted)
	}
}

func TestHaltAll(t *testing.T) {
	env := bt.NewEnv(context.Background())
	a, ha := abortCounter(t)
	b, hb := abortCounter(t)
	_ = a.Tick(env)
	_ = b.Tick(env)
	bt.HaltAll(env, a, b, nil)
	if *ha != 1 || *hb != 1 {
		t.Fatalf("HaltAll ha=%d hb=%d", *ha, *hb)
	}
}

func TestDecoratorHaltPropagates(t *testing.T) {
	env := bt.NewEnv(context.Background())
	child, halted := abortCounter(t)
	inv := bt.NewInverter(child)
	_ = inv.Tick(env)
	bt.Halt(env, inv)
	if *halted != 1 {
		t.Fatalf("Inverter.Halt halted=%d", *halted)
	}
}
