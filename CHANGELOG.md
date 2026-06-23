# Changelog

All notable changes to this project will be documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.1] — Unreleased

Initial Go release. Full coverage of the SocialAPIs.io public REST surface in Go — mirrors the Python (`socialapis-sdk` on PyPI) and JavaScript (`socialapis-sdk` on npm) SDKs.

Starts at 0.1.1 to lockstep the version number across all three language SDKs in this family — no v0.1.0 was ever released for Go.

### Added — Facebook client (35 methods)

**Pages**: `GetPageID`, `GetPageInfo` → `*PageInfo`, `GetPagePosts`, `GetPageReels`, `GetPageVideos`

**Groups**: `GetGroupID`, `GetGroupDetails` → `*GroupInfo`, `GetGroupMetadata`, `GetGroupPosts`, `GetGroupVideos`

**Posts**: `GetPostID`, `GetPostDetails`, `GetPostDetailsExtended`, `GetPostComments`, `GetCommentReplies`, `GetPostAttachments`, `GetVideoPostDetails`

**Search**: `SearchPages`, `SearchPeople`, `SearchLocations`, `SearchPosts`, `SearchVideos`

**Meta Ads Library**: `GetAdsCountries`, `SearchAds`, `GetAdsPageDetails`, `GetAdArchiveDetails`, `SearchAdsByKeywords`

**Marketplace**: `SearchMarketplace`, `GetListingDetails`, `GetSellerDetails`, `GetMarketplaceCategories`, `GetCityCoordinates`, `SearchVehicles`, `SearchRentals`

**Media**: `DownloadMedia`

### Added — Instagram client (13 methods)

**Profiles**: `GetUserID`, `GetProfileDetails` → `*ProfileInfo`, `GetProfilePosts`, `GetProfileReels`, `GetProfileHighlights`, `GetHighlightDetails`

**Posts**: `GetPostID`, `GetPostDetails`

**Reels**: `GetReelsFeed`, `GetReelsByAudio`

**Search + Locations**: `Search`, `GetLocationPosts`, `GetNearbyLocations`

### Added — Account client (3 methods)

`GetUsage`, `GetTopUps`, `GetLimits` — all free, don't consume credits.

### Added — Infrastructure

- Typed error hierarchy: `APIError`, `AuthenticationError`, `InsufficientCreditsError`, `RateLimitError`, `BadRequestError`, `APIServerError`, `ConnectionError`. Dispatch with `errors.As`.
- Typed response structs with **real API field names** verified against the live API (2026-06-22): `PageInfo`, `GroupInfo`, `ProfileInfo`
- Migration aliases `FacebookScraper`, `InstagramScraper` as **type aliases** (`type X = Y`) — identical type, identical method set
- Identifier normalisation — pass a slug or a full URL; SDK coerces to API-expected form
- `extra map[string]any` pass-through on every method — forward-compatible when API adds filters
- No `limit` parameter — cursor-based pagination via response body
- Per-endpoint envelope handling (FB pages `"0"`, IG profiles `"data"`, FB groups no wrapper)
- Idiomatic Go conventions: `ctx context.Context` first param, functional options (`WithBaseURL`, `WithHTTPClient`), no external dependencies

### Added — Tooling

- Module path: `github.com/SocialAPIsHub/socialapis-go`
- Minimum Go: 1.22
- Zero external dependencies (stdlib `net/http`, `encoding/json` only)
- Tests use `httptest` — no live API calls in CI
- CI: `go vet`, `gofmt`, `go test -race` on Go 1.22 and 1.23
- Release: GitHub Release created on `v*.*.*` tag; module is then installable via `go get github.com/SocialAPIsHub/socialapis-go@vX.Y.Z`
