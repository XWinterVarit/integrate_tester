package model

type ExecuteQueryRequest struct {
	Query   string            `json:"query"`
	Args    map[string]string `json:"args"`
	Limit   int               `json:"limit"`
	Offset  int               `json:"offset"`
	Sort    string            `json:"sort"`
	SortDir string            `json:"sort_dir"`
}

type UpdateCellRequest struct {
	Column string `json:"column"`
	Value  string `json:"value"`
	Rowid  string `json:"rowid"`
}

type ResolvePresetQueryRequest struct {
	Args map[string]string `json:"args"`
}

type DeleteRowRequest struct {
	Rowid string `json:"rowid"`
}

type InsertRowRequest struct {
	Columns []string `json:"columns"`
	Values  []string `json:"values"`
}

type RecentTouchRequest struct {
	Key string `json:"key"`
}

type SaveClientRequest struct {
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name"`
	Host        string   `json:"host"`
	Port        int      `json:"port"`
	ServiceName string   `json:"service_name"`
	Username    string   `json:"username"`
	Password    string   `json:"password"`
	Tables      []string `json:"tables"`
}

type ReorderClientsRequest struct {
	Names []string `json:"names"`
}

type TestConnectionRequest struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	ServiceName string `json:"service_name"`
	Username    string `json:"username"`
	Password    string `json:"password"`
}

type AcquireLockRequest struct {
	ResourceType string `json:"resource_type"`
	ResourceID   string `json:"resource_id"`
	ScopeClient  string `json:"scope_client"`
	SessionID    string `json:"session_id"`
}

type RenewLockRequest struct {
	SessionID string `json:"session_id"`
}

type ReleaseLockRequest struct {
	SessionID string `json:"session_id"`
}

type PresetQueryArgConfig struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

type SavePresetFilterRequest struct {
	Name    string   `json:"name"`
	Details string   `json:"details"`
	Columns []string `json:"columns"`
}

type SavePresetQueryRequest struct {
	Name      string                 `json:"name"`
	Query     string                 `json:"query"`
	Arguments []PresetQueryArgConfig `json:"arguments"`
}

type ValidateQueryRequest struct {
	Query     string                 `json:"query"`
	Arguments []PresetQueryArgConfig `json:"arguments"`
}

type RowQueryParams struct {
	Select  string
	Where   string
	Sort    string
	SortDir string
	Limit   int
	Offset  int
}
