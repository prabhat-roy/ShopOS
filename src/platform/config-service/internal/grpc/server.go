package grpc

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	pb "github.com/shopos/config-service/internal/proto"
	"github.com/shopos/config-service/internal/service"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server exposes ConfigService methods over gRPC.
// Until proto compilation is complete, HTTP/JSON is used as a stand-in on the same port.
type Server struct {
	svc *service.ConfigService
	log *zap.Logger
}

func NewServer(svc *service.ConfigService, log *zap.Logger) *Server {
	return &Server{svc: svc, log: log}
}

// RegisterHTTP registers HTTP/JSON handlers as a temporary stand-in for gRPC.
// Replace with generated RegisterConfigServiceServer(grpcSrv, s) once protos are compiled.
func (s *Server) RegisterHTTP(mux *http.ServeMux) {
	mux.HandleFunc("GET /healthz", s.health)
	mux.HandleFunc("GET /config/{key...}", s.getConfig)
	mux.HandleFunc("PUT /config/{key...}", s.setConfig)
	mux.HandleFunc("DELETE /config/{key...}", s.deleteConfig)
	mux.HandleFunc("GET /configs", s.listConfigs)
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) getConfig(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	entry, err := s.svc.Get(r.Context(), key)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "key not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, entry)
}

func (s *Server) setConfig(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	var body struct {
		Value string `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Value == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "value is required"})
		return
	}
	if err := s.svc.Set(r.Context(), key, body.Value); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) deleteConfig(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if err := s.svc.Delete(r.Context(), key); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) listConfigs(w http.ResponseWriter, r *http.Request) {
	prefix := r.URL.Query().Get("prefix")
	entries, err := s.svc.List(r.Context(), prefix)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": entries})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

// grpcStatus converts service errors to gRPC status codes for when real gRPC is wired up.
func grpcStatus(err error) error {
	if errors.Is(err, service.ErrNotFound) {
		return status.Error(codes.NotFound, err.Error())
	}
	return status.Error(codes.Internal, err.Error())
}

// GetConfig is the gRPC handler (stub — wired after proto compilation).
func (s *Server) GetConfig(_ context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	entry, err := s.svc.Get(context.Background(), req.Key)
	if err != nil {
		return nil, grpcStatus(err)
	}
	return &pb.GetResponse{Entry: entry, Found: true}, nil
}
