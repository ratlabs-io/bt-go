package bt_test

import (
	"context"
	"fmt"

	"github.com/ratlabs-io/bt-go"
)

func ExampleNewSequence() {
	env := bt.NewEnv(context.Background())
	tree := bt.NewSequence(
		bt.NewAction(func(env bt.Env) bt.RunStatus {
			fmt.Println("step1")
			return bt.Success
		}),
		bt.NewAction(func(env bt.Env) bt.RunStatus {
			fmt.Println("step2")
			return bt.Success
		}),
	)
	fmt.Println(tree.Tick(env))
	// Output:
	// step1
	// step2
	// Success
}

func ExampleNewMemorySequence() {
	env := bt.NewEnv(context.Background())
	n := 0
	tree := bt.NewMemorySequence(
		bt.NewAction(func(env bt.Env) bt.RunStatus {
			fmt.Println("prep")
			return bt.Success
		}),
		bt.NewAction(func(env bt.Env) bt.RunStatus {
			n++
			fmt.Println("work", n)
			if n < 2 {
				return bt.Running
			}
			return bt.Success
		}),
	)
	fmt.Println(tree.Tick(env))
	fmt.Println(tree.Tick(env))
	// Output:
	// prep
	// work 1
	// Running
	// work 2
	// Success
}

func ExampleNewAbortHook() {
	env := bt.NewEnv(context.Background())
	low := bt.NewAbortHook(
		bt.NewAction(func(env bt.Env) bt.RunStatus { return bt.Running }),
		func(env bt.Env) { fmt.Println("aborted") },
	)
	highReady := false
	sel := bt.NewSelector(
		bt.NewAction(func(env bt.Env) bt.RunStatus {
			if highReady {
				return bt.Success
			}
			return bt.Failure
		}),
		low,
	)
	fmt.Println(sel.Tick(env))
	highReady = true
	fmt.Println(sel.Tick(env))
	// Output:
	// Running
	// aborted
	// Success
}

func ExampleGetAs() {
	env := bt.NewEnv(context.Background())
	env.Set("health", 42)
	h, ok := bt.GetAs[int](env, "health")
	fmt.Println(h, ok)
	// Output:
	// 42 true
}

func ExampleSetKey() {
	const Health bt.Key[int] = "health"
	env := bt.NewEnv(context.Background())
	bt.SetKey(env, Health, 100)
	fmt.Println(bt.MustGetKey(env, Health))
	// Output:
	// 100
}

func ExampleNewParallel() {
	env := bt.NewEnv(context.Background())
	p := bt.NewParallel(bt.RequireAll,
		bt.NewAction(func(env bt.Env) bt.RunStatus {
			fmt.Println("a")
			return bt.Success
		}),
		bt.NewAction(func(env bt.Env) bt.RunStatus {
			fmt.Println("b")
			return bt.Success
		}),
	)
	fmt.Println(p.Tick(env))
	// Output:
	// a
	// b
	// Success
}

func ExampleInstrument() {
	env := bt.NewEnv(context.Background())
	root := bt.NewSequence(
		bt.NewNamed("A", bt.NewAction(func(env bt.Env) bt.RunStatus { return bt.Success })),
	)
	var ticks int
	tree := bt.Instrument(root, func(node bt.Behavior, status bt.RunStatus) {
		ticks++
	})
	_ = tree.Tick(env)
	fmt.Println(ticks >= 1)
	// Output:
	// true
}
