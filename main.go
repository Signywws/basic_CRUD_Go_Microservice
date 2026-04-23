package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Сущность

type Note struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
}

// Repository layer

type NoteRepository interface {
	Create(note Note) (Note, error)
	GetAll() ([]Note, error)
	GetById(id int) (Note, error)
	DeleteById(id int) error
	UpdateById(id int, text string) (Note, error)
}

type InMemoryNoteRepository struct {
	notes  []Note
	nextID int
}

func (r *InMemoryNoteRepository) Create(note Note) (Note, error) {
	note.ID = r.nextID
	r.nextID++

	r.notes = append(r.notes, note)
	return note, nil
}

func (r *InMemoryNoteRepository) GetAll() ([]Note, error) {
	return r.notes, nil
}

type PostgresNoteRepository struct {
	db *sql.DB
}

func (p *PostgresNoteRepository) Create(note Note) (Note, error) {
	query := "INSERT INTO notes (text) VALUES ($1) RETURNING id"

	err := p.db.QueryRow(query, note.Text).Scan(&note.ID)
	if err != nil {
		return Note{}, err
	}

	return note, nil
}

func (p *PostgresNoteRepository) GetAll() ([]Note, error) {
	rows, err := p.db.Query("SELECT id, text FROM notes")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var notes []Note

	for rows.Next() {
		var note Note
		err := rows.Scan(&note.ID, &note.Text)
		if err != nil {
			return nil, err
		}

		notes = append(notes, note)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return notes, nil
}

func (p *PostgresNoteRepository) GetById(id int) (Note, error) {
	var note Note
	err := p.db.QueryRow(
		"SELECT id, text FROM notes WHERE id = $1", id).Scan(&note.ID, &note.Text)
	if err != nil {
		return Note{}, err
	}
	return note, nil
}

func (p *PostgresNoteRepository) DeleteById(id int) error {
	res, err := p.db.Exec("DELETE FROM notes WHERE id = $1", id)
	if err != nil {
		return err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (p *PostgresNoteRepository) UpdateById(id int, text string) (Note, error) {

	var note Note
	query := "UPDATE notes SET text = $1 WHERE id = $2 RETURNING id, text"
	err := p.db.QueryRow(query, text, id).Scan(&note.ID, &note.Text)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Note{}, sql.ErrNoRows
		}
		return Note{}, err
	}
	return note, nil
}

// Service Layer
type Service struct {
	repo NoteRepository
}

func (s *Service) Create(text string) (Note, error) {
	note := Note{
		Text: text,
	}

	if text == "" {
		return Note{}, errors.New("Text is required")
	}

	note, err := s.repo.Create(note)
	if err != nil {
		return Note{}, err
	}

	return note, nil

}
func (s *Service) GetAll() ([]Note, error) {
	notes, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}
	return notes, nil
}

func (s *Service) GetById(id int) (Note, error) {
	var note Note
	note, err := s.repo.GetById(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Note{}, err
		}
		return Note{}, err
	}
	return note, nil
}

func (s *Service) DeleteById(id int) error {
	err := s.repo.DeleteById(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return err
		}
		return err
	}
	return nil
}

func (s *Service) UpdateById(id int, text string) (Note, error) {

	if text == "" {
		return Note{}, errors.New("Text is required")
	}

	note, err := s.repo.UpdateById(id, text)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Note{}, sql.ErrNoRows
		}
		return Note{}, err
	}

	return note, nil
}

// Hander Layer

type Handler struct {
	service *Service
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Text string `json:"text"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	note, err := h.service.Create(req.Text)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var resp struct {
		Status string `json:"status"`
		Note   Note   `json:"note"`
	}

	resp.Note = note
	resp.Status = "created"

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(&resp)
	if err != nil {
		http.Error(w, "UnCorrect request", http.StatusBadRequest)
		return
	}
}

func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "method is not allowed", http.StatusMethodNotAllowed)
		return
	}

	var resp struct {
		Notes []Note `json:"notes"`
	}

	notesGet, err := h.service.GetAll()
	if err != nil {
		http.Error(w, "Service Data Base error", http.StatusServiceUnavailable)
		return
	}

	resp.Notes = notesGet

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(&resp)
	if err != nil {
		http.Error(w, "encode error", http.StatusInternalServerError)
		return
	}

}

func (h *Handler) GetById(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method is not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	note, err := h.service.GetById(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "note is not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	var resp struct {
		Note Note `json:"note"`
	}

	resp.Note = note

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(&resp)
	if err != nil {
		http.Error(w, "encode error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) DeleteById(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method is not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	err = h.service.DeleteById(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "note is note found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) UpdateById(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	var req struct {
		Text string `json:"text"`
	}

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	note, err := h.service.UpdateById(id, req.Text)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "note not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(&note)
	if err != nil {
		http.Error(w, "encode error", http.StatusInternalServerError)
		return
	}
}

// Health

func Health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var resp struct {
		Status string `json:"status"`
	}

	resp.Status = "ok"

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, "encode error", http.StatusInternalServerError)
		return
	}

}

func Ready(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		err := db.PingContext(ctx)
		if err != nil {
			var resp struct {
				Status string `json:"status"`
			}

			resp.Status = "not ready"

			w.Header().Set("Content-type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(resp)
			return
		}

		var resp struct {
			Status string `json:"status"`
		}

		resp.Status = "ready"

		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			http.Error(w, "encode error", http.StatusInternalServerError)
			return
		}
	}
}

// Realisation
func NewInMemoryRepository() *InMemoryNoteRepository {
	return &InMemoryNoteRepository{
		notes:  []Note{},
		nextID: 1,
	}
}

func NewPostgresNoteRepository(db *sql.DB) *PostgresNoteRepository {
	return &PostgresNoteRepository{
		db: db,
	}
}

func NewService(repo NoteRepository) *Service {
	return &Service{
		repo: repo,
	}

}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// Config
type Config struct {
	HTTPport string
	DBDSN    string
}

func LoadConfig() Config {
	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = ":8080"
	}

	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("DB_DSN is required")
	}

	return Config{
		HTTPport: port,
		DBDSN:    dsn,
	}
}

func main() {
	cfg := LoadConfig()

	db, err := sql.Open("pgx", cfg.DBDSN)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	repo := NewPostgresNoteRepository(db)
	service := NewService(repo)
	handler := NewHandler(service)

	// Router
	mux := http.NewServeMux()

	mux.HandleFunc("/health", Health)
	mux.HandleFunc("/health/ready", Ready(db))
	mux.HandleFunc("/notes", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			id := r.URL.Query().Get("id")
			if id != "" {
				handler.GetById(w, r)
				return
			}
			handler.GetAll(w, r)
		case http.MethodPost:
			handler.Create(w, r)
		case http.MethodDelete:
			id := r.URL.Query().Get("id")
			if id == "" {
				http.Error(w, "missing id", http.StatusBadRequest)
				return
			}
			handler.DeleteById(w, r)
		case http.MethodPut:
			handler.UpdateById(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Server
	addr := ":" + cfg.HTTPport

	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		log.Printf("notes-service starting on %s", addr)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	log.Println("shutdown signal received")

	// graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}

	log.Println("server stopped gracefully")

}
