package handlers

import "regexp"

var nonIDChars = regexp.MustCompile(`[^a-zA-Z0-9_]+`)
var slaPattern = regexp.MustCompile(`(?i)(\d+(?:[.,]\d+)?)\s*(menit|minute|minutes|jam|hour|hours|hari|day|days|minggu|week|weeks)`)

const (
	graphConflictCode = "GRAPH_VERSION_CONFLICT"
	graphInvalidCode  = "INVALID_GRAPH_REFERENCE"
)

type graphSyncError struct {
	Code    string
	Message string
}

func (e *graphSyncError) Error() string { return e.Message }
