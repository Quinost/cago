package http_s

import (
	"cago/internal"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type HttpServer struct {
	cfg      *internal.Config
	cachesrv *internal.CacheService
	server   *http.Server
	ctx      context.Context
}

func NewHttpServer(cfg *internal.Config, cachesrv *internal.CacheService, ctx context.Context) *HttpServer {
	return &HttpServer{
		cfg:      cfg,
		cachesrv: cachesrv,
		ctx:      ctx,
	}
}

func (s *HttpServer) Run() error {
	r := chi.NewRouter()

	r.Use(middleware.SetHeader("Content-Type", "application/json"))

	r.Route("/v1", func(r chi.Router) {
		r.Get("/health", s.handleHealth)
		r.Get("/stats", s.handleStats)

		r.Route("/keys", func(r chi.Router) {
			r.Get("/", s.handleKeysList)
			r.Route("/{key}", func(r chi.Router) {
				r.Get("/", s.handleGet)
				r.Put("/", s.handleSet)
				r.Delete("/", s.handleDelete)
				r.Post("/expire", s.handleExpire)
			})
		})
	})

	httpPort := s.cfg.Port + 1000
	addr := net.JoinHostPort(s.cfg.Host, strconv.Itoa(httpPort))

	s.server = &http.Server{
		Addr:    addr,
		Handler: r,
	}

	fmt.Printf("HTTP server listening on %s\n", addr)

	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

// GET /v1/keys?pattern=user:*
func (s *HttpServer) handleKeysList(w http.ResponseWriter, r *http.Request) {
	pattern := r.URL.Query().Get("pattern")
	if pattern == "" {
		pattern = "*"
	}

	keys, err := s.cachesrv.Keys(pattern)
	if err != nil {
		s.errorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := KeysListResponse{
		Keys:    keys,
		Count:   len(keys),
		Pattern: pattern,
	}

	s.jsonResponse(w, response, http.StatusOK)
}

// GET /v1/keys/{key}
func (s *HttpServer) handleGet(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")

	value, exists, err := s.cachesrv.Get(key)
	if err != nil {
		s.errorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !exists {
		s.errorResponse(w, "key not found", http.StatusNotFound)
		return
	}

	ttl, _ := s.cachesrv.TTL(key)
	ttlSeconds := int64(-1)
	if ttl > 0 {
		ttlSeconds = int64(ttl.Seconds())
	} else if ttl == -2*time.Second {
		ttlSeconds = -2
	}

	response := GetResponse{
		Key:   key,
		Value: value,
		Ttl:   ttlSeconds,
	}

	s.jsonResponse(w, response, http.StatusOK)
}

// PUT /v1/keys/{key}
// {"value": "value", "ttl" : 60}
func (s *HttpServer) handleSet(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")

	var req SetRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.errorResponse(w, "invalid json body", http.StatusBadRequest)
		return
	}

	ttl := time.Duration(req.TTL) * time.Second

	if err := s.cachesrv.Set(key, req.Value, ttl); err != nil {
		s.errorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := SetResponse{
		Key:     key,
		Value:   req.Value,
		TTL:     req.TTL,
		Success: true,
	}

	s.jsonResponse(w, response, http.StatusOK)
}

// DELETE /v1/keys/{key}
func (s *HttpServer) handleDelete(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")

	deleted, err := s.cachesrv.Delete(key)
	if err != nil {
		s.errorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !deleted {
		s.errorResponse(w, "key not found", http.StatusNotFound)
		return
	}

	response := DeleteResponse{
		Key:     key,
		Deleted: true,
	}

	s.jsonResponse(w, response, http.StatusOK)
}

// POST /v1/keys/{key}/expire
// {"ttl": 60}
func (s *HttpServer) handleExpire(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")

	var req ExpireRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.errorResponse(w, "invalid json body", http.StatusBadRequest)
		return
	}

	ttl := time.Duration(req.TTL) * time.Second

	if err := s.cachesrv.Expire(key, ttl); err != nil {
		if err == internal.ErrKeyNotFound {
			s.errorResponse(w, "key not found", http.StatusNotFound)
			return
		}
		s.errorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := ExpireResponse{
		Key:     key,
		TTL:     req.TTL,
		Success: true,
	}

	s.jsonResponse(w, response, http.StatusOK)
}

func (s *HttpServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status: "ok",
		Server: "cago",
	}

	s.jsonResponse(w, response, http.StatusOK)
}

// GET /v1/stats
func (s *HttpServer) handleStats(w http.ResponseWriter, _ *http.Request) {
	keys, _ := s.cachesrv.Keys("*")

	response := StatsResponse{
		TotalKeys:       len(keys),
		DefaultTTL:      s.cfg.DefaultTTL.Seconds(),
		CleanupInterval: s.cfg.CleanupInterval.Seconds(),
	}

	s.jsonResponse(w, response, http.StatusOK)
}

func (s *HttpServer) jsonResponse(w http.ResponseWriter, data interface{}, status int) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (s *HttpServer) errorResponse(w http.ResponseWriter, message string, status int) {
	response := ErrorResponse{
		Error:  message,
		Status: status,
	}
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}
