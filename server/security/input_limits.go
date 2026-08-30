package security

const (
	MaxNameLength      = 32
	MaxRoomCodeLength  = 16
	MaxSongTitleLength = 200
	MaxSongURLLength   = 2048
	MaxMessageSize     = 64 * 1024
	MaxQueueSize       = 500
	MaxRoomUsers       = 100
	MaxRoomsPerRequest = 10
)

func ValidLength(value string, max int) bool {
	return len(value) > 0 && len(value) <= max
}

func ValidMessageSize(size int64) bool {
	return size > 0 && size <= MaxMessageSize
}

func ValidQueueSize(size int) bool {
	return size >= 0 && size <= MaxQueueSize
}

func ValidRoomUsers(size int) bool {
	return size >= 0 && size <= MaxRoomUsers
}
