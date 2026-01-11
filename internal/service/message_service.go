package service

import (
	"context"

	"Mmessenger/internal/models"
	"Mmessenger/internal/repository"
)

type MessageService struct {
	messageRepo *repository.MessageRepository
	memberRepo  *repository.RoomMemberRepository
	userRepo    *repository.UserRepository
}

func NewMessageService(messageRepo *repository.MessageRepository, memberRepo *repository.RoomMemberRepository, userRepo *repository.UserRepository) *MessageService {
	return &MessageService{
		messageRepo: messageRepo,
		memberRepo:  memberRepo,
		userRepo:    userRepo,
	}
}

func (s *MessageService) Create(ctx context.Context, roomID, senderID uint64, req *models.SendMessageRequest) (*models.MessageResponse, error) {
	isMember, err := s.memberRepo.IsMember(ctx, roomID, senderID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrNotMember
	}

	msg := &models.Message{
		RoomID:      roomID,
		SenderID:    senderID,
		Content:     req.Content,
		MessageType: req.MessageType,
	}

	if msg.MessageType == "" {
		msg.MessageType = models.MessageTypeText
	}

	if err := s.messageRepo.Create(ctx, msg); err != nil {
		return nil, err
	}

	// Get unread count (all members except sender haven't read yet)
	unreadCount, _ := s.messageRepo.GetUnreadCount(ctx, roomID, msg.CreatedAt, senderID)

	sender, _ := s.userRepo.GetByID(ctx, senderID)
	return msg.ToResponse(sender.ToResponse(), unreadCount), nil
}

func (s *MessageService) GetByRoomID(ctx context.Context, roomID, userID uint64, limit, offset int) ([]*models.MessageResponse, error) {
	isMember, err := s.memberRepo.IsMember(ctx, roomID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrNotMember
	}

	messages, err := s.messageRepo.GetByRoomID(ctx, roomID, limit, offset)
	if err != nil {
		return nil, err
	}

	// Cache users
	userCache := make(map[uint64]*models.UserResponse)

	var responses []*models.MessageResponse
	for _, msg := range messages {
		sender, ok := userCache[msg.SenderID]
		if !ok {
			user, err := s.userRepo.GetByID(ctx, msg.SenderID)
			if err == nil {
				sender = user.ToResponse()
				userCache[msg.SenderID] = sender
			}
		}
		unreadCount, _ := s.messageRepo.GetUnreadCount(ctx, roomID, msg.CreatedAt, msg.SenderID)
		responses = append(responses, msg.ToResponse(sender, unreadCount))
	}
	return responses, nil
}

func (s *MessageService) GetByID(ctx context.Context, roomID, msgID, userID uint64) (*models.MessageResponse, error) {
	isMember, err := s.memberRepo.IsMember(ctx, roomID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrNotMember
	}

	msg, err := s.messageRepo.GetByID(ctx, msgID)
	if err != nil {
		return nil, err
	}

	unreadCount, _ := s.messageRepo.GetUnreadCount(ctx, roomID, msg.CreatedAt, msg.SenderID)
	sender, _ := s.userRepo.GetByID(ctx, msg.SenderID)
	return msg.ToResponse(sender.ToResponse(), unreadCount), nil
}

func (s *MessageService) Update(ctx context.Context, msgID, userID uint64, req *models.UpdateMessageRequest) (*models.MessageResponse, error) {
	msg, err := s.messageRepo.GetByID(ctx, msgID)
	if err != nil {
		return nil, err
	}

	if msg.SenderID != userID {
		return nil, ErrNotOwner
	}

	msg.Content = req.Content
	if err := s.messageRepo.Update(ctx, msg); err != nil {
		return nil, err
	}

	msg.IsEdited = true
	unreadCount, _ := s.messageRepo.GetUnreadCount(ctx, msg.RoomID, msg.CreatedAt, msg.SenderID)
	sender, _ := s.userRepo.GetByID(ctx, userID)
	return msg.ToResponse(sender.ToResponse(), unreadCount), nil
}

func (s *MessageService) Delete(ctx context.Context, msgID, userID uint64) error {
	msg, err := s.messageRepo.GetByID(ctx, msgID)
	if err != nil {
		return err
	}

	if msg.SenderID != userID {
		return ErrNotOwner
	}

	return s.messageRepo.Delete(ctx, msgID)
}
