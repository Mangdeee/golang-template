package consttype

type QueueType string

const (
	SEND_EMAIL QueueType = "send_email"
)

func (q QueueType) String() string {
	return string(q)
}
