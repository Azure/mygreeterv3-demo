package demoserver

import (
	pb "dev.azure.com/service-hub-flg/service_hub_validation/_git/service_hub_validation_service.git/mygreeterv3/api/v1"
)

type Server struct {
	// When the UnimplementedMyGreeterServer struct is embedded,
	// the generated method/implementation in .pb file will be associated with this struct.
	// If this struct doesn't implment some methods,
	// the .pb ones will be used. If this struct implement the methods, it will override the .pb ones.
	// The reason is that anonymous field's methods are promoted to the struct.
	//
	// When this struct is NOT embedded,, all methods have to be implemented to meet the interface requirement.
	// See https://go.dev/ref/spec#Struct_types.
	pb.UnimplementedMyGreeterServer
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) init(options Options) {
}
