package constants

const (
	DefaultStorageSchema = "storage"
	DefaultObjectTable   = "objects"
)

// RealtimeEventFilter constants for Supabase Realtime Postgres Changes.
const (
	RealtimeEventAll    = "*"
	RealtimeEventInsert = "INSERT"
	RealtimeEventUpdate = "UPDATE"
	RealtimeEventDelete = "DELETE"
)
