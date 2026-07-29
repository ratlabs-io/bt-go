// Hello is a minimal bt-go example: a two-step Sequence that prints a greeting.
package main

import (
	"context"
	"fmt"

	"github.com/ratlabs-io/bt-go"
)

func main() {
	tree := bt.NewSequence(
		bt.NewAction(func(env bt.Env) bt.RunStatus {
			hello, ok := env.Get("greeting")
			if !ok {
				return bt.Failure
			}
			fmt.Printf("%s ", hello.(string))
			return bt.Success
		}),
		bt.NewAction(func(env bt.Env) bt.RunStatus {
			world, ok := env.Get("subject")
			if !ok {
				return bt.Failure
			}
			fmt.Printf("%s\n", world.(string))
			return bt.Success
		}),
	)

	env := bt.NewEnv(context.Background())
	env.Set("greeting", "Hello")
	env.Set("subject", "World")

	status := tree.Tick(env)
	fmt.Printf("Tree result: %s\n", status)

	// Optional: dump the tree structure.
	fmt.Print(bt.NewTreeVisualizer(tree).Visualize())
}
