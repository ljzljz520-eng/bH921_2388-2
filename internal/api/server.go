package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"weddinglive/internal/export"
	"weddinglive/internal/service"
)

type Server struct {
	service *service.Service
}

func New(svc *service.Service) *Server {
	return &Server{service: svc}
}

func (s *Server) Handler() http.Handler {
	router := chi.NewRouter()
	router.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	router.Route("/api", func(router chi.Router) {
		router.Post("/admin/accounts", s.createAccount)
		router.Get("/admin/rooms", s.adminRooms)
		router.Delete("/admin/rooms/{roomID}", s.deleteRoom)
		router.Get("/rooms", s.listRooms)
		router.Post("/rooms", s.createRoom)
		router.Get("/rooms/{roomID}/photos", s.listPhotos)
		router.Post("/rooms/{roomID}/photos", s.addPhoto)
		router.Delete("/rooms/{roomID}/photos/{photoID}", s.deletePhoto)
		router.Post("/rooms/{roomID}/exports", s.createExport)
		router.Get("/rooms/{roomID}/exports", s.listExports)
	})
	return router
}

func (s *Server) createAccount(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name string `json:"name"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	account, err := s.service.CreateAccount(r.Header.Get("X-Admin-Token"), input.Name)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, account)
}

func (s *Server) createRoom(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title string `json:"title"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	room, err := s.service.CreateRoom(r.Header.Get("X-Photographer-Token"), input.Title)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, room)
}

func (s *Server) listRooms(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.service.ListRooms())
}

func (s *Server) adminRooms(w http.ResponseWriter, r *http.Request) {
	rooms, err := s.service.AdminRooms(r.Header.Get("X-Admin-Token"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rooms)
}

func (s *Server) deleteRoom(w http.ResponseWriter, r *http.Request) {
	if err := s.service.DeleteRoom(r.Header.Get("X-Admin-Token"), chi.URLParam(r, "roomID")); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) addPhoto(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name          string `json:"name"`
		Caption       string `json:"caption"`
		ContentBase64 string `json:"content_base64"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	photo, err := s.service.AddPhoto(
		r.Header.Get("X-Photographer-Token"),
		chi.URLParam(r, "roomID"),
		input.Name,
		input.Caption,
		input.ContentBase64,
	)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, photo)
}

func (s *Server) listPhotos(w http.ResponseWriter, r *http.Request) {
	photos, err := s.service.ListPhotos(chi.URLParam(r, "roomID"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, photos)
}

func (s *Server) deletePhoto(w http.ResponseWriter, r *http.Request) {
	err := s.service.DeletePhoto(
		r.Header.Get("X-Photographer-Token"),
		chi.URLParam(r, "roomID"),
		chi.URLParam(r, "photoID"),
	)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) createExport(w http.ResponseWriter, r *http.Request) {
	result, err := s.service.ExportRoom(
		r.Context(),
		r.Header.Get("X-Photographer-Token"),
		chi.URLParam(r, "roomID"),
		export.Options{ChunkSize: 2},
	)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

func (s *Server) listExports(w http.ResponseWriter, r *http.Request) {
	results, err := s.service.ListExports(
		r.Header.Get("X-Photographer-Token"),
		chi.URLParam(r, "roomID"),
	)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, results)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, value any) bool {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return false
	}
	return true
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrForbidden):
		writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
	case errors.Is(err, service.ErrNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
	case errors.Is(err, service.ErrInvalid):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	case errors.Is(err, context.Canceled):
		writeJSON(w, 499, map[string]string{"error": "request canceled"})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
