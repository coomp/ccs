package services

type MessageQueueRepository interface {
	Send(msg string) error
	SendTimeout(msg string, timeout_sec int) error
}

type MessageQueueService struct {
	Repo MessageQueueRepository
}

func NewMessageQueueService(repo MessageQueueRepository) *MessageQueueService {
	return &MessageQueueService{
		Repo: repo,
	}
}

func (s *MessageQueueService) Send(msg string) error {
	return s.Repo.Send(msg)
}
