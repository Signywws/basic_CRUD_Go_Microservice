package repository

import (
	"database/sql"
	"errors"
	model "notes_service/internal/models"
)

// constructor

func NewInMemoryRepository() *InMemoryNoteRepository {
	return &InMemoryNoteRepository{
		notes:  []model.Note{},
		nextID: 1,
	}
}

func NewPostgresNoteRepository(db *sql.DB) *PostgresNoteRepository {
	return &PostgresNoteRepository{
		db: db,
	}
}

// Repository layer

type NoteRepository interface {
	Create(note model.Note) (model.Note, error)
	GetAll() ([]model.Note, error)
	GetById(id int) (model.Note, error)
	DeleteById(id int) error
	UpdateById(id int, text string) (model.Note, error)
}

type InMemoryNoteRepository struct {
	notes  []model.Note
	nextID int
}

func (r *InMemoryNoteRepository) Create(note model.Note) (model.Note, error) {
	note.ID = r.nextID
	r.nextID++

	r.notes = append(r.notes, note)
	return note, nil
}

func (r *InMemoryNoteRepository) GetAll() ([]model.Note, error) {
	return r.notes, nil
}

type PostgresNoteRepository struct {
	db *sql.DB
}

func (p *PostgresNoteRepository) Create(note model.Note) (model.Note, error) {
	query := "INSERT INTO notes (text) VALUES ($1) RETURNING id"

	err := p.db.QueryRow(query, note.Text).Scan(&note.ID)
	if err != nil {
		return model.Note{}, err
	}

	return note, nil
}

func (p *PostgresNoteRepository) GetAll() ([]model.Note, error) {
	rows, err := p.db.Query("SELECT id, text FROM notes")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var notes []model.Note

	for rows.Next() {
		var note model.Note
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

func (p *PostgresNoteRepository) GetById(id int) (model.Note, error) {
	var note model.Note
	err := p.db.QueryRow(
		"SELECT id, text FROM notes WHERE id = $1", id).Scan(&note.ID, &note.Text)
	if err != nil {
		return model.Note{}, err
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

func (p *PostgresNoteRepository) UpdateById(id int, text string) (model.Note, error) {

	var note model.Note
	query := "UPDATE notes SET text = $1 WHERE id = $2 RETURNING id, text"
	err := p.db.QueryRow(query, text, id).Scan(&note.ID, &note.Text)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Note{}, sql.ErrNoRows
		}
		return model.Note{}, err
	}
	return note, nil
}
