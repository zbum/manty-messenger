package service

import (
	"context"
	"fmt"
	"log"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"

	"Mmessenger/internal/config"
	"Mmessenger/internal/models"
	"Mmessenger/internal/repository"
)

type FCMService struct {
	app             *firebase.App
	client          *messaging.Client
	deviceTokenRepo *repository.DeviceTokenRepository
	enabled         bool
}

func NewFCMService(cfg *config.FCMConfig, deviceTokenRepo *repository.DeviceTokenRepository) (*FCMService, error) {
	svc := &FCMService{
		deviceTokenRepo: deviceTokenRepo,
		enabled:         false,
	}

	if cfg.CredentialsFile == "" {
		log.Println("[FCM] No credentials file configured, FCM push disabled")
		return svc, nil
	}

	app, err := firebase.NewApp(context.Background(), nil, option.WithCredentialsFile(cfg.CredentialsFile))
	if err != nil {
		return svc, fmt.Errorf("failed to initialize Firebase app: %w", err)
	}

	client, err := app.Messaging(context.Background())
	if err != nil {
		return svc, fmt.Errorf("failed to get Firebase messaging client: %w", err)
	}

	svc.app = app
	svc.client = client
	svc.enabled = true
	log.Println("[FCM] Firebase Cloud Messaging initialized successfully")

	return svc, nil
}

func (s *FCMService) IsEnabled() bool {
	return s.enabled
}

func (s *FCMService) RegisterDevice(ctx context.Context, userID uint64, req *models.RegisterDeviceRequest) error {
	token := &models.DeviceToken{
		UserID:   userID,
		DeviceID: req.DeviceID,
		Platform: req.Platform,
		Token:    req.Token,
	}
	return s.deviceTokenRepo.Upsert(ctx, token)
}

func (s *FCMService) UnregisterDevice(ctx context.Context, userID uint64, deviceID string) error {
	return s.deviceTokenRepo.DeleteByUserAndDevice(ctx, userID, deviceID)
}

func (s *FCMService) SendToUsers(ctx context.Context, userIDs []uint64, notification *models.PushNotification) error {
	if !s.enabled {
		return nil
	}

	tokens, err := s.deviceTokenRepo.GetByUserIDs(ctx, userIDs)
	if err != nil {
		return fmt.Errorf("failed to get device tokens: %w", err)
	}

	if len(tokens) == 0 {
		return nil
	}

	// Convert notification data to map[string]string for FCM
	data := make(map[string]string)
	for k, v := range notification.Data {
		data[k] = fmt.Sprintf("%v", v)
	}

	for _, token := range tokens {
		msg := &messaging.Message{
			Token: token.Token,
			Notification: &messaging.Notification{
				Title: notification.Title,
				Body:  notification.Body,
			},
			Data: data,
			Android: &messaging.AndroidConfig{
				Priority: "high",
				Notification: &messaging.AndroidNotification{
					ClickAction: "OPEN_CHAT",
					Sound:       "default",
				},
			},
			APNS: &messaging.APNSConfig{
				Payload: &messaging.APNSPayload{
					Aps: &messaging.Aps{
						Sound: "default",
						Badge: intPtr(1),
					},
				},
			},
		}

		_, err := s.client.Send(ctx, msg)
		if err != nil {
			if messaging.IsUnregistered(err) {
				log.Printf("[FCM] Token expired for device %s, removing", token.DeviceID)
				s.deviceTokenRepo.DeleteByToken(ctx, token.Token)
			} else {
				log.Printf("[FCM] Failed to send to device %s: %v", token.DeviceID, err)
			}
		}
	}

	return nil
}

func intPtr(i int) *int {
	return &i
}
