package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/launchdarkly/go-sdk-common/v3/ldcontext"
	ld "github.com/launchdarkly/go-server-sdk/v7"
	"github.com/launchdarkly/go-server-sdk/v7/ldcomponents"
)

var ldClient *ld.LDClient

func main() {
	sdkKey := os.Getenv("LAUNCHDARKLY_SDK_KEY")
	if sdkKey == "" {
		log.Fatal("set LAUNCHDARKLY_SDK_KEY to your SDK key")
	}

	// Initialize LaunchDarkly client
	config := ld.Config{
		Events: ldcomponents.SendEvents(),
	}

	var err error
	ldClient, err = ld.MakeCustomClient(sdkKey, config, 5*time.Second)
	if err != nil {
		log.Fatalf("failed to initialize LaunchDarkly SDK: %v", err)
	}
	defer ldClient.Close()

	// Wait for the client to initialize
	if !ldClient.Initialized() {
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

	// Check feature flags
	gptEnabled, _ := ldClient.BoolVariation("enable-gpt-511-codex-max", user, true)
	creditCardEnabled, _ := ldClient.BoolVariation("enable-credit-card", user, true)
	paypalEnabled, _ := ldClient.BoolVariation("enable-paypal", user, true)
	applePayEnabled, _ := ldClient.BoolVariation("enable-apple-pay", user, false)

	_ = ctx // unused for now

	gptStatus := "DISABLED"
	gptColor := "red"
	if gptEnabled {
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
		map[bool]string{true: "✓", false: "✗"}[gptEnabled],
		gptStatus)

	if creditCardEnabled {
		fmt.Fprintf(w, `
			<div class="payment-box">
				<h3>💳 Credit Card</h3>
				<p>Visa, Mastercard, Amex</p>
			</div>`)
	}

	if paypalEnabled {
		fmt.Fprintf(w, `
			<div class="payment-box">
				<h3>🅿️ PayPal</h3>
				<p>Fast & secure checkout</p>
			</div>`)
	}

	if applePayEnabled {
		fmt.Fprintf(w, `
			<div class="payment-box">
				<h3>🍎 Apple Pay</h3>
				<p>One-touch payment</p>
			</div>`)
	}

	fmt.Fprintf(w, `
		</div>
		<a href="/" class="refresh">Refresh Page</a>
	</div>
	</body></html>`)
}
