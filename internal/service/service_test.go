package service_test

import (
	"context"
	"encoding/base64"
	"errors"
	"reflect"
	"testing"

	"weddinglive/internal/domain"
	"weddinglive/internal/export"
	"weddinglive/internal/service"
	"weddinglive/internal/store"
)

func TestWeddingLiveBusinessWorkflow(t *testing.T) {
	svc := service.New(store.NewMemory(domain.State{}), "admin-token")
	owner, err := svc.CreateAccount("admin-token", "摄影师甲")
	if err != nil {
		t.Fatal(err)
	}
	other, err := svc.CreateAccount("admin-token", "摄影师乙")
	if err != nil {
		t.Fatal(err)
	}
	room, err := svc.CreateRoom(owner.Token, "海边婚礼")
	if err != nil {
		t.Fatal(err)
	}
	photo, err := svc.AddPhoto(owner.Token, room.ID, "ceremony.jpg", "交换戒指", base64.StdEncoding.EncodeToString([]byte("fixture-image")))
	if err != nil {
		t.Fatal(err)
	}
	rooms := svc.ListRooms()
	if len(rooms) != 1 || rooms[0].ID != room.ID || rooms[0].Photos != nil {
		t.Fatalf("unexpected public rooms: %+v", rooms)
	}
	photos, err := svc.ListPhotos(room.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(photos) != 1 || photos[0].ID != photo.ID {
		t.Fatalf("unexpected photos: %+v", photos)
	}
	if err := svc.DeletePhoto(other.Token, room.ID, photo.ID); !errors.Is(err, service.ErrForbidden) {
		t.Fatalf("other photographer delete error = %v", err)
	}
	if err := svc.DeletePhoto(owner.Token, room.ID, photo.ID); err != nil {
		t.Fatal(err)
	}
	adminRooms, err := svc.AdminRooms("admin-token")
	if err != nil {
		t.Fatal(err)
	}
	if len(adminRooms) != 1 || len(adminRooms[0].Photos) != 0 {
		t.Fatalf("unexpected admin rooms: %+v", adminRooms)
	}
	if err := svc.DeleteRoom("admin-token", room.ID); err != nil {
		t.Fatal(err)
	}
	if got := svc.ListRooms(); len(got) != 0 {
		t.Fatalf("rooms after delete: %+v", got)
	}
}

func TestWeddingExportStopsWhenRequestIsCanceled(t *testing.T) {
	svc := service.New(store.NewMemory(domain.State{}), "admin-token")
	owner, err := svc.CreateAccount("admin-token", "导出摄影师")
	if err != nil {
		t.Fatal(err)
	}
	room, err := svc.CreateRoom(owner.Token, "分片导出婚礼")
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 5; i++ {
		_, err := svc.AddPhoto(owner.Token, room.ID, "image.jpg", "", base64.StdEncoding.EncodeToString([]byte{byte(i)}))
		if err != nil {
			t.Fatal(err)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	started := make([]int, 0, 2)
	result, exportErr := svc.ExportRoom(ctx, owner.Token, room.ID, export.Options{
		ChunkSize: 2,
		OnChunkStart: func(index int) {
			started = append(started, index)
			if index == 1 {
				cancel()
			}
		},
	})

	if !errors.Is(exportErr, context.Canceled) {
		t.Errorf("export error = %v, want context canceled", exportErr)
	}
	if result.ID != "" {
		t.Errorf("export result = %+v, want empty result", result)
	}
	if !reflect.DeepEqual(started, []int{0, 1}) {
		t.Errorf("started chunks = %v, want [0 1]", started)
	}
	results, err := svc.ListExports(owner.Token, room.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Errorf("committed exports = %+v, want none", results)
	}
}
