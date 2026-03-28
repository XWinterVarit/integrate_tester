package model

type ExecuteQueryRequest struct {
	Query string            `json:"query"`
	Args  map[string]string `json:"args"`
	Limit int               `json:"limit"`
}

type UpdateCellRequest struct {
	Column string `json:"column"`
	Value  string `json:"value"`
	Rowid  string `json:"rowid"`
}

type ResolvePresetQueryRequest struct {
	Args map[string]string `json:"args"`
}

type RecentTouchRequest struct {
	Key string `json:"key"`
}

type RowQueryParams struct {
	Select  string
	Sort    string
	SortDir string
	Limit   int
}
