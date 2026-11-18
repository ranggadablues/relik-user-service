package endpoint

import (
	"context"
	"relik-user-service/internal/businesslogic"

	pb "github.com/ranggadablues/relik-proto-library/user-proto/pb"
)

// UserServer implements the gRPC UserServiceServer interface
type FacebookServer struct {
	pb.UnimplementedFacebookAuthServiceServer
	businessLogic businesslogic.IFacebookBusinessLogic
}

// NewFacebookServer creates a new gRPC server with business logic
func NewFacebookServer(bl businesslogic.IFacebookBusinessLogic) pb.FacebookAuthServiceServer {
	return &FacebookServer{
		businessLogic: bl,
	}
}

func (s *FacebookServer) FacebookLogin(ctx context.Context, req *pb.FacebookLoginRequest) (*pb.FacebookAuthResponse, error) {
	return s.businessLogic.FacebookLogin(ctx, req)
}
