package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/launchdarkly/go-sdk-common/v3/ldcontext"
	"github.com/launchdarkly/go-sdk-common/v3/ldvalue"
	ld "github.com/launchdarkly/go-server-sdk/v7"
	"github.com/launchdarkly/go-server-sdk/v7/ldcomponents"
)

var ldClient *ld.LDClient

const (
	defaultBannerText   = "Welcome back!"
	defaultTierLabel    = "Standard"
	defaultMaxCartItems = 5
)

type featureFlags struct {
	GPTEnabled        bool
	CreditCardEnabled bool
	PaypalEnabled     bool
	ApplePayEnabled   bool
	BannerText        string
	TierLabel         string
	MaxCartItems      int
	RecommendedPlan   ldvalue.Value
}

func getFlagValues(user ldcontext.Context) featureFlags {
	return featureFlags{
		GPTEnabled:        boolVariation("enable-gpt-511-codex-max", user, true),
		CreditCardEnabled: boolVariation("enable-credit-card", user, true),
		PaypalEnabled:     boolVariation("enable-paypal", user, true),
		ApplePayEnabled:   boolVariation("enable-apple-pay", user, false),
		BannerText:        stringVariation("checkout-banner-text", user, defaultBannerText),
		TierLabel:         stringVariation("user-tier-label", user, defaultTierLabel),
		MaxCartItems:      intVariation("max-items-in-cart", user, defaultMaxCartItems),
		RecommendedPlan:   jsonVariation("recommended-plan", user, defaultPlanValue()),
	}
}

func boolVariation(name string, user ldcontext.Context, defaultValue bool) bool {
	if ldClient == nil {
		return defaultValue
	}

	value, err := ldClient.BoolVariation(name, user, defaultValue)
	if err != nil {
		return defaultValue
	}

	return value
}

func stringVariation(name string, user ldcontext.Context, defaultValue string) string {
	if ldClient == nil {
		return defaultValue
	}

	value, err := ldClient.StringVariation(name, user, defaultValue)
	if err != nil {
		return defaultValue
	}

	return value
}

func intVariation(name string, user ldcontext.Context, defaultValue int) int {
	if ldClient == nil {
		return defaultValue
	}

	value, err := ldClient.IntVariation(name, user, defaultValue)
	if err != nil {
		return defaultValue
	}

	return value
}

func jsonVariation(name string, user ldcontext.Context, defaultValue ldvalue.Value) ldvalue.Value {
	if ldClient == nil {
		return defaultValue
	}

	value, err := ldClient.JSONVariation(name, user, defaultValue)
	if err != nil {
		return defaultValue
	}

	return value
}

func defaultPlanValue() ldvalue.Value {
	return ldvalue.ObjectBuild().
		Set("name", ldvalue.String("Starter")).
		Set("price", ldvalue.Int(9)).
		Set("perk", ldvalue.String("Basic coverage + chat support")).
		Build()
}

func formatPlan(plan ldvalue.Value) string {
	name := plan.GetByKey("name").StringValue()
	price := plan.GetByKey("price").IntValue()
	perk := plan.GetByKey("perk").StringValue()

	if name == "" {
		name = "Starter"
	}

	if perk == "" {
		perk = "No perks configured"
	}

	return fmt.Sprintf("%s plan • $%d/mo • %s", name, price, perk)
}

func main() {
	sdkKey := os.Getenv("LAUNCHDARKLY_SDK_KEY")
	if sdkKey == "" {
		log.Println("LAUNCHDARKLY_SDK_KEY not set. Running with default flag values only.")
	}

	// Initialize LaunchDarkly client
	config := ld.Config{
		Events: ldcomponents.SendEvents(),
	}

	var err error
	if sdkKey != "" {
		ldClient, err = ld.MakeCustomClient(sdkKey, config, 5*time.Second)
		if err != nil {
			log.Printf("failed to initialize LaunchDarkly SDK, falling back to defaults: %v", err)
			ldClient = nil
		} else {
			defer ldClient.Close()
		}
	}

	// Wait for the client to initialize
	if ldClient != nil && !ldClient.Initialized() {
		log.Println("Warning: LaunchDarkly client did not initialize. Using default values.")
	}

	// Set up HTTP server
	http.HandleFunc("/", handleHome)

	fmt.Println("Server starting on http://localhost:8080")
	fmt.Println("Visit http://localhost:8081 in your browser to test the feature flags")

	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Get user ID from query parameter, default to "anonymous-user"
	userID := r.URL.Query().Get("user")
	if userID == "" {
		userID = "anonymous-user"
	}

	// Create a user context for flag evaluation
	user := ldcontext.New(userID)

	flags := getFlagValues(user)

	_ = ctx // unused for now

	gptStatus := "DISABLED"
	gptColor := "red"
	if flags.GPTEnabled {
		gptStatus = "ENABLED"
		gptColor = "green"
	}

	fmt.Fprintf(w, `<html><head><style>
		body { font-family: sans-serif; padding: 40px; background: #f5f5f5; }
		.container { max-width: 800px; margin: 0 auto; }
		.status { padding: 20px; background: white; border-radius: 8px; margin-bottom: 30px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
		h1 { margin-top: 0; }
		h2 { color: #333; margin-bottom: 20px; }
		.user-info { background: #e3f2fd; padding: 10px 15px; border-radius: 4px; margin-bottom: 20px; }
		.payment-methods { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 20px; }
		.payment-box { 
			padding: 30px; 
			background: white; 
			border-radius: 8px; 
			text-align: center; 
			box-shadow: 0 2px 4px rgba(0,0,0,0.1);
			border: 2px solid #4CAF50;
		}
		.payment-box h3 { margin: 0 0 10px 0; color: #333; }
		.payment-box p { margin: 0; color: #666; font-size: 14px; }
		.refresh { display: inline-block; margin-top: 20px; padding: 10px 20px; background: #2196F3; color: white; text-decoration: none; border-radius: 4px; }
		.refresh:hover { background: #0b7dda; }
		.provider { color: #999; font-size: 12px; margin-top: 10px; }
		.flag-list { list-style: none; padding: 0; margin: 0; }
		.flag-list li { margin: 6px 0; color: #333; }
		.flag-name { color: #777; font-size: 12px; margin-left: 6px; }
	</style></head><body>
	<div class="container">
		<div class="user-info">
			👤 Current User: <strong>%s</strong> 
			<span style="color: #666; margin-left: 10px;">Try: <a href="?user=alice">alice</a> | <a href="?user=bob">bob</a> | <a href="/">default</a></span>
		</div>
		<div class="status">
			<h1 style="color: %s;">%s GPT-5.1-Codex-Max is %s</h1>
			<p>Flag: <code>enable-gpt-511-codex-max</code></p>
			<p class="provider">Powered by LaunchDarkly</p>
		</div>
		
		<h2>Available Payment Methods</h2>
		<div class="payment-methods">`, userID, gptColor,
		map[bool]string{true: "✓", false: "✗"}[flags.GPTEnabled],
		gptStatus)

	if flags.CreditCardEnabled {
		fmt.Fprintf(w, `
			<div class="payment-box">
				<h3>💳 Credit Card</h3>
				<p>Visa, Mastercard, Amex</p>
			</div>`)
	}

	if flags.PaypalEnabled {
		fmt.Fprintf(w, `
			<div class="payment-box">
				<h3>🅿️ PayPal</h3>
				<p>Fast & secure checkout</p>
			</div>`)
	}

	if flags.ApplePayEnabled {
		fmt.Fprintf(w, `
			<div class="payment-box">
				<h3>🍎 Apple Pay</h3>
				<p>One-touch payment</p>
			</div>`)
	}

	fmt.Fprintf(w, `
		</div>
		<div class="status">
			<h2>Personalized Checkout Experience</h2>
			<ul class="flag-list">
				<li><strong>Banner message:</strong> %s <span class="flag-name">(flag: checkout-banner-text)</span></li>
				<li><strong>User tier:</strong> %s <span class="flag-name">(flag: user-tier-label)</span></li>
				<li><strong>Max items in cart:</strong> %d <span class="flag-name">(flag: max-items-in-cart)</span></li>
				<li><strong>Recommended plan:</strong> %s <span class="flag-name">(flag: recommended-plan)</span></li>
			</ul>
		</div>
		<a href="/" class="refresh">Refresh Page</a>
	</div>
	</body></html>`, flags.BannerText, flags.TierLabel, flags.MaxCartItems, formatPlan(flags.RecommendedPlan))
}
