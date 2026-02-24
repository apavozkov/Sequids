package orchestrator

import (
	"context"
	"log"

	orchestratorpb "github.com/apavozkov/sequids/pkg/proto/orchestratorpb"
)

type Server struct {
	orchestratorpb.UnimplementedOrchestratorServiceServer
	store *Store
}

func NewServer(store *Store) *Server {
	return &Server{store: store}
}

func (s *Server) ReportSensorStatus(ctx context.Context, req *orchestratorpb.ReportSensorStatusRequest) (*orchestratorpb.ReportSensorStatusResponse, error) {
	log.Printf("sensor report: id=%s status=%s value=%.4f", req.GetSensorId(), req.GetStatus(), req.GetValue())
	if s.store != nil {
		s.store.UpdateStatus(req.GetSensorId(), req.GetStatus(), req.GetValue())
	}
	return &orchestratorpb.ReportSensorStatusResponse{Ack: "ok"}, nil
}
