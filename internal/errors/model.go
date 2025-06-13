package errors

var ResponseErrorNotFound = ResponseError{Message: "not found"}

type ResponseError struct {
	Message string `json:"error"`
	Details string `json:"details"`
}

func (r ResponseError) Error() string {
	return r.Message
}
