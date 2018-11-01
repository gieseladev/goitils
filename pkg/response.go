package gitils

import (
	"net/http"

	"github.com/go-chi/render"
)

type JsonResponse struct {
	Writer  http.ResponseWriter
	Request *http.Request

	Payload interface{}
}

type ErrorResponse struct {
	Code    uint16      `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func NewJsonResponse(w http.ResponseWriter, r *http.Request, payload interface{}) *JsonResponse {
	return &JsonResponse{w, r, payload}
}

func (resp *JsonResponse) Send() {
	render.JSON(resp.Writer, resp.Request, resp.Payload)
}

func SendJsonResponse(w http.ResponseWriter, r *http.Request, payload interface{}) {
	NewJsonResponse(w, r, payload).Send()
}

func SendErrorResponse(w http.ResponseWriter, r *http.Request, code uint16, msg string, data interface{}) {
	SendJsonResponse(w, r, ErrorResponse{code, msg, data})
}
