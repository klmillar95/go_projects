package querytool

//data source (includes inbound and outbound sources)
type Source struct {
	Name string	`json:"name"`
	Type string	`json:"type"`
	ConnStr string	`json:"connection_string"`
	RateLim	int	`json:"ratelimit"`
	RateType string	`json:"ratelimittype"`
}

//query
type Query struct {
	PK string		`json:"primary_key"`
	Cron string		`json:"cron"`
	Cols map[string]string	`json:"columns"`
	Table map[string]string	`json:"table"`
}

//relational database configuration file
type RDBConfig struct {
	Source string		`json:"inbound_source"`
	Destination string	`json:"outbound_source"`
	Queries []Query		`json:"queries"`
}

//distributed table
type DistTable struct {
	Name string		 `json:"table_name"`
	Directory string	 `json:"files_directory"`
	Schema map[string]string `json:"schema"`
	UseFirst bool		 `json:"use_first_row"`
}

//distributed files configuration file
type DFSConfig struct {
	Source string		`json:"inbound_source"`
	Destination string	`json:"outbound_source"`
	Queries []Query		`json:"queries"`
	Tables []DistTable	`json:"tables"`
}

//job object
type Job struct {
	Source Source		`json:"inbound_source"`
	Destination Source	`json:"outbound_source"`
	Queries []Query		`json:"queries"`
	Tables []DistTable	`json:"tables"`
}
