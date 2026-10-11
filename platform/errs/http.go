package errs

type (
	// Message use for notification responses
	Message struct {
		Message string `json:"message"` // Notification message
	}

	// HTTPError use for error responses
	HTTPError struct {
		Error   string `json:"error"`   // Shorthand error
		Message string `json:"message"` // Pretty error message
	}
)
