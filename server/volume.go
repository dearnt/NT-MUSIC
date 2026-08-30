package main

import (
	"errors"
	"sync"
)

const (
	minVolume = 0.0
	maxVolume = 1.0
)

type RoomAudioState struct {
	mu          sync.RWMutex
	Volumes     map[string]float64
	Global      float64
	GlobalMuted bool
}

var roomAudioStates = struct {
	sync.Mutex
	items map[*Room]*RoomAudioState
}{
	items: make(map[*Room]*RoomAudioState),
}

func audioState(room *Room) *RoomAudioState {
	roomAudioStates.Lock()
	defer roomAudioStates.Unlock()

	state := roomAudioStates.items[room]
	if state == nil {
		state = &RoomAudioState{
			Volumes: make(map[string]float64),
			Global:  1,
		}
		roomAudioStates.items[room] = state
	}

	return state
}

func removeAudioState(room *Room) {
	roomAudioStates.Lock()
	delete(roomAudioStates.items, room)
	roomAudioStates.Unlock()
}

func clampVolume(value float64) float64 {
	if value < minVolume {
		return minVolume
	}
	if value > maxVolume {
		return maxVolume
	}
	return value
}

func setUserVolume(room *Room, userID string, value float64) error {
	if room == nil || userID == "" {
		return errors.New("invalid audio state")
	}

	state := audioState(room)

	state.mu.Lock()
	state.Volumes[userID] = clampVolume(value)
	state.mu.Unlock()

	return nil
}

func userVolume(room *Room, userID string) float64 {
	if room == nil || userID == "" {
		return 1
	}

	state := audioState(room)

	state.mu.RLock()
	defer state.mu.RUnlock()

	value, exists := state.Volumes[userID]
	if !exists {
		return 1
	}

	return value
}

func setGlobalVolume(room *Room, value float64) {
	state := audioState(room)

	state.mu.Lock()
	state.Global = clampVolume(value)
	state.mu.Unlock()
}

func globalVolume(room *Room) float64 {
	if room == nil {
		return 1
	}

	state := audioState(room)

	state.mu.RLock()
	defer state.mu.RUnlock()

	return state.Global
}

func setGlobalMute(room *Room, muted bool) {
	state := audioState(room)

	state.mu.Lock()
	state.GlobalMuted = muted
	state.mu.Unlock()
}

func globalMuted(room *Room) bool {
	if room == nil {
		return false
	}

	state := audioState(room)

	state.mu.RLock()
	defer state.mu.RUnlock()

	return state.GlobalMuted
}

func effectiveVolume(room *Room, userID string) float64 {
	if globalMuted(room) {
		return 0
	}

	return userVolume(room, userID) * globalVolume(room)
}

func audioStateSnapshot(room *Room) map[string]any {
	state := audioState(room)

	state.mu.RLock()
	defer state.mu.RUnlock()

	volumes := make(map[string]float64, len(state.Volumes))
	for id, value := range state.Volumes {
		volumes[id] = value
	}

	return map[string]any{
		"volumes":       volumes,
		"global_volume": state.Global,
		"global_muted":  state.GlobalMuted,
	}
}
