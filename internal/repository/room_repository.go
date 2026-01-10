package repository

import (
	"context"
	"database/sql"

	"Mmessenger/internal/models"
)

type RoomRepository struct {
	db *sql.DB
}

func NewRoomRepository(db *sql.DB) *RoomRepository {
	return &RoomRepository{db: db}
}

func (r *RoomRepository) Create(ctx context.Context, room *models.Room) error {
	query := `
		INSERT INTO rooms (name, description, room_type, owner_id, max_members)
		VALUES (?, ?, ?, ?, ?)
	`
	result, err := r.db.ExecContext(ctx, query,
		room.Name, room.Description, room.RoomType, room.OwnerID, room.MaxMembers,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	room.ID = uint64(id)
	return nil
}

func (r *RoomRepository) GetByID(ctx context.Context, id uint64) (*models.Room, error) {
	query := `
		SELECT id, name, description, room_type, owner_id, avatar_url, max_members, created_at, updated_at
		FROM rooms WHERE id = ?
	`
	room := &models.Room{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&room.ID, &room.Name, &room.Description, &room.RoomType,
		&room.OwnerID, &room.AvatarURL, &room.MaxMembers,
		&room.CreatedAt, &room.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return room, nil
}

func (r *RoomRepository) GetByUserID(ctx context.Context, userID uint64) ([]*models.Room, error) {
	query := `
		SELECT r.id, r.name, r.description, r.room_type, r.owner_id, r.avatar_url, r.max_members, r.created_at, r.updated_at
		FROM rooms r
		INNER JOIN room_members rm ON r.id = rm.room_id
		WHERE rm.user_id = ?
		ORDER BY r.updated_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []*models.Room
	for rows.Next() {
		room := &models.Room{}
		err := rows.Scan(
			&room.ID, &room.Name, &room.Description, &room.RoomType,
			&room.OwnerID, &room.AvatarURL, &room.MaxMembers,
			&room.CreatedAt, &room.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}
	return rooms, nil
}

func (r *RoomRepository) Update(ctx context.Context, room *models.Room) error {
	query := `
		UPDATE rooms SET name = ?, description = ?, updated_at = NOW()
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query, room.Name, room.Description, room.ID)
	return err
}

func (r *RoomRepository) Delete(ctx context.Context, id uint64) error {
	query := `DELETE FROM rooms WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *RoomRepository) GetMemberCount(ctx context.Context, roomID uint64) (int, error) {
	query := `SELECT COUNT(*) FROM room_members WHERE room_id = ?`
	var count int
	err := r.db.QueryRowContext(ctx, query, roomID).Scan(&count)
	return count, err
}
