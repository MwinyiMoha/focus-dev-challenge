package domain

const SendMessageTask = "task:send_message"

type SendMessage struct {
	MessageID int64 `json:"message_id"`
}
