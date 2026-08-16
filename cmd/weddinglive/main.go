package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"

	"weddinglive/internal/api"
	"weddinglive/internal/config"
	"weddinglive/internal/domain"
	"weddinglive/internal/fixture"
	"weddinglive/internal/service"
	"weddinglive/internal/store"
)

func main() {
	configPath := flag.String("config", "", "path to process-only JSON configuration")
	demo := flag.Bool("demo", false, "run a deterministic in-memory workflow")
	flag.Parse()

	log.SetFlags(0)
	if err := run(*configPath, *demo); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

func run(configPath string, demo bool) error {
	initial, err := fixture.Demo()
	if err != nil {
		return err
	}
	if demo {
		return runDemo(initial)
	}

	cfg := config.Default()
	if configPath != "" {
		cfg, err = config.Load(configPath)
		if err != nil {
			return err
		}
	}
	repository, err := store.Open(cfg.DataFile, initial)
	if err != nil {
		return err
	}
	handler := api.New(service.New(repository, cfg.AdminToken)).Handler()
	log.Printf("wedding live API listening on http://%s", cfg.Address)
	return http.ListenAndServe(cfg.Address, handler)
}

func runDemo(initial domain.State) error {
	repository := store.NewMemory(initial)
	svc := service.New(repository, "admin-fixture-token")
	account, err := svc.CreateAccount("admin-fixture-token", "演示摄影师")
	if err != nil {
		return err
	}
	room, err := svc.CreateRoom(account.Token, "确定性演示婚礼")
	if err != nil {
		return err
	}
	photo, err := svc.AddPhoto(account.Token, room.ID, "demo.jpg", "现场合影", "ZGVtby1pbWFnZQ==")
	if err != nil {
		return err
	}
	output := struct {
		Account domain.Account `json:"account"`
		Room    domain.Room    `json:"room"`
		Photo   domain.Photo   `json:"photo"`
		Rooms   []domain.Room  `json:"public_rooms"`
	}{Account: account, Room: room, Photo: photo, Rooms: svc.ListRooms()}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}
