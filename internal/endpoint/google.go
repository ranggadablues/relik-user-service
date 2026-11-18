package endpoint

import (
	"context"
	"relik-user-service/internal/businesslogic"

	pb "github.com/ranggadablues/relik-proto-library/user-proto/pb"
)

// UserServer implements the gRPC UserServiceServer interface
type GoogleServer struct {
	pb.UnimplementedGoogleAuthServiceServer
	businessLogic businesslogic.IGoogleBusinessLogic
}

// NewGoogleServer creates a new gRPC server with business logic
func NewGoogleServer(bl businesslogic.IGoogleBusinessLogic) pb.GoogleAuthServiceServer {
	return &GoogleServer{
		businessLogic: bl,
	}
}

func (s *GoogleServer) GoogleLogin(ctx context.Context, req *pb.GoogleLoginRequest) (*pb.GoogleAuthResponse, error) {
	return s.businessLogic.GoogleLogin(ctx, req)
}
