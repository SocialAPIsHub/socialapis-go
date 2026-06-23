package socialapis

import (
	"context"
	"net/url"
)

// Account is the synchronous client for account-level endpoints — usage,
// credits, rate-limit info. None of these calls consume credits.
//
//	acc, err := socialapis.NewAccount("YOUR_API_TOKEN")
//	usage, err := acc.GetUsage(ctx)
type Account struct {
	*baseConfig
}

// NewAccount constructs an Account client.
func NewAccount(apiToken string, opts ...Option) (*Account, error) {
	cfg, err := newBaseConfig(apiToken, opts...)
	if err != nil {
		return nil, err
	}
	return &Account{baseConfig: cfg}, nil
}

// GetUsage returns current credit balance, usage, plan, billing period.
// Free — does not consume credits.
func (a *Account) GetUsage(ctx context.Context) (Response, error) {
	out := Response{}
	if err := a.get(ctx, "/usage", url.Values{}, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetTopUps returns auto top-up settings + recent history + lifetime spend.
func (a *Account) GetTopUps(ctx context.Context) (Response, error) {
	out := Response{}
	if err := a.get(ctx, "/usage/top-ups", url.Values{}, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetLimits returns your plan's rate limit, concurrent-task cap, allowed packages.
func (a *Account) GetLimits(ctx context.Context) (Response, error) {
	out := Response{}
	if err := a.get(ctx, "/usage/limits", url.Values{}, &out); err != nil {
		return nil, err
	}
	return out, nil
}
