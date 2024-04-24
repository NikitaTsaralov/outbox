package server

import "cqrs-layout-example/config"

type Server struct {
	cfg *config.Config
}

func NewServer(cfg *config.Config) *Server {
	return &Server{cfg: cfg}
}

func (s *Server) Start() {

}

func (s *Server) Stop() {

}
