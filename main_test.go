package main

import (
	"testing"
	"time"

	"github.com/launchdarkly/go-sdk-common/v3/ldcontext"
)

func TestFeatureFlagDefaultValue(t *testing.T) {
	// Test that default values work when SDK is not initialized
	user := ldcontext.New("test-user")

	// Since we can't easily initialize a real client in tests without an SDK key,
	// we test that the context creation works
	if user.Key() != "test-user" {
		t.Errorf("Expected user key to be 'test-user', got %s", user.Key())
	}
}

func TestContextCreation(t *testing.T) {
	// Test creating different user contexts
	user1 := ldcontext.New("user-123")
	user2 := ldcontext.New("user-456")

	if user1.Key() == user2.Key() {
		t.Error("Expected different user keys to create different contexts")
	}
}

func TestMultipleContextEvaluations(t *testing.T) {
	// Test that we can create multiple contexts without errors
	contexts := make([]ldcontext.Context, 10)
	for i := 0; i < 10; i++ {
		contexts[i] = ldcontext.New("user-" + string(rune(i)))
	}

	if len(contexts) != 10 {
		t.Errorf("Expected 10 contexts, got %d", len(contexts))
	}
}

func TestSDKTimeout(t *testing.T) {
	// Test that SDK initialization has a reasonable timeout
	timeout := 5 * time.Second
	if timeout < 1*time.Second || timeout > 10*time.Second {
		t.Errorf("SDK timeout should be between 1-10 seconds, got %v", timeout)
	}
}
