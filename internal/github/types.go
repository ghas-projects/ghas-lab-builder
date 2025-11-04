package api

// Enterprise represents the enterprise information returned from GitHub GraphQL API
type Enterprise struct {
	ID           string `json:"id"`
	BillingEmail string `json:"billingEmail"`
	Slug         string `json:"slug"`
}

type Organization struct {
	ID    string `json:"id"`
	Login string `json:"login"`
	Name  string `json:"name"`
}

type Repository struct {
	ID       int64  `json:"id"`
	FullName string `json:"full_name"`
	HTMLURL  string `json:"html_url"`
}
