package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"percentman/models"

	"github.com/google/uuid"
)

const (
	maxHistoryItems = 50
	appDirName      = ".gopostman"
	templatesFile   = "templates.json"
	historyFile     = "history.json"
)

// Storage handles persistence of templates and history
type Storage struct {
	mu        sync.RWMutex
	templates []models.Template
	history   []models.HistoryItem
	dataDir   string
}

// NewStorage creates a new storage instance
func NewStorage() (*Storage, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	dataDir := filepath.Join(homeDir, appDirName)
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	s := &Storage{
		dataDir:   dataDir,
		templates: []models.Template{},
		history:   []models.HistoryItem{},
	}

	// Load existing data
	s.loadTemplates()
	s.loadHistory()

	return s, nil
}

// Templates

func (s *Storage) loadTemplates() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(filepath.Join(s.dataDir, templatesFile))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	return json.Unmarshal(data, &s.templates)
}

func (s *Storage) saveTemplates() error {
	data, err := json.MarshalIndent(s.templates, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(s.dataDir, templatesFile), data, 0644)
}

// GetTemplates returns all templates
func (s *Storage) GetTemplates() []models.Template {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]models.Template, len(s.templates))
	copy(result, s.templates)
	return result
}

// SaveTemplate saves a new template or updates existing one
func (s *Storage) SaveTemplate(name string, req *models.Request) (*models.Template, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	// Check if template with same name exists
	for i, t := range s.templates {
		if t.Name == name {
			s.templates[i].Request = *req.Clone()
			s.templates[i].UpdatedAt = now
			if err := s.saveTemplates(); err != nil {
				return nil, err
			}
			return &s.templates[i], nil
		}
	}

	// Create new template
	template := models.Template{
		ID:        uuid.New().String(),
		Name:      name,
		Request:   *req.Clone(),
		CreatedAt: now,
		UpdatedAt: now,
	}

	s.templates = append(s.templates, template)

	// Sort by name
	sort.Slice(s.templates, func(i, j int) bool {
		return s.templates[i].Name < s.templates[j].Name
	})

	if err := s.saveTemplates(); err != nil {
		return nil, err
	}

	return &template, nil
}

// DeleteTemplate deletes a template by ID
func (s *Storage) DeleteTemplate(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, t := range s.templates {
		if t.ID == id {
			s.templates = append(s.templates[:i], s.templates[i+1:]...)
			return s.saveTemplates()
		}
	}
	return nil
}

// GetTemplateByID returns a template by ID
func (s *Storage) GetTemplateByID(id string) *models.Template {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, t := range s.templates {
		if t.ID == id {
			return &t
		}
	}
	return nil
}

// TemplateNameExists checks if a template name already exists
func (s *Storage) TemplateNameExists(name string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, t := range s.templates {
		if t.Name == name {
			return true
		}
	}
	return false
}

// History

func (s *Storage) loadHistory() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(filepath.Join(s.dataDir, historyFile))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	return json.Unmarshal(data, &s.history)
}

func (s *Storage) saveHistory() error {
	data, err := json.MarshalIndent(s.history, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(s.dataDir, historyFile), data, 0644)
}

// GetHistory returns all history items
func (s *Storage) GetHistory() []models.HistoryItem {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]models.HistoryItem, len(s.history))
	copy(result, s.history)
	return result
}

// AddHistory adds a new history item
func (s *Storage) AddHistory(req *models.Request, resp *models.Response) (*models.HistoryItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item := models.HistoryItem{
		ID:        uuid.New().String(),
		Request:   *req.Clone(),
		Response:  *resp,
		Timestamp: time.Now(),
	}

	// Prepend to history (newest first)
	s.history = append([]models.HistoryItem{item}, s.history...)

	// Limit history size
	if len(s.history) > maxHistoryItems {
		s.history = s.history[:maxHistoryItems]
	}

	if err := s.saveHistory(); err != nil {
		return nil, err
	}

	return &item, nil
}

// ClearHistory removes all history items
func (s *Storage) ClearHistory() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.history = []models.HistoryItem{}
	return s.saveHistory()
}

// GetHistoryByID returns a history item by ID
func (s *Storage) GetHistoryByID(id string) *models.HistoryItem {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, h := range s.history {
		if h.ID == id {
			return &h
		}
	}
	return nil
}
