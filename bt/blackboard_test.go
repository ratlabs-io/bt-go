package bt_test

import (
	"context"
	"sync"
	"testing"

	"github.com/ratlabs-io/bt-go/bt"
)

func TestBlackboardBasic(t *testing.T) {
	bb := bt.NewBlackboard()

	bb.Set("key1", "value1")
	value, ok := bb.Get("key1")
	if !ok {
		t.Errorf("Expected to find key1 in blackboard")
	}
	if value != "value1" {
		t.Errorf("Expected value1, got %v", value)
	}

	if !bb.Has("key1") {
		t.Errorf("Expected Has to return true for key1")
	}
	if bb.Has("nonexistent") {
		t.Errorf("Expected Has to return false for nonexistent key")
	}

	bb.Delete("key1")
	if bb.Has("key1") {
		t.Errorf("Expected key1 to be deleted")
	}

	bb.Set("key1", "value1")
	bb.Set("key2", "value2")
	bb.Clear()
	if bb.Has("key1") || bb.Has("key2") {
		t.Errorf("Expected all keys to be cleared")
	}
}

func TestBlackboardHierarchy(t *testing.T) {
	parent := bt.NewBlackboard()
	child := bt.NewBlackboardWithParent(parent)

	parent.Set("parentKey", "parentValue")
	child.Set("childKey", "childValue")

	value, ok := child.Get("parentKey")
	if !ok {
		t.Errorf("Child should be able to access parent's values")
	}
	if value != "parentValue" {
		t.Errorf("Expected parentValue, got %v", value)
	}

	_, ok = parent.Get("childKey")
	if ok {
		t.Errorf("Parent should not be able to access child's values")
	}

	if child.HasLocal("parentKey") {
		t.Errorf("HasLocal should return false for parent keys")
	}
	if !child.Has("parentKey") {
		t.Errorf("Has should return true for parent keys")
	}

	child.Set("parentKey", "overriddenValue")
	value, _ = child.Get("parentKey")
	if value != "overriddenValue" {
		t.Errorf("Child should override parent values, got %v", value)
	}
	value, _ = parent.Get("parentKey")
	if value != "parentValue" {
		t.Errorf("Parent should keep its original value, got %v", value)
	}

	child.Set("childKey2", "childValue2")

	entries := child.Entries()
	if len(entries) != 3 { // childKey, childKey2, parentKey (overridden)
		t.Errorf("Expected 3 local entries, got %d", len(entries))
	}

	allEntries := child.AllEntries()
	if len(allEntries) != 3 { // local keys only; parentKey already shadowed
		t.Errorf("Expected 3 total entries, got %d", len(allEntries))
	}

	// Parent-only key appears in AllEntries but not Entries.
	parent.Set("onlyParent", true)
	if child.HasLocal("onlyParent") {
		t.Error("onlyParent must not be local to child")
	}
	if !child.Has("onlyParent") {
		t.Error("child should see onlyParent via parent")
	}
	if len(child.AllEntries()) != 4 {
		t.Errorf("Expected 4 AllEntries after parent-only key, got %d", len(child.AllEntries()))
	}
}

func TestBlackboardWithEnv(t *testing.T) {
	bb := bt.NewBlackboard()
	bb.Set("testKey", "testValue")

	env := bt.NewEnv(context.Background(), bt.WithBlackboard(bb))

	envBB := env.Blackboard()
	if envBB == nil {
		t.Errorf("Expected env to have a blackboard")
	}

	value, ok := envBB.Get("testKey")
	if !ok {
		t.Errorf("Expected to find testKey in env's blackboard")
	}
	if value != "testValue" {
		t.Errorf("Expected testValue, got %v", value)
	}

	// Env Set/Get and blackboard are the same store.
	if v, ok := env.Get("testKey"); !ok || v != "testValue" {
		t.Errorf("env.Get should read blackboard values")
	}
	env.Set("viaEnv", 42)
	if v, ok := bb.Get("viaEnv"); !ok || v != 42 {
		t.Errorf("env.Set should write through to the shared blackboard")
	}

	bb.Set("newKey", "newValue")
	value, ok = envBB.Get("newKey")
	if !ok {
		t.Errorf("Expected to find newKey after updating original blackboard")
	}
	if value != "newValue" {
		t.Errorf("Expected newValue, got %v", value)
	}
}

func TestWithBlackboardNil(t *testing.T) {
	env := bt.NewEnv(context.Background(), bt.WithBlackboard(nil))
	if env.Blackboard() == nil {
		t.Fatal("nil WithBlackboard should keep a default blackboard")
	}
	env.Set("x", 1)
	if !env.Has("x") {
		t.Fatal("default blackboard should work after nil option")
	}
}

func TestBlackboardConcurrent(t *testing.T) {
	bb := bt.NewBlackboard()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := "k"
			bb.Set(key, n)
			bb.Get(key)
			bb.Has(key)
		}(i)
	}
	wg.Wait()
}
