package rocket

import (
	"github.com/google/uuid"
	"go.uber.org/zap"
	"sync"
)

// Store - interface for rocket state storage
type Store interface {
	// SaveRocket saves the current state of a rocket
	SaveRocket(state State)
	// GetRocketByID retrieves the state of a rocket by its ID
	GetRocketByID(id uuid.UUID) (State, bool)
	// ListAllRockets lists all rockets in the store
	ListAllRockets() []State
}

var _ Store = (*InMemoryRocketStore)(nil)

type InMemoryRocketStore struct {
	mu      sync.RWMutex
	rockets map[uuid.UUID]State
	logger  *zap.Logger
}

// NewInMemoryRocketStore creates a new instance of InMemoryRocketStore with an initialized map for storing rocket states.
func NewInMemoryRocketStore(logger *zap.Logger) *InMemoryRocketStore {
	return &InMemoryRocketStore{
		rockets: make(map[uuid.UUID]State),
		logger:  logger,
	}
}

// SaveRocket saves the current state of a rocket
func (s *InMemoryRocketStore) SaveRocket(state State) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rockets[state.ID] = state
	s.logger.Info("Rocket state saved", zap.String("rocket_id", state.ID.String()), zap.Any("state", state))
}

// GetRocketByID retrieves the state of a rocket by its ID
func (s *InMemoryRocketStore) GetRocketByID(id uuid.UUID) (State, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rocket, ok := s.rockets[id]
	return rocket, ok
}

// ListAllRockets lists all rockets in the store
func (s *InMemoryRocketStore) ListAllRockets() []State {
	s.mu.RLock()
	defer s.mu.RUnlock()
	states := make([]State, 0, len(s.rockets))
	for _, rocket := range s.rockets {
		states = append(states, rocket)
	}
	return states
}
