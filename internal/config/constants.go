package config

type AppContextKey string

const (
	AppIDKey          AppContextKey = "app_id"
	PrivateKeyKey     AppContextKey = "private_key" // Changed from PrivateKeyPathKey
	TokenKey          AppContextKey = "tokens"
	BaseURLKey        AppContextKey = "base_url"
	DryRunKey         AppContextKey = "dry_run"
	LabDateKey        AppContextKey = "lab_date"
	FacilitatorsKey   AppContextKey = "facilitators"
	EnterpriseSlugKey AppContextKey = "enterprise_slug"
	DefaultBaseURL    string        = "https://api.github.com"
	OrgKey            AppContextKey = "org"
	EnterpriseType    string        = "Enterprise"
	OrganizationType  string        = "Organization"
)
