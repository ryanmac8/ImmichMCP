package upload

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// Status represents the state of an upload session.
type Status int

const (
	StatusPending   Status = iota
	StatusUploading
	StatusCompleted
	StatusFailed
	StatusExpired
)

func (s Status) String() string {
	switch s {
	case StatusPending:
		return "pending"
	case StatusUploading:
		return "uploading"
	case StatusCompleted:
		return "completed"
	case StatusFailed:
		return "failed"
	case StatusExpired:
		return "expired"
	default:
		return "unknown"
	}
}

// Session represents an upload session.
type Session struct {
	SessionID  string
	CreatedAt  time.Time
	ExpiresAt  time.Time
	FileName   *string
	IsFavorite *bool
	IsArchived *bool
	Status     Status
	AssetID    *string
	Error      *string
}

// SessionService manages upload sessions.
type SessionService struct {
	mu       sync.RWMutex
	sessions map[string]*Session
	timeout  time.Duration
	stop     chan struct{}
}

// NewSessionService creates a new upload session manager.
func NewSessionService(timeout time.Duration) *SessionService {
	if timeout == 0 {
		timeout = 30 * time.Minute
	}
	svc := &SessionService{
		sessions: make(map[string]*Session),
		timeout:  timeout,
		stop:     make(chan struct{}),
	}
	go svc.cleanupLoop()
	return svc
}

// CreateSession creates a new upload session and returns it.
func (s *SessionService) CreateSession(fileName *string, isFavorite, isArchived *bool) *Session {
	sess := &Session{
		SessionID:  uuid.New().String(),
		CreatedAt:  time.Now().UTC(),
		ExpiresAt:  time.Now().UTC().Add(s.timeout),
		FileName:   fileName,
		IsFavorite: isFavorite,
		IsArchived: isArchived,
		Status:     StatusPending,
	}
	s.mu.Lock()
	s.sessions[sess.SessionID] = sess
	s.mu.Unlock()
	return sess
}

// GetSession retrieves a session by ID. Returns nil if not found.
func (s *SessionService) GetSession(id string) *Session {
	s.mu.RLock()
	sess, ok := s.sessions[id]
	s.mu.RUnlock()
	if !ok {
		return nil
	}
	if time.Now().UTC().After(sess.ExpiresAt) && sess.Status == StatusPending {
		s.mu.Lock()
		sess.Status = StatusExpired
		s.mu.Unlock()
	}
	return sess
}

// SetStatus updates the status of a session.
func (s *SessionService) SetStatus(id string, status Status) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if sess, ok := s.sessions[id]; ok {
		sess.Status = status
	}
}

// SetCompleted marks a session as completed with the resulting asset ID.
func (s *SessionService) SetCompleted(id, assetID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if sess, ok := s.sessions[id]; ok {
		sess.Status = StatusCompleted
		sess.AssetID = &assetID
	}
}

// SetFailed marks a session as failed with an error message.
func (s *SessionService) SetFailed(id, errMsg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if sess, ok := s.sessions[id]; ok {
		sess.Status = StatusFailed
		sess.Error = &errMsg
	}
}

// Close stops the background cleanup goroutine.
func (s *SessionService) Close() {
	close(s.stop)
}

func (s *SessionService) cleanupLoop() {
	t := time.NewTicker(time.Minute)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			s.cleanup()
		case <-s.stop:
			return
		}
	}
}

func (s *SessionService) cleanup() {
	now := time.Now().UTC()
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, sess := range s.sessions {
		if now.After(sess.ExpiresAt) {
			delete(s.sessions, id)
		}
	}
}
