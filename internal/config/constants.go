package config

type AppContextKey string

const (
	TokenKey          AppContextKey = "token"
	TokenManagerKey   AppContextKey = "token_manager"
	BaseURLKey        AppContextKey = "base_url"
	DryRunKey         AppContextKey = "dry_run"
	LabDateKey        AppContextKey = "lab_date"
	FacilitatorsKey   AppContextKey = "facilitators"
	EnterpriseSlugKey AppContextKey = "enterprise_slug"
	DefaultBaseURL    string        = "https://api.github.com"
	OrgKey            AppContextKey = "org"
)
