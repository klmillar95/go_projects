package querytool

// data source (includes inbound and outbound sources)
// ---------------------------------------------------
// Name: This is the unique identifier for this inbound/outbound data source
// Type: This is the specific type of data source (different from the "type" found in the .conf json. These are not located in the same Source struct because the feeder needs to know what type of Config object to unmarshal the json into
// ConnStr: This is the connection string or main file directory for the data source 
// RateLim: This is the rate limit integer value for the data source
// RateType: This is the rate limit type for the data source, i.e. rows, MB
type Source struct {
	Name string	`json:"name"`
	Type string	`json:"type"`
	ConnStr string	`json:"connection_string"`
	RateLim	int	`json:"ratelimit"`
	RateType string	`json:"ratelimittype"`
}

// query
// -----
// PK: Primary Key, basically which column the primary key is (either a name, or a string integer for the column number)
// Cron: A 5 element cron string. Accepts dashes and commas as well as integers and *.
// Cols: The columns to pull (string or integer string) as the key. Values are blank if the column names will be retained, otherwise columns will be renamed. Useful especially for csv where the column names might not exist.
// Table: This is a map because a non blank value will be what the table will be renamed to in the destination (i.e. BugEvent -> bug_event)
type Query struct {
	PK string		`json:"primary_key"`
	Cron string		`json:"cron"`
	Cols map[string]string	`json:"columns"`
	Table map[string]string	`json:"table"`
}

// relational database configuration file
// --------------------------------------
// Source: This is the unique identifier of the source the configuration object pulls data from
// Destination: This is the unique identifier of the destination the configuration object sends data to
// Queries: A slice of query objects pertaining to this unique Source -> Destination pair
type RDBConfig struct {
	Source string		`json:"inbound_source"`
	Destination string	`json:"outbound_source"`
	Queries []Query		`json:"queries"`
}

// distributed table
// -----------------
// *NOTE: This is necessary for data sources such a .csv and .json with multiple files, since we are essentially viewing them as distributed components of a table
// Name: The name of the table that this group of files "belongs" to
// Directory: The directory under which all files are considered part of this "table"
// Schema: The first item in the map is the first column in the file, etc. This allows for each column to be assigned a name and a type
// UseFirst: Whether or not the first row in each file in the directory should be used (useful when ignoring headers)
type DistTable struct {
	Name string		 `json:"table_name"`
	Directory string	 `json:"files_directory"`
	Schema map[string]string `json:"schema"`
	UseFirst bool		 `json:"use_first_row"`
}

// distributed files configuration file
// ------------------------------------
// Similar to the relational database configuration object, except also contains a slice of DistTable objects
type DFSConfig struct {
	Source string		`json:"inbound_source"`
	Destination string	`json:"outbound_source"`
	Queries []Query		`json:"queries"`
	Tables []DistTable	`json:"tables"`
}

// job object
// ----------
// Source: Source object for the source of the data
// Destination: Source object for the destination of the extracted data
// Queries: Slice of all queries belonging to this job
// Tables: If the inbound data source is one of the distributed files type, i.e. .csv, .json etc., then there will be a slice of DistTables objects. Otherwise, this is nil
// IsDist: Boolean indicating if this uses the Tables slice.
type Job struct {
	Source Source		`json:"inbound_source"`
	Destination Source	`json:"outbound_source"`
	Queries []Query		`json:"queries"`
	Tables []DistTable	`json:"tables"`
	IsDist bool		`json:"is_dist"`
}
