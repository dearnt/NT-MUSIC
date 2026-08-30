package main

type ClientMessage struct {
	Type string            `json:"type"`
	Data ClientMessageData `json:"data"`
}

type ClientMessageData struct {
	UserID   string  `json:"user_id,omitempty"`
	Name     string  `json:"name,omitempty"`
	RoomCode string  `json:"room_code,omitempty"`
	URL      string  `json:"url,omitempty"`
	Position float64 `json:"position,omitempty"`
}

type ServerMessage struct {
	Type string `json:"type"`
	Data any    `json:"data,omitempty"`
}

type RoomResponse struct {
	ID        string `json:"id"`
	Code      string `json:"code"`
	OwnerID   string `json:"owner_id"`
	UserCount int    `json:"user_count"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

type PlaybackState struct {
	SongID   string  `json:"song_id"`
	Playing  bool    `json:"playing"`
	Position float64 `json:"position"`
	Updated  int64   `json:"updated"`
}
