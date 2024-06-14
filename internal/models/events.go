package models

type (
	BaseEvent struct {
		Event string `json:"event"`
	}
	BaseSubEvent struct {
		Event  string `json:"event"`
		UserId int    `json:"user_id"`
	}
	SuccessResponse struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	}
)

const (
	MessageRead = "MessageRead"
	Auth        = "Auth"
)
