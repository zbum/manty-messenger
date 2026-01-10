package service

import (
	"context"
	"database/sql"
	"errors"

	"Mmessenger/internal/models"
	"Mmessenger/internal/repository"
)

var (
	ErrNotMember        = errors.New("not a member of this room")
	ErrNotOwner         = errors.New("not the owner of this room")
	ErrOwnerCannotLeave = errors.New("owner cannot leave the room")
)

type RoomService struct {
	roomRepo       *repository.RoomRepository
	memberRepo     *repository.RoomMemberRepository
	userRepo       *repository.UserRepository
}

func NewRoomService(roomRepo *repository.RoomRepository, memberRepo *repository.RoomMemberRepository, userRepo *repository.UserRepository) *RoomService {
	return &RoomService{
		roomRepo:   roomRepo,
		memberRepo: memberRepo,
		userRepo:   userRepo,
	}
}

func (s *RoomService) Create(ctx context.Context, ownerID uint64, req *models.CreateRoomRequest) (*models.RoomResponse, error) {
	room := &models.Room{
		Name:       req.Name,
		RoomType:   req.RoomType,
		OwnerID:    ownerID,
		MaxMembers: 100,
	}

	if req.Description != "" {
		room.Description = sql.NullString{String: req.Description, Valid: true}
	}

	if room.RoomType == "" {
		room.RoomType = models.RoomTypeGroup
	}

	if err := s.roomRepo.Create(ctx, room); err != nil {
		return nil, err
	}

	// Add owner as member
	member := &models.RoomMember{
		RoomID: room.ID,
		UserID: ownerID,
		Role:   models.MemberRoleOwner,
	}
	if err := s.memberRepo.Add(ctx, member); err != nil {
		return nil, err
	}

	// Add other members if specified
	for _, userID := range req.MemberIDs {
		if userID != ownerID {
			m := &models.RoomMember{
				RoomID: room.ID,
				UserID: userID,
				Role:   models.MemberRoleMember,
			}
			s.memberRepo.Add(ctx, m)
		}
	}

	resp := room.ToResponse()
	resp.MemberCount = 1 + len(req.MemberIDs)
	return resp, nil
}

func (s *RoomService) GetByID(ctx context.Context, roomID, userID uint64) (*models.RoomResponse, error) {
	isMember, err := s.memberRepo.IsMember(ctx, roomID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrNotMember
	}

	room, err := s.roomRepo.GetByID(ctx, roomID)
	if err != nil {
		return nil, err
	}

	count, _ := s.roomRepo.GetMemberCount(ctx, roomID)
	resp := room.ToResponse()
	resp.MemberCount = count
	return resp, nil
}

func (s *RoomService) GetByUserID(ctx context.Context, userID uint64) ([]*models.RoomResponse, error) {
	rooms, err := s.roomRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var responses []*models.RoomResponse
	for _, room := range rooms {
		resp := room.ToResponse()
		count, _ := s.roomRepo.GetMemberCount(ctx, room.ID)
		resp.MemberCount = count
		responses = append(responses, resp)
	}
	return responses, nil
}

func (s *RoomService) Update(ctx context.Context, roomID, userID uint64, req *models.UpdateRoomRequest) (*models.RoomResponse, error) {
	room, err := s.roomRepo.GetByID(ctx, roomID)
	if err != nil {
		return nil, err
	}

	if room.OwnerID != userID {
		return nil, ErrNotOwner
	}

	if req.Name != nil {
		room.Name = *req.Name
	}
	if req.Description != nil {
		room.Description = sql.NullString{String: *req.Description, Valid: true}
	}

	if err := s.roomRepo.Update(ctx, room); err != nil {
		return nil, err
	}

	return room.ToResponse(), nil
}

func (s *RoomService) Delete(ctx context.Context, roomID, userID uint64) error {
	room, err := s.roomRepo.GetByID(ctx, roomID)
	if err != nil {
		return err
	}

	if room.OwnerID != userID {
		return ErrNotOwner
	}

	return s.roomRepo.Delete(ctx, roomID)
}

func (s *RoomService) GetMembers(ctx context.Context, roomID, userID uint64) ([]*models.RoomMemberResponse, error) {
	isMember, err := s.memberRepo.IsMember(ctx, roomID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrNotMember
	}

	members, err := s.memberRepo.GetByRoomID(ctx, roomID)
	if err != nil {
		return nil, err
	}

	var responses []*models.RoomMemberResponse
	for _, member := range members {
		user, err := s.userRepo.GetByID(ctx, member.UserID)
		if err != nil {
			continue
		}
		responses = append(responses, &models.RoomMemberResponse{
			User:     user.ToResponse(),
			Role:     member.Role,
			JoinedAt: member.JoinedAt,
		})
	}
	return responses, nil
}

func (s *RoomService) AddMember(ctx context.Context, roomID, requesterID, userID uint64) error {
	isMember, err := s.memberRepo.IsMember(ctx, roomID, requesterID)
	if err != nil {
		return err
	}
	if !isMember {
		return ErrNotMember
	}

	member := &models.RoomMember{
		RoomID: roomID,
		UserID: userID,
		Role:   models.MemberRoleMember,
	}
	return s.memberRepo.Add(ctx, member)
}

func (s *RoomService) RemoveMember(ctx context.Context, roomID, requesterID, userID uint64) error {
	room, err := s.roomRepo.GetByID(ctx, roomID)
	if err != nil {
		return err
	}

	if room.OwnerID != requesterID {
		return ErrNotOwner
	}

	return s.memberRepo.Remove(ctx, roomID, userID)
}

func (s *RoomService) Leave(ctx context.Context, roomID, userID uint64) error {
	room, err := s.roomRepo.GetByID(ctx, roomID)
	if err != nil {
		return err
	}

	if room.OwnerID == userID {
		return ErrOwnerCannotLeave
	}

	return s.memberRepo.Remove(ctx, roomID, userID)
}
