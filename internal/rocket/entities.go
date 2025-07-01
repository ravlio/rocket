package rocket

import (
	"github.com/google/uuid"
	"time"
)

// Status - rocket status
type Status string

const (
	StatusLaunched Status = "LAUNCHED"
	StatusExploded Status = "EXPLODED"
)

// MessageType - telemetry message type
type MessageType string

const (
	MessageTypeExploded       MessageType = "RocketExploded"
	MessageTypeLaunched       MessageType = "RocketLaunched"
	MessageTypeMissionChanged MessageType = "RocketMissionChanged"
	MessageTypeSpeedDecreased MessageType = "RocketSpeedDecreased"
	MessageTypeSpeedIncreased MessageType = "RocketSpeedIncreased"
)

// State - rocket state
type State struct {
	ID                         uuid.UUID `json:"id"`
	Type                       string    `json:"type"`
	CurrentSpeed               int64     `json:"currentSpeed"`
	Mission                    string    `json:"mission"`
	Status                     Status    `json:"status"`
	Reason                     *string   `json:"reason,omitempty"`
	LastUpdateTime             time.Time `json:"lastUpdateTime"`
	LastProcessedMessageNumber int64     `json:"lastProcessedMessageNumber"`
}

// MessageMetadata - metadata for telemetry messages
type MessageMetadata struct {
	Channel       uuid.UUID   `json:"channel"`
	MessageNumber int64       `json:"messageNumber"`
	MessageTime   time.Time   `json:"messageTime"`
	MessageType   MessageType `json:"messageType"`
}

// Message - structure for telemetry messages
type Message struct {
	By          *int64  `json:"by,omitempty"`
	LaunchSpeed *int64  `json:"launchSpeed,omitempty"`
	Mission     *string `json:"mission,omitempty"`
	NewMission  *string `json:"newMission,omitempty"`
	Reason      *string `json:"reason,omitempty"`
	Type        *string `json:"type,omitempty"`
}

// TelemetryMessage - structure for telemetry messages with metadata
type TelemetryMessage struct {
	Metadata MessageMetadata `json:"metadata"`
	Message  Message         `json:"message"`
}
