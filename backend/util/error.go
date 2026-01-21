package util

type ErrorResponse struct {
	Header  string `json:"Header"`
	Body    string `json:"Body"`
	Message string `json:"Message"` // For backward compatibility or extra info
}
