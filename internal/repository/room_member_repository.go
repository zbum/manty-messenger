package repository

import (
	"context"
	"database/sql"

	"Mmessenger/internal/models"
)

type RoomMemberRepository struct {
	db *sql.DB
}

func NewRoomMemberRepository(db *sql.DB) *RoomMemberRepository {
	return &RoomMemberRepository{db: db}
}

func (r *RoomMemberRepository) Add(ctx context.Context, member *models.RoomMember) error {
	query := `
		INSERT INTO room_members (room_id, user_id, role)
		VALUES (?, ?, ?)
	`
	result, err := r.db.ExecContext(ctx, query, member.RoomID, member.UserID, member.Role)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	member.ID = uint64(id)
	return nil
}

func (r *RoomMemberRepository) GetByRoomID(ctx context.Context, roomID uint64) ([]*models.RoomMember, error) {
	query := `
		SELECT id, room_id, user_id, role, joined_at, last_read_at
		FROM room_members WHERE room_id = ?
	`
	rows, err := r.db.QueryContext(ctx, query, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*models.RoomMember
	for rows.Next() {
		member := &models.RoomMember{}
		err := rows.Scan(
			&member.ID, &member.RoomID, &member.UserID,
			&member.Role, &member.JoinedAt, &member.LastReadAt,
		)
		if err != nil {
			return nil, err
		}
		members = append(members, member)
	}
	return members, nil
}

func (r *RoomMemberRepository) GetMember(ctx context.Context, roomID, userID uint64) (*models.RoomMember, error) {
	query := `
		SELECT id, room_id, user_id, role, joined_at, last_read_at
		FROM room_members WHERE room_id = ? AND user_id = ?
	`
	member := &models.RoomMember{}
	err := r.db.QueryRowContext(ctx, query, roomID, userID).Scan(
		&member.ID, &member.RoomID, &member.UserID,
		&member.Role, &member.JoinedAt, &member.LastReadAt,
	)
	if err != nil {
		return nil, err
	}
	return member, nil
}

func (r *RoomMemberRepository) IsMember(ctx context.Context, roomID, userID uint64) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM room_members WHERE room_id = ? AND user_id = ?)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, roomID, userID).Scan(&exists)
	return exists, err
}

func (r *RoomMemberRepository) Remove(ctx context.Context, roomID, userID uint64) error {
	query := `DELETE FROM room_members WHERE room_id = ? AND user_id = ?`
	_, err := r.db.ExecContext(ctx, query, roomID, userID)
	return err
}

func (r *RoomMemberRepository) UpdateLastRead(ctx context.Context, roomID, userID uint64) error {
	query := `UPDATE room_members SET last_read_at = NOW() WHERE room_id = ? AND user_id = ?`
	_, err := r.db.ExecContext(ctx, query, roomID, userID)
	return err
}

func (r *RoomMemberRepository) GetUserIDsByRoomID(ctx context.Context, roomID uint64) ([]uint64, error) {
	query := `SELECT user_id FROM room_members WHERE room_id = ?`
	rows, err := r.db.QueryContext(ctx, query, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userIDs []uint64
	for rows.Next() {
		var userID uint64
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}
	return userIDs, nil
}
