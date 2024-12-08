package response

import "time"

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   *Error      `json:"error,omitempty"`
}

// Error represents an error in the response
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	Items      interface{} `json:"items"`
	TotalItems int64       `json:"total_items" example:"100"`
	Page       int         `json:"page" example:"1"`
	PageSize   int         `json:"page_size" example:"20"`
	TotalPages int         `json:"total_pages" example:"5"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Database  string    `json:"database"`
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
}

// NewResponse creates a new successful response
func NewResponse(message string, data interface{}) Response {
	return Response{
		Success: true,
		Message: message,
		Data:    data,
	}
}

// NewErrorResponse creates a new error response
func NewErrorResponse(code, message, details string) Response {
	return Response{
		Success: false,
		Error: &Error{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
}

// NewPaginatedResponse creates a new paginated response
func NewPaginatedResponse(items interface{}, totalItems int64, page, pageSize int) Response {
	totalPages := (int(totalItems) + pageSize - 1) / pageSize
	return NewResponse("", PaginatedResponse{
		Items:      items,
		TotalItems: totalItems,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

// NewHealthResponse creates a new health check response
func NewHealthResponse(version string) Response {
	return NewResponse("", HealthResponse{
		Status:    "ok",
		Timestamp: time.Now().UTC(),
		Version:   version,
	})
}
