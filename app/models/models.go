package models

type PageResult struct {
	Page     uint        `json:"page"`
	PageSize uint        `json:"pageSize"`
	Data     interface{} `json:"data"`
}

func NewPageResult(page, pageSize uint, data []interface{}) PageResult {
	return PageResult{
		Page:     page,
		PageSize: pageSize,
		Data:     data,
	}
}
