package domain

type Account struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Token string `json:"token"`
}

type Photo struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Caption       string `json:"caption,omitempty"`
	ContentBase64 string `json:"content_base64"`
}

type Room struct {
	ID      string  `json:"id"`
	Title   string  `json:"title"`
	OwnerID string  `json:"owner_id"`
	Photos  []Photo `json:"photos"`
}

type ExportResult struct {
	ID            string `json:"id"`
	RoomID        string `json:"room_id"`
	PhotoCount    int    `json:"photo_count"`
	ContentBase64 string `json:"content_base64"`
}

type State struct {
	NextAccountID int            `json:"next_account_id"`
	NextRoomID    int            `json:"next_room_id"`
	NextPhotoID   int            `json:"next_photo_id"`
	NextExportID  int            `json:"next_export_id"`
	Accounts      []Account      `json:"accounts"`
	Rooms         []Room         `json:"rooms"`
	Exports       []ExportResult `json:"exports"`
}

func (s State) Clone() State {
	copyState := s
	copyState.Accounts = append([]Account(nil), s.Accounts...)
	copyState.Rooms = make([]Room, len(s.Rooms))
	for i, room := range s.Rooms {
		copyState.Rooms[i] = room
		copyState.Rooms[i].Photos = append([]Photo(nil), room.Photos...)
	}
	copyState.Exports = append([]ExportResult(nil), s.Exports...)
	return copyState
}
