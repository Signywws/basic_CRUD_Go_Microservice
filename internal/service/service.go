package service

import (
	"database/sql"
	"errors"
	model "notes_service/internal/models"
	service "notes_service/internal/repository"
)

type Service struct {
	repo service.NoteRepository
}

func (s *Service) Create(text string) (model.Note, error) {
	note := model.Note{
		Text: text,
	}

	if text == "" {
		return model.Note{}, errors.New("Text is required")
	}

	note, err := s.repo.Create(note)
	if err != nil {
		return model.Note{}, err
	}

	return note, nil

}
func (s *Service) GetAll() ([]model.Note, error) {
	notes, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}
	return notes, nil
}

func (s *Service) GetById(id int) (model.Note, error) {
	var note model.Note
	note, err := s.repo.GetById(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Note{}, err
		}
		return model.Note{}, err
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

func (s *Service) UpdateById(id int, text string) (model.Note, error) {

	if text == "" {
		return model.Note{}, errors.New("Text is required")
	}

	note, err := s.repo.UpdateById(id, text)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Note{}, sql.ErrNoRows
		}
		return model.Note{}, err
	}

	return note, nil
}

func NewService(repo service.NoteRepository) *Service {
	return &Service{
		repo: repo,
	}

}
