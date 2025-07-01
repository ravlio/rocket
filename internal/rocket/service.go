package rocket

import (
	"context"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"sort"
	"strings"
)

// Service - interface for rocket service
type Service interface {
	// ProcessMessage processes a telemetry message and updates the rocket state accordingly
	ProcessMessage(ctx context.Context, msg TelemetryMessage) error
	// GetRocketState retrieves the current state of a rocket by its ID
	GetRocketState(ctx context.Context, id uuid.UUID) (State, bool)
	// ListAllRockets lists all rockets, optionally sorted by a specified field and order
	ListAllRockets(ctx context.Context, sortBy, sortOrder string) []State
}

var _ Service = (*ServiceImpl)(nil)

// ServiceImpl - implementation of the rocket service
type ServiceImpl struct {
	store  Store
	logger *zap.Logger
}

// NewRocketService creates a new instance of the rocket service with the provided store and logger.
func NewRocketService(store Store, logger *zap.Logger) *ServiceImpl {
	return &ServiceImpl{
		store:  store,
		logger: logger,
	}
}

// ProcessMessage processes a telemetry message and updates the rocket state accordingly
func (s *ServiceImpl) ProcessMessage(_ context.Context, msg TelemetryMessage) error {
	s.logger.Info(
		"Processing message",
		zap.String("channel", msg.Metadata.Channel.String()),
		zap.String("type", string(msg.Metadata.MessageType)),
		zap.Int64("number", msg.Metadata.MessageNumber),
	)

	rocketID := msg.Metadata.Channel
	currentState, exists := s.store.GetRocketByID(rocketID)

	// Check if the message is old or a duplicate
	if exists && msg.Metadata.MessageNumber <= currentState.LastProcessedMessageNumber {
		s.logger.Warn("Ignoring old or duplicate message",
			zap.String("rocket_id", rocketID.String()),
			zap.Int64("current_num", currentState.LastProcessedMessageNumber),
			zap.Int64("msg_num", msg.Metadata.MessageNumber),
		)
		return nil
	}

	newState := currentState
	if !exists {
		s.logger.Info("New rocket detected", zap.String("id", rocketID.String()))
		newState = State{
			ID:     rocketID,
			Status: "UNKNOWN",
		}
	}

	newState.LastProcessedMessageNumber = msg.Metadata.MessageNumber
	newState.LastUpdateTime = msg.Metadata.MessageTime

	switch msg.Metadata.MessageType {
	case MessageTypeLaunched:
		newState.Type = *msg.Message.Type
		newState.CurrentSpeed = *msg.Message.LaunchSpeed
		newState.Mission = *msg.Message.Mission
		newState.Status = StatusLaunched
	case MessageTypeSpeedIncreased:
		newState.CurrentSpeed += *msg.Message.By
	case MessageTypeSpeedDecreased:
		newState.CurrentSpeed -= *msg.Message.By
	case MessageTypeExploded:
		newState.CurrentSpeed = 0
		newState.Status = StatusExploded
		newState.Reason = msg.Message.Reason
	case MessageTypeMissionChanged:
		newState.Mission = *msg.Message.NewMission
	}

	s.store.SaveRocket(newState)
	s.logger.Info(
		"Rocket state updated successfully",
		zap.String("rocket_id", rocketID.String()),
		zap.Any("new_speed", newState.CurrentSpeed),
		zap.String("new_status", string(newState.Status)),
	)
	return nil
}

// GetRocketState retrieves the current state of a rocket by its ID
func (s *ServiceImpl) GetRocketState(_ context.Context, id uuid.UUID) (State, bool) {
	return s.store.GetRocketByID(id)
}

// ListAllRockets lists all rockets, optionally sorted by a specified field and order
func (s *ServiceImpl) ListAllRockets(_ context.Context, sortBy, sortOrder string) []State {
	rockets := s.store.ListAllRockets()

	if sortBy == "" {
		return rockets // No sorting
	}

	sort.Slice(rockets, func(i, j int) bool {
		var less bool
		switch strings.ToLower(sortBy) {
		case "id":
			less = rockets[i].ID.String() < rockets[j].ID.String()
		case "type":
			less = rockets[i].Type < rockets[j].Type
		case "speed":
			less = rockets[i].CurrentSpeed < rockets[j].CurrentSpeed
		case "mission":
			less = rockets[i].Mission < rockets[j].Mission
		case "lastupdatetime":
			less = rockets[i].LastUpdateTime.Before(rockets[j].LastUpdateTime)
		default:
			return i < j // Stable sort if unknown sort by field
		}

		if strings.ToLower(sortOrder) == "desc" {
			return !less
		}
		return less
	})

	return rockets
}
