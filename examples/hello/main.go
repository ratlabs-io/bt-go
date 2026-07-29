// Hello is a minimal bt-go example: a two-step Sequence that prints a greeting.
package main

import (
	"context"
	"fmt"

	"github.com/ratlabs-io/bt-go/bt"
)

func main() {
	tree := bt.NewSequence(
		bt.NewAction(func(ctx bt.BehaviorContext) bt.RunStatus {
			hello, ok := ctx.Get("greeting")
			if !ok {
				return bt.Failure
			}
			fmt.Printf("%s ", hello.(string))
			return bt.Success
		}),
		bt.NewAction(func(ctx bt.BehaviorContext) bt.RunStatus {
			world, ok := ctx.Get("subject")
			if !ok {
				return bt.Failure
			}
			fmt.Printf("%s\n", world.(string))
			return bt.Success
		}),
	)

	ctx := bt.NewBehaviorContext(context.Background())
	ctx.Set("greeting", "Hello")
	ctx.Set("subject", "World")

	status := tree.Tick(ctx)
	fmt.Printf("Tree result: %s\n", status)

	// Optional: dump the tree structure.
	fmt.Print(bt.NewTreeVisualizer(tree).Visualize())
}
