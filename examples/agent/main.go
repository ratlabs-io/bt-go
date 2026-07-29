// Agent demonstrates reactive vs memory control flow for a simple NPC.
//
// Priority (reactive Selector): flee preempts combat/patrol when health is low.
// Combat uses a MemorySequence so attack wind-up is not restarted every tick.
// Patrol is the low-priority idle branch.
package main

import (
	"context"
	"fmt"

	"github.com/ratlabs-io/bt-go"
)

func main() {
	env := bt.NewEnv(context.Background())
	env.Set("health", 100)
	env.Set("enemy_visible", false)
	env.Set("state", "spawn")

	// Long-running combat step: counts ticks on the blackboard.
	attack := bt.NewNamed("Attack", bt.NewAction(func(env bt.Env) bt.RunStatus {
		progress, _ := bt.GetAs[int](env, "attack_progress")
		progress++
		env.Set("attack_progress", progress)
		fmt.Printf("  attack wind-up %d/3\n", progress)
		if progress < 3 {
			return bt.Running
		}
		env.Set("attack_progress", 0)
		env.Set("state", "struck")
		return bt.Success
	}))

	// Cleanup if combat is preempted (e.g. health drops mid-attack).
	combat := bt.NewAbortHook(
		bt.NewMemorySequence(
			bt.NewNamed("FaceEnemy", bt.NewAction(func(env bt.Env) bt.RunStatus {
				fmt.Println("  facing enemy")
				return bt.Success
			})),
			attack,
		),
		func(env bt.Env) {
			fmt.Println("  ! combat aborted — clearing wind-up")
			env.Set("attack_progress", 0)
		},
	)

	tree := bt.NewSelector(
		// Highest priority: flee when critically wounded (reactive re-check).
		bt.NewNamed("Flee", bt.NewSequence(
			bt.NewCondition(func(env bt.Env) bool {
				h, _ := bt.GetAs[int](env, "health")
				return h < 25
			}),
			bt.NewAction(func(env bt.Env) bt.RunStatus {
				fmt.Println("state: fleeing")
				env.Set("state", "fleeing")
				return bt.Success
			}),
		)),

		// Medium: fight if enemy visible (memory sequence inside).
		bt.NewNamed("Fight", bt.NewSequence(
			bt.NewCondition(func(env bt.Env) bool {
				v, _ := bt.GetAs[bool](env, "enemy_visible")
				return v
			}),
			combat,
		)),

		// Lowest: patrol.
		bt.NewNamed("Patrol", bt.NewAction(func(env bt.Env) bt.RunStatus {
			fmt.Println("state: patrolling")
			env.Set("state", "patrolling")
			return bt.Success
		})),
	)

	fmt.Println(bt.NewTreeVisualizer(tree).Visualize())
	fmt.Println("--- tick while healthy, no enemy ---")
	tree.Tick(env)

	fmt.Println("--- enemy appears; multi-tick memory attack ---")
	env.Set("enemy_visible", true)
	for i := 0; i < 4; i++ {
		fmt.Printf("tick %d → %s\n", i+1, tree.Tick(env))
	}

	fmt.Println("--- mid-fight health crash; flee preempts and aborts combat ---")
	env.Set("enemy_visible", true)
	env.Set("attack_progress", 0)
	// Start wind-up again.
	_ = tree.Tick(env)
	_ = tree.Tick(env)
	env.Set("health", 10)
	fmt.Printf("tick after injury → %s (state=%s)\n", tree.Tick(env), bt.MustGet[string](env, "state"))
}
