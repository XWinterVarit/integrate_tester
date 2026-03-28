package model

type ExecuteQueryRequest struct {
	Query  string            `json:"query"`
	Args   map[string]string `json:"args"`
	Limit  int               `json:"limit"`
	Offset int               `json:"offset"`
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

type RowQueryParams struct {
	Select  string
	Sort    string
	SortDir string
	Limit   int
	Offset  int
}
