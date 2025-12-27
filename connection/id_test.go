package connection

import (
	"testing"
)

func TestConnectionIDFactory(t *testing.T) {
	factory := connectionIDFactory()

	// Test that factory returns a function
	if factory == nil {
		t.Fatal("connectionIDFactory returned nil")
	}

	// Test that successive calls increment the ID
	id1 := factory()
	id2 := factory()
	id3 := factory()

	if id1 != 1 {
		t.Error("First connection ID should be 1")
	}

	if id2 <= id1 {
		t.Errorf("Second ID (%d) should be greater than first (%d)", id2, id1)
	}

	if id3 <= id2 {
		t.Errorf("Third ID (%d) should be greater than second (%d)", id3, id2)
	}

	// Verify that IDs increment by 1
	if id2 != id1+1 {
		t.Errorf("Expected ID to increment by 1, got %d to %d", id1, id2)
	}

	if id3 != id2+1 {
		t.Errorf("Expected ID to increment by 1, got %d to %d", id2, id3)
	}
}

func TestConnectionIDFactoryConcurrency(t *testing.T) {
	factory := connectionIDFactory()
	idsChan := make(chan ConnectionID, 100)

	// Launch multiple goroutines to get IDs concurrently
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				idsChan <- factory()
			}
		}()
	}

	// Collect all IDs
	ids := make([]ConnectionID, 0, 100)
	for i := 0; i < 100; i++ {
		ids = append(ids, <-idsChan)
	}

	// Verify all IDs are unique
	seenIDs := make(map[ConnectionID]bool)
	for _, id := range ids {
		if seenIDs[id] {
			t.Errorf("Duplicate connection ID found: %d", id)
		}
		seenIDs[id] = true
	}

	if len(seenIDs) != 100 {
		t.Errorf("Expected 100 unique IDs, got %d", len(seenIDs))
	}
}

func TestMultipleConnectionIDFactories(t *testing.T) {
	factory1 := connectionIDFactory()
	factory2 := connectionIDFactory()

	// Each factory should have its own counter
	id1a := factory1()
	id1b := factory1()
	id2a := factory2()
	id2b := factory2()

	if id1a != 1 || id2a != 1 {
		t.Error("Connection IDs should be 1")
	}

	// Both factories should increment independently
	if id1b != id1a+1 {
		t.Errorf("Factory1: expected %d, got %d", id1a+1, id1b)
	}

	if id2b != id2a+1 {
		t.Errorf("Factory2: expected %d, got %d", id2a+1, id2b)
	}
}
