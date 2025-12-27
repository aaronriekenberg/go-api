package request

import (
	"context"
	"testing"
)

func TestRequestIDFactory(t *testing.T) {
	factory := RequestIDFactory()

	// Test that factory returns a function
	if factory == nil {
		t.Fatal("RequestIDFactory returned nil")
	}

	// Test that successive calls increment the ID
	id1 := factory()
	id2 := factory()
	id3 := factory()

	if id1 != 1 {
		t.Error("First request ID should be 1")
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

func TestRequestIDFactoryConcurrency(t *testing.T) {
	factory := RequestIDFactory()
	idsChan := make(chan RequestID, 100)

	// Launch multiple goroutines to get IDs concurrently
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				idsChan <- factory()
			}
		}()
	}

	// Collect all IDs
	ids := make([]RequestID, 0, 100)
	for i := 0; i < 100; i++ {
		ids = append(ids, <-idsChan)
	}

	// Verify all IDs are unique
	seenIDs := make(map[RequestID]bool)
	for _, id := range ids {
		if seenIDs[id] {
			t.Errorf("Duplicate request ID found: %d", id)
		}
		seenIDs[id] = true
	}

	if len(seenIDs) != 100 {
		t.Errorf("Expected 100 unique IDs, got %d", len(seenIDs))
	}
}

func TestAddRequestIDToContext(t *testing.T) {
	ctx := context.Background()
	requestID := RequestID(12345)

	// Add request ID to context
	newCtx := AddRequestIDToContext(ctx, requestID)

	if newCtx == ctx {
		t.Error("AddRequestIDToContext should return a new context")
	}

	// Verify the ID can be retrieved
	retrievedID := RequestIDFromContext(newCtx)
	if retrievedID != requestID {
		t.Errorf("Expected request ID %d, got %d", requestID, retrievedID)
	}
}

func TestRequestIDFromContext_NotFound(t *testing.T) {
	ctx := context.Background()

	// Try to get request ID from context without adding it
	retrievedID := RequestIDFromContext(ctx)

	if retrievedID != 0 {
		t.Errorf("Expected request ID 0 for empty context, got %d", retrievedID)
	}
}

func TestRequestIDFromContext_WithOtherValues(t *testing.T) {
	ctx := context.Background()
	requestID := RequestID(99999)

	// Add the request ID to context
	ctx = AddRequestIDToContext(ctx, requestID)

	// Add other values to context
	type key string
	const otherKey key = "other"
	ctx = context.WithValue(ctx, otherKey, "some value")

	// Verify the request ID is still retrievable
	retrievedID := RequestIDFromContext(ctx)
	if retrievedID != requestID {
		t.Errorf("Expected request ID %d, got %d", requestID, retrievedID)
	}

	// Verify other values are still in context
	if ctx.Value(otherKey) != "some value" {
		t.Error("Other context values should not be affected")
	}
}

func TestMultipleFactories(t *testing.T) {
	factory1 := RequestIDFactory()
	factory2 := RequestIDFactory()

	// Each factory should have its own counter
	id1a := factory1()
	id1b := factory1()
	id2a := factory2()
	id2b := factory2()

	if id1a == 0 || id2a == 0 {
		t.Error("Request IDs should not be 0")
	}

	// IDs from different factories may overlap, which is fine
	// but both factories should increment independently
	if id1b != id1a+1 {
		t.Errorf("Factory1: expected %d, got %d", id1a+1, id1b)
	}

	if id2b != id2a+1 {
		t.Errorf("Factory2: expected %d, got %d", id2a+1, id2b)
	}
}

func TestContextChaining(t *testing.T) {
	ctx := context.Background()
	requestID := RequestID(55555)

	// Add request ID and create a cancellable context
	ctx = AddRequestIDToContext(ctx, requestID)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Verify the request ID is still accessible
	retrievedID := RequestIDFromContext(ctx)
	if retrievedID != requestID {
		t.Errorf("Expected request ID %d, got %d", requestID, retrievedID)
	}
}
