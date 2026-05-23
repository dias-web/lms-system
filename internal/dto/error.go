package dto

// ErrorResponse is the unified error envelope returned by all endpoints.
// @name ErrorResponse
type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

// ErrorBody carries the machine-readable code and human-readable message.
// @name ErrorBody
type ErrorBody struct {
	Code    string `json:"code"    example:"COURSE_NOT_FOUND"`
	Message string `json:"message" example:"course not found"`
}

func NewErrorResponse(code, message string) ErrorResponse {
	return ErrorResponse{
		Error: ErrorBody{Code: code, Message: message},
	}
}
