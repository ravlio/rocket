package rocket

import (
	"github.com/google/uuid"
	"go.uber.org/zap"
	"reflect"
	"testing"
	"time"
)

func TestInMemoryRocketStore_SaveAndGet(t *testing.T) {
	logger := zap.NewNop()
	store := NewInMemoryRocketStore(logger)

	rocketID := uuid.New()
	launchTime := time.Now().UTC().Truncate(time.Millisecond)
	initialState := State{
		ID:                         rocketID,
		Type:                       "Falcon-9",
		CurrentSpeed:               500,
		Mission:                    "ARTEMIS",
		Status:                     "LAUNCHED",
		LastUpdateTime:             launchTime,
		LastProcessedMessageNumber: 1,
	}

	store.SaveRocket(initialState)

	retrievedState, ok := store.GetRocketByID(rocketID)
	if !ok {
		t.Fatalf("Expected rocket with ID %s to be found, but it was not", rocketID)
	}
	if !reflect.DeepEqual(retrievedState, initialState) {
		t.Errorf("Retrieved state does not match initial state.\nExpected: %+v\nGot: %+v", initialState, retrievedState)
	}

	nonExistentID := uuid.New()
	_, ok = store.GetRocketByID(nonExistentID)
	if ok {
		t.Errorf("Expected rocket with ID %s not to be found, but it was", nonExistentID)
	}

	updatedSpeed := int64(3500)
	updatedMessageNumber := int64(2)
	updatedTime := time.Now().UTC().Add(time.Minute).Truncate(time.Millisecond)
	updatedState := State{
		ID:                         rocketID,
		Type:                       "Falcon-9",
		CurrentSpeed:               updatedSpeed,
		Mission:                    "ARTEMIS",
		Status:                     "IN_FLIGHT",
		LastUpdateTime:             updatedTime,
		LastProcessedMessageNumber: updatedMessageNumber,
	}
	store.SaveRocket(updatedState)

	retrievedUpdatedState, ok := store.GetRocketByID(rocketID)
	if !ok {
		t.Fatalf("Expected updated rocket with ID %s to be found, but it was not", rocketID)
	}
	if !reflect.DeepEqual(retrievedUpdatedState, updatedState) {
		t.Errorf("Retrieved updated state does not match updated state.\nExpected: %+v\nGot: %+v", updatedState, retrievedUpdatedState)
	}
	if retrievedUpdatedState.CurrentSpeed != updatedSpeed {
		t.Errorf("Expected rocket speed to be %d, got %d", updatedSpeed, retrievedUpdatedState.CurrentSpeed)
	}
	if retrievedUpdatedState.LastProcessedMessageNumber != updatedMessageNumber {
		t.Errorf("Expected last message number to be %d, got %d", updatedMessageNumber, retrievedUpdatedState.LastProcessedMessageNumber)
	}
}

func TestInMemoryRocketStore_ListAllRockets(t *testing.T) {
	logger := zap.NewNop()
	store := NewInMemoryRocketStore(logger)

	rockets := store.ListAllRockets()
	if len(rockets) != 0 {
		t.Errorf("Expected 0 rockets in an empty store, got %d", len(rockets))
	}

	rocketID1 := uuid.New()
	rocketID2 := uuid.New()
	rocketID3 := uuid.New()

	state1 := State{ID: rocketID1, Type: "Falcon-9", CurrentSpeed: 100, LastUpdateTime: time.Now().UTC().Add(-2 * time.Hour)}
	state2 := State{ID: rocketID2, Type: "Soyuz", CurrentSpeed: 200, LastUpdateTime: time.Now().UTC().Add(-1 * time.Hour)}
	state3 := State{ID: rocketID3, Type: "Starship", CurrentSpeed: 300, LastUpdateTime: time.Now().UTC()}

	store.SaveRocket(state1)
	store.SaveRocket(state2)
	store.SaveRocket(state3)

	rockets = store.ListAllRockets()
	if len(rockets) != 3 {
		t.Errorf("Expected 3 rockets, got %d", len(rockets))
	}

	foundIDs := make(map[uuid.UUID]bool)
	for _, r := range rockets {
		foundIDs[r.ID] = true
	}
	if !foundIDs[rocketID1] || !foundIDs[rocketID2] || !foundIDs[rocketID3] {
		t.Errorf("Not all saved rockets were found in the list. Found IDs: %+v", foundIDs)
	}

	updatedState1 := State{ID: rocketID1, Type: "Falcon-9", CurrentSpeed: 150, LastUpdateTime: time.Now().UTC().Add(time.Minute)}
	store.SaveRocket(updatedState1)
	rockets = store.ListAllRockets()
	if len(rockets) != 3 {
		t.Errorf("Expected 3 rockets after update, got %d", len(rockets))
	}
}

func TestInMemoryRocketStore_Concurrency(t *testing.T) {
	logger := zap.NewNop()
	store := NewInMemoryRocketStore(logger)
	rocketID := uuid.New()

	go func() {
		for i := 0; i < 100; i++ {
			state := State{
				ID:                         rocketID,
				CurrentSpeed:               int64(i),
				LastUpdateTime:             time.Now(),
				LastProcessedMessageNumber: int64(i),
			}
			store.SaveRocket(state)
		}
	}()

	for i := 0; i < 100; i++ {
		_, _ = store.GetRocketByID(rocketID)
		_ = store.ListAllRockets()
	}

	time.Sleep(50 * time.Millisecond)
}
