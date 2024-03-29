package objects

type Project struct {
	Id             string    `json:"id"`
	OrganizationId string    `json:"organization_id"`
	Name           string    `json:"name"`
	Region         string    `json:"region"`
	CreatedAt      string    `json:"created_at"`
	Database       *Database `json:"database,omitempty"`
}

type Database struct {
	Host    string `json:"host"`
	Version string `json:"version"`
}
