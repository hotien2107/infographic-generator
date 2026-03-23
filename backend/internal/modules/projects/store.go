package projects

import (
	"errors"
	"sync"
	"time"

	"infographic-generator/backend/internal/modules/documents"
	"infographic-generator/backend/internal/utils"
)

var ErrProjectNotFound = errors.New("project not found")

type Store struct {
	mu        sync.RWMutex
	projects  map[string]Project
	documents map[string][]documents.Document
}

func NewStore() *Store {
	return &Store{
		projects:  make(map[string]Project),
		documents: make(map[string][]documents.Document),
	}
}

func (s *Store) CreateProject(title string, inputMode InputMode) Project {
	now := time.Now().UTC()
	project := Project{
		ID:          utils.NewUUID(),
		Title:       title,
		InputMode:   inputMode,
		Status:      StatusDraft,
		CurrentStep: StepWaitingUpload,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.projects[project.ID] = project

	return project
}

func (s *Store) GetProject(projectID string) (Project, []documents.Document, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	project, ok := s.projects[projectID]
	if !ok {
		return Project{}, nil, ErrProjectNotFound
	}

	docs := append([]documents.Document(nil), s.documents[projectID]...)
	return project, docs, nil
}

func (s *Store) AddDocument(projectID string, document documents.Document) (Project, []documents.Document, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	project, ok := s.projects[projectID]
	if !ok {
		return Project{}, nil, ErrProjectNotFound
	}

	project.Status = StatusUploaded
	project.CurrentStep = StepUploaded
	project.UpdatedAt = time.Now().UTC()
	s.projects[projectID] = project
	s.documents[projectID] = append(s.documents[projectID], document)

	docs := append([]documents.Document(nil), s.documents[projectID]...)
	return project, docs, nil
}
