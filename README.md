# socialapis-go — Go SDK for Facebook + Instagram public data

[![Go Reference](https://pkg.go.dev/badge/github.com/SocialAPIsHub/socialapis-go.svg)](https://pkg.go.dev/github.com/SocialAPIsHub/socialapis-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/SocialAPIsHub/socialapis-go)](https://goreportcard.com/report/github.com/SocialAPIsHub/socialapis-go)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Go](https://img.shields.io/badge/go-%3E%3D1.22-blue.svg)](https://go.dev)

Idiomatic Go client for the [socialapis.io](https://socialapis.io) REST API. Mirrors the [Python](https://pypi.org/project/socialapis-sdk/) and [JavaScript](https://www.npmjs.com/package/socialapis-sdk) SDKs — same 51 endpoints, same envelope handling, same migration aliases — but with Go conventions: context-aware methods, functional options, typed errors, zero external dependencies.

```bash
go get github.com/SocialAPIsHub/socialapis-go@latest
```

```go
package main

import (
    "context"
    "fmt"
    "log"

    socialapis "github.com/SocialAPIsHub/socialapis-go"
)

func main() {
    fb, err := socialapis.NewFacebook("YOUR_API_TOKEN")
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()
    page, err := fb.GetPageInfo(ctx, "EngenSA", nil)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(page.Title, page.FollowersCount, page.LikesCount)

    ig, _ := socialapis.NewInstagram("YOUR_API_TOKEN")
    profile, err := ig.GetProfileDetails(ctx, "instagram", nil)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(profile.Username, profile.FollowersCount, profile.IsVerified)
}
```

**[Get a free API token →](https://socialapis.io/auth/signup)** — 200 calls/month, no credit card

## One-line migration

If you're moving from popular abandoned scraper packages, the migration aliases keep the import path greppable:

```go
import socialapis "github.com/SocialAPIsHub/socialapis-go"

fb, _ := socialapis.NewFacebookScraper("YOUR_API_TOKEN")  // alias of Facebook
page, _ := fb.GetPageInfo(ctx, "EngenSA", nil)
```

`FacebookScraper` and `InstagramScraper` are **Go type aliases** (`type FacebookScraper = Facebook`) — identical type, identical method set, just different names. Defined in [`alias.go`](alias.go).

## Idiomatic Go conventions

- **`context.Context` first**: every endpoint method takes `ctx context.Context` as the first argument. Supports cancellation and timeouts via `context.WithTimeout(...)`.
- **Functional options**: client config via `WithBaseURL(url)`, `WithHTTPClient(client)` instead of a settings struct. Easy to extend later without breaking existing callers.
- **Typed errors via `errors.As`**: dispatch on `*AuthenticationError`, `*RateLimitError`, etc. — never string-match error messages.
- **Zero external dependencies**: stdlib only (`net/http`, `encoding/json`, `context`). No deps in `go.mod`.
- **`extra map[string]any`**: every method's last arg accepts arbitrary forward-compat query params. When the API adds a filter, you use it without an SDK release.

## What's covered (v0.1.1)

### Facebook client (35 methods)

**Pages**: `GetPageID`, `GetPageInfo` → `*PageInfo`, `GetPagePosts`, `GetPageReels`, `GetPageVideos`

**Groups**: `GetGroupID`, `GetGroupDetails` → `*GroupInfo`, `GetGroupMetadata`, `GetGroupPosts`, `GetGroupVideos`

**Posts**: `GetPostID`, `GetPostDetails`, `GetPostDetailsExtended`, `GetPostComments`, `GetCommentReplies`, `GetPostAttachments`, `GetVideoPostDetails`

**Search**: `SearchPages`, `SearchPeople`, `SearchLocations`, `SearchPosts`, `SearchVideos`

**Meta Ads Library**: `GetAdsCountries`, `SearchAds`, `GetAdsPageDetails`, `GetAdArchiveDetails`, `SearchAdsByKeywords`

**Marketplace**: `SearchMarketplace`, `GetListingDetails`, `GetSellerDetails`, `GetMarketplaceCategories`, `GetCityCoordinates`, `SearchVehicles`, `SearchRentals`

**Media**: `DownloadMedia`

### Instagram client (13 methods)

**Profiles**: `GetUserID`, `GetProfileDetails` → `*ProfileInfo`, `GetProfilePosts`, `GetProfileReels`, `GetProfileHighlights`, `GetHighlightDetails`

**Posts**: `GetPostID`, `GetPostDetails`

**Reels**: `GetReelsFeed`, `GetReelsByAudio`

**Search + Locations**: `Search`, `GetLocationPosts`, `GetNearbyLocations`

### Account client (3 methods)

Free — no credits charged.

`GetUsage`, `GetTopUps`, `GetLimits`

## Pagination — no `limit`, cursor-based

The API decides page size. Take the cursor from the response and pass it back via `extra`:

```go
fb, _ := socialapis.NewFacebook("YOUR_API_TOKEN")
ctx := context.Background()

result, _ := fb.GetPagePosts(ctx, "EngenSA", nil)
var allPosts []any
if posts, ok := result["posts"].([]any); ok {
    allPosts = append(allPosts, posts...)
}

for {
    cursor, _ := result["next_cursor"].(string)
    if cursor == "" {
        break
    }
    result, _ = fb.GetPagePosts(ctx, "EngenSA", map[string]any{"cursor": cursor})
    if posts, ok := result["posts"].([]any); ok {
        allPosts = append(allPosts, posts...)
    }
}
```

## Forward-compat via `extra`

Every method accepts `extra map[string]any` for arbitrary query params:

```go
fb.SearchAds(ctx, "fitness", map[string]any{
    "country":         "US",
    "activeStatus":    "Active",
    "some_new_filter": "x",
})
// → ?query=fitness&country=US&activeStatus=Active&some_new_filter=x
```

## Error handling

Dispatch on typed errors with `errors.As`:

```go
import (
    "errors"
    "log"
    "time"

    socialapis "github.com/SocialAPIsHub/socialapis-go"
)

page, err := fb.GetPageInfo(ctx, "EngenSA", nil)
if err != nil {
    var rateErr *socialapis.RateLimitError
    var creditsErr *socialapis.InsufficientCreditsError
    var authErr *socialapis.AuthenticationError

    switch {
    case errors.As(err, &rateErr):
        time.Sleep(time.Duration(rateErr.RetryAfterSeconds) * time.Second)
    case errors.As(err, &creditsErr):
        log.Fatal("Out of credits. Upgrade at https://socialapis.io/pricing")
    case errors.As(err, &authErr):
        log.Fatal("Bad token. Get one at https://socialapis.io/auth/signup")
    default:
        log.Fatal(err)
    }
}
```

Every typed error embeds `APIError`, which exposes `.StatusCode`, `.RequestID`, and `.Body` for debugging. The `RequestID` is what our backend logs — paste it into a support email and we can find the exact call.

## Configuration

```go
fb, err := socialapis.NewFacebook(
    "YOUR_API_TOKEN",
    socialapis.WithBaseURL("https://api.socialapis.io"),
    socialapis.WithHTTPClient(&http.Client{Timeout: 60 * time.Second}),
)
```

### Custom HTTP client

Drop in your own `http.Client` for things like:
- **Custom timeouts** beyond the default 30s
- **Retry middleware** (e.g. wrap with `github.com/hashicorp/go-retryablehttp`)
- **Tracing** via OpenTelemetry's `otelhttp.NewTransport`
- **Logging** every request/response

```go
client := &http.Client{
    Timeout:   60 * time.Second,
    Transport: otelhttp.NewTransport(http.DefaultTransport),
}
fb, _ := socialapis.NewFacebook("...", socialapis.WithHTTPClient(client))
```

## Pricing

| Tier | Calls / month | Price |
|---|---|---|
| **Free** | 200 | $0 |
| Pro | 1,500 | $4.99 |
| Ultra | 30,000 | $49 |
| Mega | 120,000 | $179 |
| Enterprise | Custom | [Contact us](https://socialapis.io/contact-us) |

One credit per successful response. Failed calls (4xx caused by bad input) don't consume credits.

## Other languages

- **Python**: [`socialapis-sdk`](https://pypi.org/project/socialapis-sdk/) on PyPI — same surface
- **TypeScript/JavaScript**: [`socialapis-sdk`](https://www.npmjs.com/package/socialapis-sdk) on npm
- **PHP**: coming soon — [notify me](https://socialapis.io/api-sources)
- Any language right now: hit the REST API directly with `curl` / `fetch`. Docs at [docs.socialapis.io](https://docs.socialapis.io).

## Support

- Docs: [docs.socialapis.io](https://docs.socialapis.io)
- API reference: [pkg.go.dev/github.com/SocialAPIsHub/socialapis-go](https://pkg.go.dev/github.com/SocialAPIsHub/socialapis-go)
- Issues: [github.com/SocialAPIsHub/socialapis-go/issues](https://github.com/SocialAPIsHub/socialapis-go/issues)
- Email: [support@socialapis.io](mailto:support@socialapis.io)
- Telegram (fastest): [t.me/socialapis](https://t.me/socialapis)

## License

MIT — see [LICENSE](LICENSE).
