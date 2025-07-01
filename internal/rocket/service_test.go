package rocket

import (
	"context"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"reflect"
	"sort"
	"testing"
	"time"
)

func ptr[T any](v T) *T {
	return &v
}

func TestRocketService_ProcessMessage_RocketLaunched_Integration(t *testing.T) {
	logger := zap.NewNop()
	store := NewInMemoryRocketStore(logger) // Use actual store
	service := NewRocketService(store, logger)

	rocketID := uuid.New()
	launchTime := time.Now().UTC().Truncate(time.Millisecond)

	msg := TelemetryMessage{
		Metadata: MessageMetadata{
			Channel:       rocketID,
			MessageNumber: 1,
			MessageTime:   launchTime,
			MessageType:   MessageTypeLaunched,
		},
		Message: Message{
			Type:        ptr("IntegrationRocket"),
			LaunchSpeed: ptr(int64(1000)),
			Mission:     ptr("IntegrationLaunch"),
		},
	}

	err := service.ProcessMessage(context.Background(), msg)
	if err != nil {
		t.Fatalf("ProcessMessage failed: %v", err)
	}

	state, ok := store.GetRocketByID(rocketID)
	if !ok {
		t.Fatalf("Rocket %s not found in store after launch message", rocketID.String())
	}

	expectedState := State{
		ID:                         rocketID,
		Type:                       "IntegrationRocket",
		CurrentSpeed:               1000,
		Mission:                    "IntegrationLaunch",
		Status:                     StatusLaunched,
		Reason:                     nil,
		LastUpdateTime:             launchTime,
		LastProcessedMessageNumber: 1,
	}

	if !reflect.DeepEqual(state, expectedState) {
		t.Errorf("Rocket state mismatch after launch.\nExpected: %+v\nGot: %+v", expectedState, state)
	}
}

func TestRocketService_ListAllRockets_Sorting_Integration(t *testing.T) {
	logger := zap.NewNop()
	store := NewInMemoryRocketStore(logger) // Use actual store
	service := NewRocketService(store, logger)

	// Save some rockets directly to the store for testing list/sort
	t1 := time.Now().UTC().Add(-3 * time.Hour).Truncate(time.Millisecond)
	t2 := time.Now().UTC().Add(-2 * time.Hour).Truncate(time.Millisecond)
	t3 := time.Now().UTC().Add(-1 * time.Hour).Truncate(time.Millisecond)

	r1 := State{ID: uuid.New(), Type: "Alpha", CurrentSpeed: 300, Mission: "Z", LastUpdateTime: t3, LastProcessedMessageNumber: 3}
	r2 := State{ID: uuid.New(), Type: "Beta", CurrentSpeed: 100, Mission: "X", LastUpdateTime: t1, LastProcessedMessageNumber: 1}
	r3 := State{ID: uuid.New(), Type: "Gamma", CurrentSpeed: 200, Mission: "Y", LastUpdateTime: t2, LastProcessedMessageNumber: 2}

	store.SaveRocket(r1)
	store.SaveRocket(r2)
	store.SaveRocket(r3)

	// Test: No sorting
	rockets := service.ListAllRockets(context.Background(), "", "")
	if len(rockets) != 3 {
		t.Fatalf("Expected 3 rockets, got %d", len(rockets))
	}

	// Test: Sort by ID (asc)
	rockets = service.ListAllRockets(context.Background(), "id", "asc")
	// The exact UUIDs are random, so we need to sort the expected slice as well for direct comparison
	expectedIDsSorted := []string{r1.ID.String(), r2.ID.String(), r3.ID.String()}
	sort.Strings(expectedIDsSorted) // Sort UUID strings

	if rockets[0].ID.String() != expectedIDsSorted[0] ||
		rockets[1].ID.String() != expectedIDsSorted[1] ||
		rockets[2].ID.String() != expectedIDsSorted[2] {
		t.Errorf("Sorting by ID ASC failed. Expected IDs: %v. Got %s, %s, %s", expectedIDsSorted, rockets[0].ID.String(), rockets[1].ID.String(), rockets[2].ID.String())
	}

	// Test: Sort by Speed (desc)
	rockets = service.ListAllRockets(context.Background(), "speed", "desc")
	if rockets[0].ID != r1.ID || rockets[1].ID != r3.ID || rockets[2].ID != r2.ID { // C (300), B (200), A (100)
		t.Errorf("Sorting by Speed DESC failed. Expected C, B, A. Got %s, %s, %s", rockets[0].ID.String(), rockets[1].ID.String(), rockets[2].ID.String())
	}

	// Test: Sort by LastUpdateTime (asc)
	rockets = service.ListAllRockets(context.Background(), "lastupdatetime", "asc")
	if rockets[0].ID != r2.ID || rockets[1].ID != r3.ID || rockets[2].ID != r1.ID { // A (t1), B (t2), C (t3)
		t.Errorf("Sorting by LastUpdateTime ASC failed. Expected A, B, C. Got %s, %s, %s", rockets[0].ID.String(), rockets[1].ID.String(), rockets[2].ID.String())
	}
}
