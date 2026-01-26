# LaunchDarkly Feature Flags Go Example

Small Go console app that uses LaunchDarkly to manage feature flags including:
- `enable-gpt-511-codex-max` - Enable GPT-5.1-Codex-Max feature
- `enable-credit-card` - Show Credit Card payment option
- `enable-paypal` - Show PayPal payment option
- `enable-apple-pay` - Show Apple Pay payment option

## Prerequisites
- Go 1.21+
- LaunchDarkly SDK key (server-side) set as `LAUNCHDARKLY_SDK_KEY`

## Setup

1. Sign up for a free LaunchDarkly account at https://launchdarkly.com
2. Create a new project and get your SDK key from the Account Settings
3. Create the following feature flags in your LaunchDarkly dashboard:
   - `enable-gpt-511-codex-max` (Boolean flag)
   - `enable-credit-card` (Boolean flag)
   - `enable-paypal` (Boolean flag)
   - `enable-apple-pay` (Boolean flag)

## Run
```sh
export LAUNCHDARKLY_SDK_KEY="your_sdk_key_here"
go run .
```

Visit http://localhost:8080 in your browser to see the feature flags in action.

## Test
```sh
go test -v
```

## How it works

The app:
1. Connects to LaunchDarkly using the SDK key
2. Starts a web server on port 8080
3. On each request, evaluates the feature flags for an anonymous user
4. Displays enabled features and payment methods based on flag values

Toggle flags in the LaunchDarkly dashboard and refresh the page to see changes instantly!
