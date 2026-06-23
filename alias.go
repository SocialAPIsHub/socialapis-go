package socialapis

// FacebookScraper is an alias for Facebook. It exists so users
// migrating from popular abandoned scrapers can keep their import
// path / type name greppable in their codebase.
//
// FacebookScraper and Facebook are the EXACT same type — not a wrapper.
type FacebookScraper = Facebook

// NewFacebookScraper constructs a FacebookScraper (= Facebook) client.
//
// Provided as a sibling to NewFacebook so the migration story stays
// a one-line constructor swap.
func NewFacebookScraper(apiToken string, opts ...Option) (*FacebookScraper, error) {
	return NewFacebook(apiToken, opts...)
}

// InstagramScraper is an alias for Instagram. Same migration story
// as FacebookScraper.
type InstagramScraper = Instagram

// NewInstagramScraper constructs an InstagramScraper (= Instagram) client.
func NewInstagramScraper(apiToken string, opts ...Option) (*InstagramScraper, error) {
	return NewInstagram(apiToken, opts...)
}
