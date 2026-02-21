package service

import (
	"context"
	"log"
	"sync"

	"Mmessenger/internal/models"
	"Mmessenger/internal/repository"
)

// OnlineChecker checks if a user has an active WebSocket connection.
// The websocket.Hub already implements this interface.
type OnlineChecker interface {
	IsUserOnline(userID uint64) bool
}

type NotificationService struct {
	pushService    *PushService
	fcmService     *FCMService
	memberRepo     *repository.RoomMemberRepository
	onlineChecker  OnlineChecker
}

func NewNotificationService(
	pushService *PushService,
	fcmService *FCMService,
	memberRepo *repository.RoomMemberRepository,
	onlineChecker OnlineChecker,
) *NotificationService {
	return &NotificationService{
		pushService:   pushService,
		fcmService:    fcmService,
		memberRepo:    memberRepo,
		onlineChecker: onlineChecker,
	}
}

// NotifyRoomMembers sends push notifications to offline room members.
// It sends both Web Push and FCM notifications concurrently.
func (s *NotificationService) NotifyRoomMembers(ctx context.Context, roomID, senderID uint64, notification *models.PushNotification) error {
	members, err := s.memberRepo.GetByRoomID(ctx, roomID)
	if err != nil {
		log.Printf("[Notification] Failed to get room members: %v", err)
		return err
	}

	// Collect offline user IDs (exclude sender)
	var offlineUserIDs []uint64
	for _, member := range members {
		if member.UserID == senderID {
			continue
		}
		if !s.onlineChecker.IsUserOnline(member.UserID) {
			offlineUserIDs = append(offlineUserIDs, member.UserID)
		}
	}

	if len(offlineUserIDs) == 0 {
		return nil
	}

	log.Printf("[Notification] Sending notifications to %d offline users", len(offlineUserIDs))

	var wg sync.WaitGroup

	// Send Web Push notifications
	if s.pushService != nil && s.pushService.IsConfigured() {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for _, userID := range offlineUserIDs {
				if err := s.pushService.SendToUser(ctx, userID, notification); err != nil {
					log.Printf("[Notification] Web Push failed for user %d: %v", userID, err)
				}
			}
		}()
	}

	// Send FCM notifications
	if s.fcmService != nil && s.fcmService.IsEnabled() {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := s.fcmService.SendToUsers(ctx, offlineUserIDs, notification); err != nil {
				log.Printf("[Notification] FCM failed: %v", err)
			}
		}()
	}

	wg.Wait()
	return nil
}
