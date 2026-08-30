package security

import "errors"

var (
	ErrRoomAccessDenied = errors.New("room access denied")
	ErrRoomFull         = errors.New("room is full")
)

func CanJoinRoom(roomCode string, userCount int) error {
	if ValidateRoomCode(roomCode) != nil {
		return ErrRoomAccessDenied
	}

	if userCount < 0 || userCount >= MaxRoomUsers {
		return ErrRoomFull
	}

	return nil
}

func CanLeaveRoom(userID string) bool {
	return userID != ""
}
