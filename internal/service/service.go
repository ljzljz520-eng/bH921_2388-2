package service

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"weddinglive/internal/domain"
	"weddinglive/internal/export"
	"weddinglive/internal/store"
)

var (
	ErrForbidden = errors.New("forbidden")
	ErrNotFound  = errors.New("not found")
	ErrInvalid   = errors.New("invalid input")
)

type Service struct {
	store      *store.JSONStore
	adminToken string
	builder    *export.Builder
}

func New(repository *store.JSONStore, adminToken string) *Service {
	return &Service{store: repository, adminToken: adminToken, builder: export.NewBuilder()}
}

func (s *Service) CreateAccount(adminToken, name string) (domain.Account, error) {
	if adminToken != s.adminToken {
		return domain.Account{}, ErrForbidden
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return domain.Account{}, ErrInvalid
	}

	var created domain.Account
	err := s.store.Update(func(state *domain.State) error {
		created = domain.Account{
			ID:    fmt.Sprintf("acct-%03d", state.NextAccountID),
			Name:  name,
			Token: fmt.Sprintf("photo-token-%03d", state.NextAccountID),
		}
		state.NextAccountID++
		state.Accounts = append(state.Accounts, created)
		return nil
	})
	return created, err
}

func (s *Service) CreateRoom(token, title string) (domain.Room, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return domain.Room{}, ErrInvalid
	}

	var created domain.Room
	err := s.store.Update(func(state *domain.State) error {
		account, ok := accountByToken(*state, token)
		if !ok {
			return ErrForbidden
		}
		created = domain.Room{
			ID:      fmt.Sprintf("room-%03d", state.NextRoomID),
			Title:   title,
			OwnerID: account.ID,
			Photos:  []domain.Photo{},
		}
		state.NextRoomID++
		state.Rooms = append(state.Rooms, created)
		return nil
	})
	return created, err
}

func (s *Service) ListRooms() []domain.Room {
	state := s.store.Snapshot()
	rooms := make([]domain.Room, len(state.Rooms))
	for i, room := range state.Rooms {
		rooms[i] = room
		rooms[i].Photos = nil
	}
	return rooms
}

func (s *Service) AdminRooms(adminToken string) ([]domain.Room, error) {
	if adminToken != s.adminToken {
		return nil, ErrForbidden
	}
	return s.store.Snapshot().Rooms, nil
}

func (s *Service) DeleteRoom(adminToken, roomID string) error {
	if adminToken != s.adminToken {
		return ErrForbidden
	}
	return s.store.Update(func(state *domain.State) error {
		index := roomIndex(*state, roomID)
		if index < 0 {
			return ErrNotFound
		}
		state.Rooms = append(state.Rooms[:index], state.Rooms[index+1:]...)
		kept := state.Exports[:0]
		for _, result := range state.Exports {
			if result.RoomID != roomID {
				kept = append(kept, result)
			}
		}
		state.Exports = kept
		return nil
	})
}

func (s *Service) AddPhoto(token, roomID, name, caption, contentBase64 string) (domain.Photo, error) {
	name = strings.TrimSpace(name)
	if name == "" || contentBase64 == "" {
		return domain.Photo{}, ErrInvalid
	}
	if _, err := base64.StdEncoding.DecodeString(contentBase64); err != nil {
		return domain.Photo{}, ErrInvalid
	}

	var created domain.Photo
	err := s.store.Update(func(state *domain.State) error {
		room, err := ownedRoom(state, token, roomID)
		if err != nil {
			return err
		}
		created = domain.Photo{
			ID:            fmt.Sprintf("photo-%03d", state.NextPhotoID),
			Name:          name,
			Caption:       strings.TrimSpace(caption),
			ContentBase64: contentBase64,
		}
		state.NextPhotoID++
		room.Photos = append(room.Photos, created)
		return nil
	})
	return created, err
}

func (s *Service) ListPhotos(roomID string) ([]domain.Photo, error) {
	state := s.store.Snapshot()
	index := roomIndex(state, roomID)
	if index < 0 {
		return nil, ErrNotFound
	}
	return append([]domain.Photo(nil), state.Rooms[index].Photos...), nil
}

func (s *Service) DeletePhoto(token, roomID, photoID string) error {
	return s.store.Update(func(state *domain.State) error {
		room, err := ownedRoom(state, token, roomID)
		if err != nil {
			return err
		}
		for i, photo := range room.Photos {
			if photo.ID == photoID {
				room.Photos = append(room.Photos[:i], room.Photos[i+1:]...)
				return nil
			}
		}
		return ErrNotFound
	})
}

func (s *Service) ExportRoom(ctx context.Context, token, roomID string, options export.Options) (domain.ExportResult, error) {
	state := s.store.Snapshot()
	room, err := ownedRoom(&state, token, roomID)
	if err != nil {
		return domain.ExportResult{}, err
	}

	content, err := s.builder.Build(ctx, room.ID, room.Photos, options)
	if err != nil {
		return domain.ExportResult{}, err
	}

	var result domain.ExportResult
	err = s.store.Update(func(state *domain.State) error {
		if _, err := ownedRoom(state, token, roomID); err != nil {
			return err
		}
		result = domain.ExportResult{
			ID:            fmt.Sprintf("export-%03d", state.NextExportID),
			RoomID:        roomID,
			PhotoCount:    len(room.Photos),
			ContentBase64: base64.StdEncoding.EncodeToString(content),
		}
		state.NextExportID++
		state.Exports = append(state.Exports, result)
		return nil
	})
	return result, err
}

func (s *Service) ListExports(token, roomID string) ([]domain.ExportResult, error) {
	state := s.store.Snapshot()
	if _, err := ownedRoom(&state, token, roomID); err != nil {
		return nil, err
	}
	results := make([]domain.ExportResult, 0)
	for _, result := range state.Exports {
		if result.RoomID == roomID {
			results = append(results, result)
		}
	}
	return results, nil
}

func accountByToken(state domain.State, token string) (domain.Account, bool) {
	for _, account := range state.Accounts {
		if account.Token == token {
			return account, true
		}
	}
	return domain.Account{}, false
}

func roomIndex(state domain.State, roomID string) int {
	for index, room := range state.Rooms {
		if room.ID == roomID {
			return index
		}
	}
	return -1
}

func ownedRoom(state *domain.State, token, roomID string) (*domain.Room, error) {
	account, ok := accountByToken(*state, token)
	if !ok {
		return nil, ErrForbidden
	}
	index := roomIndex(*state, roomID)
	if index < 0 {
		return nil, ErrNotFound
	}
	if state.Rooms[index].OwnerID != account.ID {
		return nil, ErrForbidden
	}
	return &state.Rooms[index], nil
}
