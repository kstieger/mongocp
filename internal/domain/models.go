package domain

// Database represents a MongoDB database to be copied.
type Database struct {
	Name string
}

// Collection represents a MongoDB collection to be copied.
type Collection struct {
	Database string
	Name     string
	Count    int64 // document count, for progress
}

// CopyTask represents a single collection copy job.
type CopyTask struct {
	SrcDB    string
	DstDB    string
	Coll     string
	DocCount int64
}
