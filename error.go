package raiden

type ErrorResponse struct {
	StatusCode int    `json:"-"`
	Code       string `json:"code,omitempty"`
	Details    any    `json:"details,omitempty"`
	Hint       string `json:"hint,omitempty"`
	Message    string `json:"message"`
}

func (err *ErrorResponse) Error() string {
	return err.Message
}
