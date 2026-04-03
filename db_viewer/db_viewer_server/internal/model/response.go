package model

type ClientInfo struct {
	Name   string `json:"name"`
	Schema string `json:"schema"`
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
