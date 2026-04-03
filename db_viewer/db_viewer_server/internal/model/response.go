package model

type ClientInfo struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Schema      string `json:"schema"`
}

type ClientConfigResponse struct {
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name"`
	Host        string   `json:"host"`
	Port        int      `json:"port"`
	ServiceName string   `json:"service_name"`
	Username    string   `json:"username"`
	Tables      []string `json:"tables"`
}

type LockResponse struct {
	Key          string `json:"key"`
	LockedBy     string `json:"locked_by"`
	LockedAt     string `json:"locked_at"`
	ExpiresAt    string `json:"expires_at"`
	ResourceType string `json:"resource_type"`
	ResourceID   string `json:"resource_id"`
}

type TestConnectionResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type ListTablesFromDBResponse struct {
	Tables []string `json:"tables"`
}

type PresetFilterResponse struct {
	Name    string   `json:"name"`
	Details string   `json:"details"`
	Columns []string `json:"columns"`
}

type PresetQueryArgResponse struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

type PresetQueryResponse struct {
	Name      string                   `json:"name"`
	Query     string                   `json:"query"`
	Arguments []PresetQueryArgResponse `json:"arguments"`
}

type ValidateQueryResponse struct {
	Valid         bool     `json:"valid"`
	Error         string   `json:"error,omitempty"`
	UndefinedArgs []string `json:"undefined_args,omitempty"`
}

type ResolvedQueryResponse struct {
	ResolvedQuery string `json:"resolved_query"`
}

type StatusResponse struct {
	Status string `json:"status"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
