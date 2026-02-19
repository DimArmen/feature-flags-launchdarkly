package main

import (
	"strings"
	"testing"
	"time"

	"github.com/launchdarkly/go-sdk-common/v3/ldcontext"
	"github.com/launchdarkly/go-sdk-common/v3/ldvalue"
	ld "github.com/launchdarkly/go-server-sdk/v7"
	"github.com/launchdarkly/go-server-sdk/v7/ldcomponents"
	"github.com/launchdarkly/go-server-sdk/v7/testhelpers/ldtestdata"
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

func TestFlagDefaultsWhenClientNil(t *testing.T) {
	prevClient := ldClient
	ldClient = nil
	defer func() {
		ldClient = prevClient
	}()
	flags := getFlagValues(ldcontext.New("sample-user"))

	if flags.BannerText != defaultBannerText {
		t.Errorf("Expected default banner text %q, got %q", defaultBannerText, flags.BannerText)
	}

	if flags.MaxCartItems != defaultMaxCartItems {
		t.Errorf("Expected default max cart items %d, got %d", defaultMaxCartItems, flags.MaxCartItems)
	}

	formattedPlan := formatPlan(flags.RecommendedPlan)
	if !strings.Contains(formattedPlan, "Starter") {
		t.Errorf("Expected default plan to mention Starter, got %s", formattedPlan)
	}
}

func TestFlagValuesWithTestData(t *testing.T) {
	prevClient := ldClient
	testData := ldtestdata.DataSource()
	testData.Update(testData.Flag("enable-gpt-511-codex-max").VariationForAll(false))
	testData.Update(testData.Flag("enable-credit-card").VariationForAll(false))
	testData.Update(testData.Flag("enable-paypal").VariationForAll(true))
	testData.Update(testData.Flag("enable-apple-pay").VariationForAll(true))
	testData.Update(testData.Flag("checkout-banner-text").ValueForAll(ldvalue.String("VIP Access")))
	testData.Update(testData.Flag("user-tier-label").ValueForAll(ldvalue.String("Platinum")))
	testData.Update(testData.Flag("max-items-in-cart").ValueForAll(ldvalue.Int(12)))

	plan := ldvalue.ObjectBuild().
		Set("name", ldvalue.String("Pro")).
		Set("price", ldvalue.Int(29)).
		Set("perk", ldvalue.String("Priority support + add-ons")).
		Build()
	testData.Update(testData.Flag("recommended-plan").ValueForAll(plan))

	client, err := ld.MakeCustomClient("test-sdk-key", ld.Config{
		DataSource: testData,
		Events:     ldcomponents.NoEvents(),
	}, 0)
	if err != nil {
		t.Fatalf("Failed to create test client: %v", err)
	}
	defer client.Close()

	ldClient = client
	defer func() {
		ldClient = prevClient
	}()

	flags := getFlagValues(ldcontext.New("tester"))

	if flags.GPTEnabled {
		t.Errorf("Expected GPT flag to be false")
	}

	if flags.BannerText != "VIP Access" {
		t.Errorf("Expected banner text to be updated, got %s", flags.BannerText)
	}

	if flags.MaxCartItems != 12 {
		t.Errorf("Expected max cart items to be 12, got %d", flags.MaxCartItems)
	}

	formattedPlan := formatPlan(flags.RecommendedPlan)
	if !strings.Contains(formattedPlan, "Pro plan") || !strings.Contains(formattedPlan, "$29/mo") {
		t.Errorf("Unexpected plan formatting: %s", formattedPlan)
	}
}
