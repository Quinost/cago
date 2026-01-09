package http_s

type HealthResponse struct {
	Status string `json:"status"`
	Server string `json:"server"`
}

type ErrorResponse struct {
	Error  string `json:"error"`
	Status int    `json:"status"`
}

type KeysListResponse struct {
	Keys    []string `json:"keys"`
	Count   int      `json:"count"`
	Pattern string   `json:"pattern"`
}

type GetResponse struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Ttl   int64  `json:"ttl"`
}

type SetRequest struct {
	Value string `jsoin:"value"`
	TTL   int64  `json:"ttl,omitempty"`
}

type SetResponse struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	TTL     int64  `json:"ttl"`
	Success bool   `json:"success"`
}

type DeleteResponse struct {
	Key     string `json:"key"`
	Deleted bool   `json:"deleted"`
}

type ExpireRequest struct {
	TTL int64 `json:"ttl"`
}

type ExpireResponse struct {
	Key     string `json:"key"`
	TTL     int64  `json:"ttl"`
	Success bool   `json:"success"`
}

type StatsResponse struct {
	TotalKeys       int     `json:"total_keys"`
	DefaultTTL      float64 `json:"default_ttl_seconds"`
	CleanupInterval float64 `json:"cleanup_interval_seconds"`
}
