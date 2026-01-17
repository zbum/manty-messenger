package service

import (
	"context"
	"encoding/json"
	"log"

	"github.com/SherClockHolmes/webpush-go"

	"Mmessenger/internal/config"
	"Mmessenger/internal/models"
	"Mmessenger/internal/repository"
)

type PushService struct {
	pushRepo   *repository.PushRepository
	memberRepo *repository.RoomMemberRepository
	vapidCfg   *config.WebPushConfig
}

func NewPushService(
	pushRepo *repository.PushRepository,
	memberRepo *repository.RoomMemberRepository,
	vapidCfg *config.WebPushConfig,
) *PushService {
	return &PushService{
		pushRepo:   pushRepo,
		memberRepo: memberRepo,
		vapidCfg:   vapidCfg,
	}
}

// Subscribe creates or updates a push subscription for a user
func (s *PushService) Subscribe(ctx context.Context, userID uint64, req *models.SubscribePushRequest) error {
	sub := &models.PushSubscription{
		UserID:   userID,
		Endpoint: req.Endpoint,
		P256dh:   req.Keys.P256dh,
		Auth:     req.Keys.Auth,
	}
	return s.pushRepo.Create(ctx, sub)
}

// Unsubscribe removes all push subscriptions for a user
func (s *PushService) Unsubscribe(ctx context.Context, userID uint64) error {
	return s.pushRepo.DeleteByUserID(ctx, userID)
}

// GetVAPIDPublicKey returns the VAPID public key
func (s *PushService) GetVAPIDPublicKey() string {
	return s.vapidCfg.VAPIDPublicKey
}

// IsConfigured returns true if VAPID keys are configured
func (s *PushService) IsConfigured() bool {
	return s.vapidCfg.VAPIDPublicKey != "" && s.vapidCfg.VAPIDPrivateKey != ""
}

// SendToUser sends a push notification to all devices of a user
func (s *PushService) SendToUser(ctx context.Context, userID uint64, notification *models.PushNotification) error {
	if !s.IsConfigured() {
		log.Println("Web Push not configured, skipping notification")
		return nil
	}

	subs, err := s.pushRepo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}

	for _, sub := range subs {
		if err := s.sendNotification(sub, notification); err != nil {
			log.Printf("Failed to send push to user %d: %v", userID, err)
			// If subscription is invalid, remove it
			if isSubscriptionGone(err) {
				s.pushRepo.DeleteByEndpoint(ctx, sub.Endpoint)
			}
		}
	}

	return nil
}

// SendToRoomMembers sends a push notification to all members of a room except the sender
func (s *PushService) SendToRoomMembers(ctx context.Context, roomID, senderID uint64, notification *models.PushNotification) error {
	if !s.IsConfigured() {
		return nil
	}

	// Get room members
	members, err := s.memberRepo.GetByRoomID(ctx, roomID)
	if err != nil {
		return err
	}

	// Collect user IDs excluding sender
	var userIDs []uint64
	for _, member := range members {
		if member.UserID != senderID {
			userIDs = append(userIDs, member.UserID)
		}
	}

	if len(userIDs) == 0 {
		return nil
	}

	// Get all subscriptions for these users
	subs, err := s.pushRepo.GetByUserIDs(ctx, userIDs)
	if err != nil {
		return err
	}

	// Send notifications
	for _, sub := range subs {
		if err := s.sendNotification(sub, notification); err != nil {
			log.Printf("Failed to send push to subscription %d: %v", sub.ID, err)
			if isSubscriptionGone(err) {
				s.pushRepo.DeleteByEndpoint(ctx, sub.Endpoint)
			}
		}
	}

	return nil
}

// sendNotification sends a push notification to a single subscription
func (s *PushService) sendNotification(sub *models.PushSubscription, notification *models.PushNotification) error {
	payload, err := json.Marshal(notification)
	if err != nil {
		return err
	}

	subscription := &webpush.Subscription{
		Endpoint: sub.Endpoint,
		Keys: webpush.Keys{
			P256dh: sub.P256dh,
			Auth:   sub.Auth,
		},
	}

	resp, err := webpush.SendNotification(payload, subscription, &webpush.Options{
		Subscriber:      s.vapidCfg.VAPIDSubject,
		VAPIDPublicKey:  s.vapidCfg.VAPIDPublicKey,
		VAPIDPrivateKey: s.vapidCfg.VAPIDPrivateKey,
		TTL:             60,
	})

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check for errors
	if resp.StatusCode >= 400 {
		return &pushError{StatusCode: resp.StatusCode}
	}

	return nil
}

type pushError struct {
	StatusCode int
}

func (e *pushError) Error() string {
	return "push notification failed"
}

func isSubscriptionGone(err error) bool {
	if pe, ok := err.(*pushError); ok {
		// 404 (Not Found) or 410 (Gone) means subscription is no longer valid
		return pe.StatusCode == 404 || pe.StatusCode == 410
	}
	return false
}
