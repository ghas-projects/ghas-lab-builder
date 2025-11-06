package config

type AppContextKey string

const (
	AppIDKey          AppContextKey = "app_id"
	PrivateKeyPathKey AppContextKey = "private_key_path"
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
