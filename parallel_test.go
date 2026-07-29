package bt_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ratlabs-io/bt-go"
)

func createCountingAction(counter *int, mutex *sync.Mutex, delay time.Duration, status bt.RunStatus) bt.Behavior {
	return bt.NewAction(func(env bt.Env) bt.RunStatus {
		if delay > 0 {
			time.Sleep(delay)
		}
		mutex.Lock()
		*counter++
		mutex.Unlock()
		return status
	})
}

func TestParallelPolicyString(t *testing.T) {
	if bt.RequireAll.String() != "RequireAll" {
		t.Errorf("unexpected policy string %q", bt.RequireAll.String())
	}
}

func TestParallelRequireOne(t *testing.T) {
	env := bt.NewEnv(context.Background())
	var counter int
	var mu sync.Mutex

	parallel := bt.NewParallel(bt.RequireOne,
		createCountingAction(&counter, &mu, 10*time.Millisecond, bt.Failure),
		createCountingAction(&counter, &mu, 20*time.Millisecond, bt.Failure),
	)
	if result := parallel.Tick(env); result != bt.Failure {
		t.Errorf("Expected Failure when all children fail, got %v", result)
	}
	if counter != 2 {
		t.Errorf("Expected all actions to run, got counter = %d", counter)
	}

	counter = 0
	parallel = bt.NewParallel(bt.RequireOne,
		createCountingAction(&counter, &mu, 10*time.Millisecond, bt.Failure),
		createCountingAction(&counter, &mu, 20*time.Millisecond, bt.Success),
	)
	if result := parallel.Tick(env); result != bt.Success {
		t.Errorf("Expected Success when at least one child succeeds, got %v", result)
	}
	if counter != 2 {
		t.Errorf("Expected all actions to run, got counter = %d", counter)
	}

	counter = 0
	parallel = bt.NewParallel(bt.RequireOne,
		createCountingAction(&counter, &mu, 10*time.Millisecond, bt.Failure),
		createCountingAction(&counter, &mu, 20*time.Millisecond, bt.Running),
	)
	if result := parallel.Tick(env); result != bt.Running {
		t.Errorf("Expected Running when at least one child is running, got %v", result)
	}
	if counter != 2 {
		t.Errorf("Expected all actions to run, got counter = %d", counter)
	}
}

func TestParallelRequireAll(t *testing.T) {
	env := bt.NewEnv(context.Background())
	var counter int
	var mu sync.Mutex

	parallel := bt.NewParallel(bt.RequireAll,
		createCountingAction(&counter, &mu, 10*time.Millisecond, bt.Success),
		createCountingAction(&counter, &mu, 20*time.Millisecond, bt.Failure),
	)
	if result := parallel.Tick(env); result != bt.Failure {
		t.Errorf("Expected Failure when any child fails, got %v", result)
	}

	counter = 0
	parallel = bt.NewParallel(bt.RequireAll,
		createCountingAction(&counter, &mu, 10*time.Millisecond, bt.Success),
		createCountingAction(&counter, &mu, 20*time.Millisecond, bt.Success),
	)
	if result := parallel.Tick(env); result != bt.Success {
		t.Errorf("Expected Success when all children succeed, got %v", result)
	}

	counter = 0
	parallel = bt.NewParallel(bt.RequireAll,
		createCountingAction(&counter, &mu, 10*time.Millisecond, bt.Success),
		createCountingAction(&counter, &mu, 20*time.Millisecond, bt.Running),
	)
	if result := parallel.Tick(env); result != bt.Running {
		t.Errorf("Expected Running when any child is running, got %v", result)
	}
}

func TestParallelSuccessOnAll(t *testing.T) {
	env := bt.NewEnv(context.Background())
	var counter int
	var mu sync.Mutex

	parallel := bt.NewParallel(bt.SuccessOnAll,
		createCountingAction(&counter, &mu, 0, bt.Success),
		createCountingAction(&counter, &mu, 0, bt.Failure),
	)
	if result := parallel.Tick(env); result != bt.Running {
		t.Errorf("Expected Running when any child fails, got %v", result)
	}

	counter = 0
	parallel = bt.NewParallel(bt.SuccessOnAll,
		createCountingAction(&counter, &mu, 0, bt.Success),
		createCountingAction(&counter, &mu, 0, bt.Success),
	)
	if result := parallel.Tick(env); result != bt.Success {
		t.Errorf("Expected Success when all children succeed, got %v", result)
	}
}

func TestParallelSuccessOnOne(t *testing.T) {
	env := bt.NewEnv(context.Background())
	var counter int
	var mu sync.Mutex

	parallel := bt.NewParallel(bt.SuccessOnOne,
		createCountingAction(&counter, &mu, 0, bt.Failure),
		createCountingAction(&counter, &mu, 0, bt.Failure),
	)
	if result := parallel.Tick(env); result != bt.Running {
		t.Errorf("Expected Running when all children fail, got %v", result)
	}

	counter = 0
	parallel = bt.NewParallel(bt.SuccessOnOne,
		createCountingAction(&counter, &mu, 0, bt.Failure),
		createCountingAction(&counter, &mu, 0, bt.Success),
	)
	if result := parallel.Tick(env); result != bt.Success {
		t.Errorf("Expected Success when at least one child succeeds, got %v", result)
	}
}

func TestParallelNilChild(t *testing.T) {
	env := bt.NewEnv(context.Background())
	var counter int
	var mu sync.Mutex

	parallel := bt.NewParallel(bt.RequireAll,
		createCountingAction(&counter, &mu, 0, bt.Success),
		nil,
	)
	if result := parallel.Tick(env); result != bt.Failure {
		t.Errorf("Expected Failure when a child is nil with RequireAll, got %v", result)
	}
	if counter != 1 {
		t.Errorf("Expected only valid actions to run, got counter = %d", counter)
	}
}

func TestParallelEmpty(t *testing.T) {
	env := bt.NewEnv(context.Background())
	if bt.NewParallel(bt.RequireAll).Tick(env) != bt.Success {
		t.Error("empty parallel should succeed")
	}
}

func TestParallelConcurrency(t *testing.T) {
	env := bt.NewEnv(context.Background())
	var counter int
	var mu sync.Mutex

	// Concurrent variant should overlap delays.
	parallel := bt.NewConcurrentParallel(bt.RequireAll,
		createCountingAction(&counter, &mu, 50*time.Millisecond, bt.Success),
		createCountingAction(&counter, &mu, 10*time.Millisecond, bt.Success),
		createCountingAction(&counter, &mu, 30*time.Millisecond, bt.Success),
	)

	start := time.Now()
	result := parallel.Tick(env)
	elapsed := time.Since(start)

	if result != bt.Success {
		t.Errorf("Expected Success, got %v", result)
	}
	if elapsed > 90*time.Millisecond {
		t.Errorf("Execution took too long (%v), suggesting sequential execution", elapsed)
	}
	if counter != 3 {
		t.Errorf("Expected all actions to run, got counter = %d", counter)
	}
	if !parallel.Concurrent() {
		t.Error("NewConcurrentParallel should report Concurrent true")
	}
}

func TestParallelSequentialDefault(t *testing.T) {
	env := bt.NewEnv(context.Background())
	var order []int
	var mu sync.Mutex

	mk := func(id int) bt.Behavior {
		return bt.NewAction(func(env bt.Env) bt.RunStatus {
			mu.Lock()
			order = append(order, id)
			mu.Unlock()
			return bt.Success
		})
	}
	p := bt.NewParallel(bt.RequireAll, mk(1), mk(2), mk(3))
	if p.Concurrent() {
		t.Fatal("NewParallel should be sequential by default")
	}
	if p.Tick(env) != bt.Success {
		t.Fatal("expected Success")
	}
	if len(order) != 3 || order[0] != 1 || order[1] != 2 || order[2] != 3 {
		t.Fatalf("expected sequential order 1,2,3 got %v", order)
	}
}

func TestParallelSharedContext(t *testing.T) {
	// Parallel children must share the parent context / blackboard.
	env := bt.NewEnv(context.Background())
	env.Set("seed", 10)

	var mu sync.Mutex
	var seen []int

	child := func(delta int) bt.Behavior {
		return bt.NewAction(func(env bt.Env) bt.RunStatus {
			v, _ := env.Get("seed")
			mu.Lock()
			seen = append(seen, v.(int)+delta)
			mu.Unlock()
			env.Set("from_child", delta)
			return bt.Success
		})
	}

	p := bt.NewParallel(bt.RequireAll, child(1), child(2))
	if p.Tick(env) != bt.Success {
		t.Fatal("expected Success")
	}
	if len(seen) != 2 {
		t.Fatalf("expected both children to see seed, got %v", seen)
	}
	for _, s := range seen {
		if s != 11 && s != 12 {
			t.Errorf("child did not see parent seed value, got %d", s)
		}
	}
	// At least one child write should be visible on parent context.
	if !env.Has("from_child") {
		t.Error("child Set should be visible on shared context")
	}
}

func TestParallelReticksAllChildren(t *testing.T) {
	// Children are re-ticked every Parallel.Tick (no sticky "completed" memory).
	env := bt.NewEnv(context.Background())
	var calls int32

	action := bt.NewAction(func(env bt.Env) bt.RunStatus {
		atomic.AddInt32(&calls, 1)
		return bt.Success
	})
	p := bt.NewParallel(bt.RequireAll, action)

	if p.Tick(env) != bt.Success {
		t.Fatal("first tick")
	}
	if p.Tick(env) != bt.Success {
		t.Fatal("second tick")
	}
	if atomic.LoadInt32(&calls) != 2 {
		t.Errorf("expected child ticked twice, got %d", calls)
	}
}

func TestParallelSuccessOnAllRecoversFromFailure(t *testing.T) {
	// Regression: sticky child status prevented re-trying a failed child.
	env := bt.NewEnv(context.Background())
	var phase int32

	flaky := bt.NewAction(func(env bt.Env) bt.RunStatus {
		if atomic.AddInt32(&phase, 1) == 1 {
			return bt.Failure
		}
		return bt.Success
	})
	ok := bt.NewAction(func(env bt.Env) bt.RunStatus {
		return bt.Success
	})

	p := bt.NewParallel(bt.SuccessOnAll, flaky, ok)
	if p.Tick(env) != bt.Running {
		t.Fatal("expected Running while flaky fails")
	}
	if p.Tick(env) != bt.Success {
		t.Fatal("expected Success once flaky recovers")
	}
}

func TestParallelStateTracking(t *testing.T) {
	env := bt.NewEnv(context.Background())
	var mu sync.Mutex

	stateValue := 0
	statefulAction := bt.NewAction(func(env bt.Env) bt.RunStatus {
		mu.Lock()
		defer mu.Unlock()
		if stateValue == 0 {
			stateValue = 1
			return bt.Running
		}
		return bt.Success
	})

	var actionCount int32
	countingAction := bt.NewAction(func(env bt.Env) bt.RunStatus {
		atomic.AddInt32(&actionCount, 1)
		return bt.Running
	})

	parallel := bt.NewParallel(bt.RequireAll, statefulAction, countingAction)

	if result := parallel.Tick(env); result != bt.Running {
		t.Errorf("Expected Running on first tick, got %v", result)
	}
	if stateValue != 1 {
		t.Errorf("Expected stateValue to be 1 after first tick, got %d", stateValue)
	}

	if result := parallel.Tick(env); result != bt.Running {
		t.Errorf("Expected Running on second tick, got %v", result)
	}
	if count := atomic.LoadInt32(&actionCount); count != 2 {
		t.Errorf("Expected countingAction to be called twice, got %d", count)
	}
}
