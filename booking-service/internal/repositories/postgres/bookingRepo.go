package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/ollatomiwa/hotelsystem/booking-service/internal/models"
	"github.com/ollatomiwa/hotelsystem/booking-service/internal/repositories"
)

type BookingRepository struct {
	db *sql.DB 
}

func NewBookingRepository(db *sql.DB) *BookingRepository {
	return &BookingRepository{db :db}
}

var _ repositories.BookingRepository = (*BookingRepository)(nil)

//ceate creates a new booking with transaction
func (r *BookingRepository) CreateBooking(ctx context.Context, booking *models.Booking) error {
	//start transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil{
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	//check if room is available
	isAvailable, err := r.IsRoomAvailable(ctx, booking.RoomId, booking.CheckIn, booking.CheckOut)
	if err != nil {
		return fmt.Errorf("room availability check failed: %w", err)
	}
	if !isAvailable{
		return fmt.Errorf("room is not available for the selected date")
	}
	//insert booking
	query := `INSERT INTO bookings(id, user_id, room_id, room_type, check_in, check_out, guests, total_amount, status, created_at, updated_at) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	//insert into db
	_, err = tx.ExecContext(ctx, query,
		booking.Id,
		booking.UserId,
		booking.RoomId,
		booking.RoomType,
		booking.CheckIn,
		booking.CheckOut,
		booking.Guest,
		booking.TotalAmount,
		booking.Status,
		booking.CreatedAt,
		booking.UpdatedAt,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code.Name() == "foreign_key_violation" {
				return fmt.Errorf("room not found")
			}
		}
		return fmt.Errorf("failed to create booking: %w", err)
	}
	return tx.Commit()
}

//checks if a room is available for a given date
func (r *BookingRepository) IsRoomAvailable(ctx context.Context, roomID string, checkIn, checkOut time.Time) (bool, error) {
    query := `
        SELECT COUNT(*) FROM bookings 
        WHERE room_id = $1 
        AND status IN ('pending', 'confirmed')
        AND (check_in, check_out) OVERLAPS ($2, $3)
    `
    
    var count int
    err := r.db.QueryRowContext(ctx, query, roomID, checkIn, checkOut).Scan(&count)
    if err != nil {
        return false, fmt.Errorf("failed to check room availability: %w", err)
    }
    
    return count == 0, nil
}
//checking to find available rooms for given criteria
func (r *BookingRepository) GetAvailableRooms(ctx context.Context, req *models.AvailabilityRequest) ([]models.RoomAvailability, error) {
	checkIn, err := time.Parse("2006-01-02", req.CheckIn)
	if err != nil {
		return nil, fmt.Errorf("invalid check_in date fornat: %w", err)
	}

	checkOut, err := time.Parse("2006-01-02", req.CheckOut)
	if err != nil {
		return nil, fmt.Errorf("invalid check_out date fornat: %w", err)
	}

	query := `
		SELECT r.id, r.room_number, r.room_type, r.price_per_night, r.max_guests 
		FROM rooms r
		WHERE r.room_type = $1
		AND r.available = TRUE
		AND r.max_guests >= $2
		AND r.id NOT IN (
			SELECT b.room_id FROM bookings b
			WHERE b.status IN ('pending', 'confirmed')
			AND (b.check_in, b.check_out) OVERLAPS ($3, $4)
		)
		ORDER BY r.price_per_night ASC
	`
	rows, err := r.db.QueryContext(ctx, query,
		req.RoomType,
		req.Guests,
		checkIn,
		checkOut,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query available rooms: %w", err)
	}
	defer rows.Close()

	var AvailableRooms []models.RoomAvailability
	nights := int(checkOut.Sub(checkIn).Hours() / 24)

	for rows.Next() {
		var room models.RoomAvailability
		var pricePerNight float64

		err := rows.Scan(
			&room.RoomId,
			&room.RoomNumber,
			&room.RoomType,
			&room.PricePerNight,
			&room.MaxGuests,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan room: %w", err)
		}

		room.PricePerNight = pricePerNight
		room.TotalPrice =pricePerNight * float64(nights)
		AvailableRooms = append(AvailableRooms, room)
	}
	return AvailableRooms, nil
}

//retrieves bookings by its Id
func (r *BookingRepository) GetBookingById(ctx context.Context, id string) (*models.Booking, error) {
	query := `
	SELECT id, user_id, room_id, room_type, check_in, check_out, guests, total_amount, status, created_at, updated_at
	FROM bookings WHERE id = $1
	`
	var booking models.Booking
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&booking.Id,
		&booking.UserId,
		&booking.RoomId,
		&booking.RoomType,
		&booking.CheckIn,
		&booking.CheckOut,
		&booking.Guest,
		&booking.TotalAmount,
		&booking.Status,
		&booking.CreatedAt,
		&booking.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("booking not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get booking: %w", err)
	}

	return &booking, nil
}

// retrieves all bookinfs for a user
func (r *BookingRepository) GetUserBookings(ctx context.Context, userId string) ([]models.Booking, error) {
	query := `
		SELECT id, user_id, room_id, room_type, check_in, check_out, guests, total_amount, status, created_at, updated_at
		FROM bookings WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to query user bookings: %w", err)
	}
	defer rows.Close()

	var bookings []models.Booking
		for rows.Next() {
			var booking models.Booking
			err := rows.Scan(
				&booking.Id,
				&booking.UserId,
				&booking.RoomId,
				&booking.RoomType,
				&booking.CheckIn,
				&booking.CheckOut,
				&booking.Guest,
				&booking.TotalAmount,
				&booking.Status,
				&booking.CreatedAt,
				&booking.UpdatedAt,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to scan booking: %w", err)
			}
			bookings = append(bookings, booking)
		}
		
	return bookings, nil
	
}

//update the status of a booking
func (r *BookingRepository) UpdateBookingStatus(ctx context.Context, id string, status models.BookingStatus ) error {
	query := `
		UPDATE bookings SET status = $1, updated_at =NOW() WHERE id =$2`
		
		results, err := r.db.ExecContext(ctx, query, status, id)
		if err != nil {
			return fmt.Errorf("failed to updated booking statud: %w", err)
		}

		rowsAffected, err := results.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}

		if rowsAffected == 0 {
			return fmt.Errorf("booking not found")
		}

		return nil
}