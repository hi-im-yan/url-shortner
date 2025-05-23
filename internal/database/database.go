package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/joho/godotenv/autoload"
)

// Service represents a service that interacts with a database.
type Service interface {
	// Health returns a map of health status information.
	// The keys and values in the map are service-specific.
	Health() map[string]string

	// Close terminates the database connection.
	// It returns an error if the connection cannot be closed.
	Close() error

	// Insert into database
	SaveShortUrl(*ShortUrlModel) (*ShortUrlModel, error)

	// Get the Shortned URL entity
	GetShortUrl(shortCode string) (*ShortUrlModel, error)

	// Update the shortned URL times_cliecked attribute
	UpdateTimesClicked(shortCode string) error

	// Delete expired links
	DeleteExpiredLinks() error
}

type service struct {
	db *sql.DB
}

var (
	database   = os.Getenv("BLUEPRINT_DB_DATABASE")
	password   = os.Getenv("BLUEPRINT_DB_PASSWORD")
	username   = os.Getenv("BLUEPRINT_DB_USERNAME")
	port       = os.Getenv("BLUEPRINT_DB_PORT")
	host       = os.Getenv("BLUEPRINT_DB_HOST")
	schema     = os.Getenv("BLUEPRINT_DB_SCHEMA")
	dbInstance *service
)

func New() Service {
	// Reuse Connection
	if dbInstance != nil {
		return dbInstance
	}
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&search_path=%s", username, password, host, port, database, schema)
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatal(err)
	}
	dbInstance = &service{
		db: db,
	}
	return dbInstance
}

// Health checks the health of the database connection by pinging the database.
// It returns a map with keys indicating various health statistics.
func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	stats := make(map[string]string)

	// Ping the database
	err := s.db.PingContext(ctx)
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		log.Fatalf("db down: %v", err) // Log the error and terminate the program
		return stats
	}

	// Database is up, add more statistics
	stats["status"] = "up"
	stats["message"] = "It's healthy"

	// Get database stats (like open connections, in use, idle, etc.)
	dbStats := s.db.Stats()
	stats["open_connections"] = strconv.Itoa(dbStats.OpenConnections)
	stats["in_use"] = strconv.Itoa(dbStats.InUse)
	stats["idle"] = strconv.Itoa(dbStats.Idle)
	stats["wait_count"] = strconv.FormatInt(dbStats.WaitCount, 10)
	stats["wait_duration"] = dbStats.WaitDuration.String()
	stats["max_idle_closed"] = strconv.FormatInt(dbStats.MaxIdleClosed, 10)
	stats["max_lifetime_closed"] = strconv.FormatInt(dbStats.MaxLifetimeClosed, 10)

	// Evaluate stats to provide a health message
	if dbStats.OpenConnections > 40 { // Assuming 50 is the max for this example
		stats["message"] = "The database is experiencing heavy load."
	}

	if dbStats.WaitCount > 1000 {
		stats["message"] = "The database has a high number of wait events, indicating potential bottlenecks."
	}

	if dbStats.MaxIdleClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many idle connections are being closed, consider revising the connection pool settings."
	}

	if dbStats.MaxLifetimeClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many connections are being closed due to max lifetime, consider increasing max lifetime or revising the connection usage pattern."
	}

	return stats
}

// Close closes the database connection.
// It logs a message indicating the disconnection from the specific database.
// If the connection is successfully closed, it returns nil.
// If an error occurs while closing the connection, it returns the error.
func (s *service) Close() error {
	log.Printf("Disconnected from database: %s", database)
	return s.db.Close()
}

func (s *service) SaveShortUrl(shortUrlModel *ShortUrlModel) (*ShortUrlModel, error) {
	query := "INSERT INTO short_url (link, times_clicked, exp_time_minutes, short_code) VALUES ($1, 0, $2, $3) RETURNING id, link, times_clicked, exp_time_minutes, short_code;"

	inserted := &ShortUrlModel{}
	err := s.db.QueryRow(query, shortUrlModel.Link, shortUrlModel.ExpTimeMinutes, shortUrlModel.ShortCode).Scan(&inserted.Id, &inserted.Link, &inserted.TimesClicked, &inserted.ExpTimeMinutes, &inserted.ShortCode)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			fmt.Println(pgErr.Message) // => syntax error at end of input
			fmt.Println(pgErr.Code)    // => 42601
			log.Printf("[database:SaveShortUrl] Error inserting short_url: %v", err)
			return nil, err
		}
	}

	log.Printf("[database:SaveShortUrl] Inserted: %+v", inserted)

	return inserted, nil
}

func (s *service) GetShortUrl(shortCode string) (*ShortUrlModel, error) {
	log.Printf("[database:GetShortUrl] Querying for shortCode: {%s}", shortCode)

	query := "SELECT link, times_clicked, exp_time_minutes, short_code, created_at FROM short_url WHERE short_code=$1;"

	searched := &ShortUrlModel{}
	err := s.db.QueryRow(query, shortCode).Scan(&searched.Link, &searched.TimesClicked, &searched.ExpTimeMinutes, &searched.ShortCode, &searched.CreatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			fmt.Println(pgErr.Message) // => syntax error at end of input
			fmt.Println(pgErr.Code)    // => 42601
			log.Printf("[database:GetShortUrl] Something went wrong: %v", err)
			return nil, err
		}

		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("[database:GetShortUrl] Query returned no rows: %+v", err)
			return nil, err
		}

	}

	log.Printf("[database:GetShortUrl] Found a url: %+v", searched)

	return searched, nil
}

func (s *service) UpdateTimesClicked(shortCode string) error {
	log.Printf("[database:UpdateTimesClicked] Updating times_clicked for shortCode: {%s}", shortCode)

	query := "UPDATE short_url SET times_clicked = times_clicked + 1 WHERE short_code = $1;"

	_, err := s.db.Exec(query, shortCode)

	if err != nil {
		log.Printf("[database:UpdateTimesClicked] something went wrong while updating for shortCode {%s}: %v", shortCode, err)
		return err
	}
	log.Printf("[database:UpdateTimesClicked] Times_clicked updated for shortCode: {%s}", shortCode)

	return nil
}

func (s *service) DeleteExpiredLinks() error {
	log.Printf("[database:DeleteExpiredLinks] Deleting expired links")

	query := "DELETE FROM short_url WHERE NOW() >= created_at + (exp_time_minutes || ' minutes')::interval;"

	_, err := s.db.Exec(query)

	if err != nil {
		log.Printf("[database:DeleteExpiredLinks] something went wrong: %v", err)
		return err
	}
	log.Printf("[database:DeleteExpiredLinks] Expired links deleted")

	return nil
}