package response

import (
	"encoding/json"
	"net/http"
)

type Meta struct {
	Code      int    `json:"code"`
	Status    string `json:"status"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}

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
