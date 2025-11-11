package models

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message"`
	Meta    *Meta       `json:"meta,omitempty"`
}

type Meta struct {
	Total       int  `json:"total"`
	Limit       int  `json:"limit"`
	Page        int  `json:"page"`
	TotalPages  int  `json:"total_pages"`
	HasNext     bool `json:"has_next"`
	HasPrevious bool `json:"has_previous"`
}
