package bt_test

import (
	"context"
	"testing"

	"github.com/ratlabs-io/bt-go/bt"
)

func TestBlackboardBasic(t *testing.T) {
	// Create a new blackboard
	bb := bt.NewBlackboard()

	// Test Set and Get
	bb.Set("key1", "value1")
	value, ok := bb.Get("key1")
	if !ok {
		t.Errorf("Expected to find key1 in blackboard")
	}
	if value != "value1" {
		t.Errorf("Expected value1, got %v", value)
	}

	// Test Has
	if !bb.Has("key1") {
		t.Errorf("Expected Has to return true for key1")
	}
	if bb.Has("nonexistent") {
		t.Errorf("Expected Has to return false for nonexistent key")
	}

	// Test Delete
	bb.Delete("key1")
	if bb.Has("key1") {
		t.Errorf("Expected key1 to be deleted")
	}

	// Test Clear
	bb.Set("key1", "value1")
	bb.Set("key2", "value2")
	bb.Clear()
	if bb.Has("key1") || bb.Has("key2") {
		t.Errorf("Expected all keys to be cleared")
	}
}

func TestBlackboardHierarchy(t *testing.T) {
	// Create a parent and child blackboard
	parent := bt.NewBlackboard()
	child := bt.NewBlackboardWithParent(parent)

	// Test parent-child relationship
	parent.Set("parentKey", "parentValue")
	child.Set("childKey", "childValue")

	// Child should be able to access parent's values
	value, ok := child.Get("parentKey")
	if !ok {
		t.Errorf("Child should be able to access parent's values")
	}
	if value != "parentValue" {
		t.Errorf("Expected parentValue, got %v", value)
	}

	// Parent should not be able to access child's values
	_, ok = parent.Get("childKey")
	if ok {
		t.Errorf("Parent should not be able to access child's values")
	}

	// Test HasLocal vs Has
	if child.HasLocal("parentKey") {
		t.Errorf("HasLocal should return false for parent keys")
	}
	if !child.Has("parentKey") {
		t.Errorf("Has should return true for parent keys")
	}

	// Test overriding parent values
	child.Set("parentKey", "overriddenValue")
	value, _ = child.Get("parentKey")
	if value != "overriddenValue" {
		t.Errorf("Child should override parent values, got %v", value)
	}
	value, _ = parent.Get("parentKey")
	if value != "parentValue" {
		t.Errorf("Parent should keep its original value, got %v", value)
	}

	// Test Entries vs AllEntries
	child.Set("childKey2", "childValue2")

	entries := child.Entries()
	if len(entries) != 3 { // childKey, childKey2, parentKey (overridden)
		t.Errorf("Expected 3 local entries, got %d", len(entries))
	}

	allEntries := child.AllEntries()
	if len(allEntries) != 3 { // 3 local keys (including the overridden parentKey)
		t.Errorf("Expected 3 total entries, got %d", len(allEntries))
	}
}

func TestBlackboardWithContext(t *testing.T) {
	// Create a blackboard and context
	bb := bt.NewBlackboard()
	bb.Set("testKey", "testValue")

	ctx := bt.NewBehaviorContext(context.Background(), bt.WithBlackboard(bb))

	// Check if the context has the blackboard
	ctxBB := ctx.GetBlackboard()
	if ctxBB == nil {
		t.Errorf("Expected context to have a blackboard")
	}

	// Check if the blackboard has the correct values
	value, ok := ctxBB.Get("testKey")
	if !ok {
		t.Errorf("Expected to find testKey in context's blackboard")
	}
	if value != "testValue" {
		t.Errorf("Expected testValue, got %v", value)
	}

	// Test that changes to the blackboard are reflected in the context
	bb.Set("newKey", "newValue")
	value, ok = ctxBB.Get("newKey")
	if !ok {
		t.Errorf("Expected to find newKey in context's blackboard after updating the original blackboard")
	}
	if value != "newValue" {
		t.Errorf("Expected newValue, got %v", value)
	}

	// Test that changes to the context's blackboard are reflected in the original blackboard
	ctxBB.Set("contextKey", "contextValue")
	value, ok = bb.Get("contextKey")
	if !ok {
		t.Errorf("Expected to find contextKey in original blackboard after updating via context")
	}
	if value != "contextValue" {
		t.Errorf("Expected contextValue, got %v", value)
	}
}
