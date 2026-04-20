// Hand-written proto-equivalent types until proto compilation is set up.
// Replace with generated code once proto/platform/config.proto is compiled.
package proto

// ConfigEntry is the wire type for a configuration key-value pair.
type ConfigEntry struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	Version int64  `json:"version"`
}

// GetRequest is sent by callers to retrieve a config value.
type GetRequest struct {
	Key string
}

// GetResponse carries the config entry back to the caller.
type GetResponse struct {
	Entry *ConfigEntry
	Found bool
}

// SetRequest creates or updates a configuration key.
type SetRequest struct {
	Key   string
	Value string
}

// DeleteRequest removes a configuration key.
type DeleteRequest struct {
	Key string
}

// WatchRequest opens a server-side stream for a key prefix.
type WatchRequest struct {
	Prefix string
}

// WatchEvent is streamed to the caller when a watched key changes.
type WatchEvent struct {
	Entry  *ConfigEntry
	Deleted bool
}

// ListRequest retrieves all keys under a given prefix.
type ListRequest struct {
	Prefix string
}

// ListResponse returns all matching config entries.
type ListResponse struct {
	Entries []*ConfigEntry
}
