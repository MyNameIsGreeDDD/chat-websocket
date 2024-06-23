package models

type (
	BaseEvent struct {
		Event     string `json:"event" validate:"required"`
		SessionId string `json:"session_id" validate:"required"`
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
	Connected   = "Connected"
)
