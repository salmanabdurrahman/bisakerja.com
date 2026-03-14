package response

import (
	"encoding/json"
	"net/http"
)

// Meta represents meta.
type Meta struct {
	Code       int         `json:"code"`
	Status     string      `json:"status"`
	Message    string      `json:"message"`
	RequestID  string      `json:"request_id,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

// Pagination represents pagination.
type Pagination struct {
	Page         int `json:"page"`
	Limit        int `json:"limit"`
	TotalPages   int `json:"total_pages"`
	TotalRecords int `json:"total_records"`
}

// ErrorItem represents error item.
type ErrorItem struct {
	Field   string `json:"field,omitempty"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type successEnvelope struct {
	Meta Meta `json:"meta"`
	Data any  `json:"data"`
}

type errorEnvelope struct {
	Meta   Meta        `json:"meta"`
	Errors []ErrorItem `json:"errors,omitempty"`
}

// WriteSuccess writes success.
func WriteSuccess(w http.ResponseWriter, code int, message, requestID string, data any) {
	writeJSON(w, code, successEnvelope{
		Meta: Meta{
			Code:      code,
			Status:    "success",
			Message:   message,
			RequestID: requestID,
		},
		Data: data,
	})
}

// WriteSuccessWithPagination writes success with pagination.
func WriteSuccessWithPagination(
	w http.ResponseWriter,
	code int,
	message, requestID string,
	data any,
	pagination Pagination,
) {
	writeJSON(w, code, successEnvelope{
		Meta: Meta{
			Code:       code,
			Status:     "success",
			Message:    message,
			RequestID:  requestID,
			Pagination: &pagination,
		},
		Data: data,
	})
}

// WriteError writes error.
func WriteError(w http.ResponseWriter, code int, message, requestID string, errors []ErrorItem) {
	writeJSON(w, code, errorEnvelope{
		Meta: Meta{
			Code:      code,
			Status:    "error",
			Message:   message,
			RequestID: requestID,
		},
		Errors: errors,
	})
}

func writeJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}
