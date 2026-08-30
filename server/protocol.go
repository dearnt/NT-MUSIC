package main

type ClientMessage struct {
	Type string            `json:"type"`
	Data ClientMessageData `json:"data"`
}

type ClientMessageData struct {
	UserID       string  `json:"user_id,omitempty"`
	Name         string  `json:"name,omitempty"`
	RoomCode     string  `json:"room_code,omitempty"`
	URL          string  `json:"url,omitempty"`
	Volume       float64 `json:"volume,omitempty"`
	GlobalVolume float64 `json:"global_volume,omitempty"`
	TargetUserID string  `json:"target_user_id,omitempty"`
	Enabled      bool    `json:"enabled,omitempty"`
	Position     float64 `json:"position,omitempty"`
	SongID       string  `json:"song_id,omitempty"`
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
