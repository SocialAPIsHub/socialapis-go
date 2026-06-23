package socialapis

// Response is the generic untyped response a method returns when the
// endpoint doesn't have a hand-typed struct yet. Most endpoints in
// v0.1.0 return Response — only the three headline endpoints return
// typed structs (PageInfo, GroupInfo, ProfileInfo).
//
// Forward-compat: any field the API returns is preserved in the map
// regardless of whether we've typed it; callers can access via
// `r["some_field"]`.
type Response map[string]any

// PageInfo is the response from `Facebook.GetPageInfo()` —
// `GET /facebook/pages/details`.
//
// The API wraps this payload under string key "0" in the response
// envelope; the SDK unwraps before returning. Field names match the
// LIVE API exactly (verified 2026-06-22).
//
// Forward-compat: any field the API adds that we haven't typed yet
// lives on Extra — access via `page.Extra["new_field"]`.
type PageInfo struct {
	// Identifiers
	AdPageID string `json:"ad_page_id,omitempty"`
	UserID   string `json:"user_id,omitempty"`

	// Display
	Title    string `json:"title,omitempty"`
	URL      string `json:"url,omitempty"`
	Category any    `json:"category,omitempty"` // can be a list or a string
	Status   string `json:"status,omitempty"`

	// Content
	Bio         string `json:"bio,omitempty"`
	Description string `json:"description,omitempty"`

	// Contact
	Address     string `json:"address,omitempty"`
	Phone       string `json:"phone,omitempty"`
	Email       string `json:"email,omitempty"`
	Website     string `json:"website,omitempty"`
	MapsAddress string `json:"maps_address,omitempty"`

	// Engagement
	FollowersCount   int    `json:"followers_count,omitempty"`
	FollowersDisplay string `json:"followers_display,omitempty"`
	LikesCount       int    `json:"likes_count,omitempty"`
	LikesDisplay     string `json:"likes_display,omitempty"`

	// Media
	Image    string `json:"image,omitempty"`
	ImageAlt string `json:"image_alt,omitempty"`

	// Ratings
	Rating        string `json:"rating,omitempty"`
	RatingCount   int    `json:"rating_count,omitempty"`
	RatingOverall string `json:"rating_overall,omitempty"`

	// Business
	BusinessHours        string `json:"business_hours,omitempty"`
	BusinessPrice        string `json:"business_price,omitempty"`
	BusinessServices     string `json:"business_services,omitempty"`
	IsBusinessPageActive bool   `json:"is_business_page_active,omitempty"`
	ConfirmedOwnerLabel  string `json:"confirmed_owner_label,omitempty"`

	// Linked socials
	Twitter   string `json:"twitter,omitempty"`
	Instagram string `json:"instagram,omitempty"`
	LinkedIn  string `json:"linkedin,omitempty"`
	Pinterest string `json:"pinterest,omitempty"`
	Telegram  string `json:"telegram,omitempty"`
	YouTube   string `json:"youtube,omitempty"`

	// Forward-compat — any field the API adds that we haven't typed yet
	Extra map[string]any `json:"-"`
}

// GroupInfo is the response from `Facebook.GetGroupDetails()` —
// `GET /facebook/groups/details`.
//
// NOTE: this endpoint has NO envelope wrapper — payload sits at the
// top level alongside `message` and `meta`.
type GroupInfo struct {
	GroupID                     string         `json:"group_id,omitempty"`
	GroupMemberCount            string         `json:"group_member_count,omitempty"`
	GroupTotalMembersInfoText   string         `json:"group_total_members_info_text,omitempty"`
	GroupNewMembersInfoText     string         `json:"group_new_members_info_text,omitempty"`
	DescriptionText             string         `json:"description_text,omitempty"`
	PrivacyInfoText             map[string]any `json:"privacy_info_text,omitempty"`
	CreatedTime                 int64          `json:"created_time,omitempty"`
	GroupRules                  []any          `json:"group_rules,omitempty"`
	GroupHistory                map[string]any `json:"group_history,omitempty"`
	AdminTags                   []any          `json:"admin_tags,omitempty"`
	GroupLocations              []any          `json:"group_locations,omitempty"`
	NumberOfPostsInLastDay      int            `json:"number_of_posts_in_last_day,omitempty"`
	NumberOfPostsInLastMonth    int            `json:"number_of_posts_in_last_month,omitempty"`

	// Forward-compat
	Extra map[string]any `json:"-"`
}

// ProfileInfo is the response from `Instagram.GetProfileDetails()` —
// `GET /instagram/profile/details`.
//
// The API wraps the payload under "data" in the envelope (alongside
// `success`, `message`, `meta`); the SDK unwraps before returning.
type ProfileInfo struct {
	// Identifiers
	ID   string `json:"id,omitempty"`
	PK   string `json:"pk,omitempty"`
	FBID string `json:"fbid,omitempty"`

	// Display
	Username     string `json:"username,omitempty"`
	FullName     string `json:"full_name,omitempty"`
	Biography    string `json:"biography,omitempty"`
	CategoryName string `json:"category_name,omitempty"`

	// Media URLs
	ProfilePicURL          string `json:"profile_pic_url,omitempty"`
	ProfilePicURLHD        string `json:"profile_pic_url_hd,omitempty"`
	ExternalURL            string `json:"external_url,omitempty"`
	ExternalURLLinkshimmed string `json:"external_url_linkshimmed,omitempty"`

	// Counts
	FollowersCount       int `json:"followers_count,omitempty"`
	FollowingCount       int `json:"following_count,omitempty"`
	MediaCount           int `json:"media_count,omitempty"`
	TotalClipsCount      int `json:"total_clips_count,omitempty"`
	HighlightReelCount   int `json:"highlight_reel_count,omitempty"`
	MutualFollowersCount int `json:"mutual_followers_count,omitempty"`

	// Flags
	IsPrivate             bool `json:"is_private,omitempty"`
	IsVerified            bool `json:"is_verified,omitempty"`
	IsBusinessAccount     bool `json:"is_business_account,omitempty"`
	IsProfessionalAccount bool `json:"is_professional_account,omitempty"`
	IsMemorialized        bool `json:"is_memorialized,omitempty"`
	IsUnpublished         bool `json:"is_unpublished,omitempty"`
	IsEmbedsDisabled      bool `json:"is_embeds_disabled,omitempty"`
	IsJoinedRecently      bool `json:"is_joined_recently,omitempty"`
	AccountType           int  `json:"account_type,omitempty"`

	// Features
	HasClips     bool `json:"has_clips,omitempty"`
	HasGuides    bool `json:"has_guides,omitempty"`
	HasChannel   bool `json:"has_channel,omitempty"`
	HasArEffects bool `json:"has_ar_effects,omitempty"`

	// Business contact
	BusinessCategoryName  string `json:"business_category_name,omitempty"`
	BusinessEmail         string `json:"business_email,omitempty"`
	BusinessPhoneNumber   string `json:"business_phone_number,omitempty"`
	BusinessContactMethod string `json:"business_contact_method,omitempty"`
	AddressStreet         string `json:"address_street,omitempty"`
	CityName              string `json:"city_name,omitempty"`
	Zip                   string `json:"zip,omitempty"`

	// Forward-compat
	Extra map[string]any `json:"-"`
}
