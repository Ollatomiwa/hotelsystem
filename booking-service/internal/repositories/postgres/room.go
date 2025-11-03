package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ollatomiwa/hotelsystem/booking-service/internal/models"
)

type RoomRepository struct {
	db *sql.DB 
}
func NewRoomRepository(db *sql.DB) *RoomRepository {
	return &RoomRepository{db:db}
}

//retrieves room by its Id
func (r *RoomRepository) GetRoomById(ctx context.Context, id string) (*models.Room, error) {
	query := `
		SELECT id, room_number, room_type, price_per_night, max_guests, available, description`

		var room models.Room
		err := r.db.QueryRowContext(ctx, query, id).Scan(
			&room.Id,
			&room.RoomNumber,
			&room.RoomType,
			&room.PricePerNight,
			&room.MaxGuests,
			&room.Avalaible,
			&room.Description,
		)

		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("room not found")
		}
		if err != nil {
			return nil, fmt.Errorf("failed to get room: %w", err)
		}

		return &room, nil
}

//creates a new room (for admin purposes only)
func (r *RoomRepository) CreateRoom(ctx context.Context, room *models.Room) error {
	query := `INSERT INTO rooms (id. room_number, room_type, price_per_night, max_guests, available, description) VALUES ($1, $2, $3, $4. $5, $6. $7)`

	_, err := r.db.ExecContext(ctx, query, 
		room.Id,
		room.RoomNumber,
		room.RoomType,
		room.PricePerNight,
		room.MaxGuests,
		room.Avalaible,
		room.Description,
	)
	if err != nil {
		return fmt.Errorf("failed to create room: %w", err)
	}
	return nil 
}

//retrieves all rooms (for admin purposes only)
func (r *RoomRepository) GetAllRooms(ctx context.Context) ([]models.Room, error) {
	query := `SELECT id, room_number, room_type, price_per_night, max_guests, available, description FROM rooms ORDER BY room_number`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query rooms: %w", err)
	}
	defer rows.Close()

	var rooms []models.Room
	for rows.Next() {
		var room models.Room
		err := rows.Scan(
			&room.Id,
			&room.RoomNumber,
			&room.RoomType,
			&room.PricePerNight,
			&room.MaxGuests,
			&room.Avalaible,
			&room.Description,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan room: %w", err)
		}
		rooms = append(rooms, room)
	}
	return rooms, nil
}