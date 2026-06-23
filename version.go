// Package socialapis is the Go SDK for the SocialAPIs.io REST API.
// It mirrors the Python `socialapis-sdk` and TypeScript `socialapis-sdk`
// — same endpoint coverage, idiomatic per-language conventions.
package socialapis

// Version of the SDK. Bumped by the release workflow on `git tag vX.Y.Z`.
//
// Lockstep with the Python (`socialapis-sdk` on PyPI) and JavaScript
// (`socialapis-sdk` on npm) SDKs in this family. All three start at
// 0.1.1 so users know the SDKs are at feature parity.
const Version = "0.1.1"

// userAgent is the User-Agent header value sent on every request.
const userAgent = "socialapis-go/" + Version
