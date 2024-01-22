package raiden

type ErrorResponse struct {
	StatusCode int
	Code       string `json:"code"`
	Details    any    `json:"details"`
	Hint       string `json:"hint"`
	Message    string `json:"message"`
}

func (err *ErrorResponse) Error() string {
	return err.Message
}
