package externalserver

import (
	pb "dev.azure.com/service-hub-flg/service_hub_validation/_git/service_hub_validation_service.git/mygreeterv3/api/v1"
)

type Server struct {
	pb.UnimplementedMyGreeterServer
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) init(options Options) {
}
