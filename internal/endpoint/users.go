package endpoint

import (
	"context"
	"relik-user-service/internal/businesslogic"

	pb "github.com/ranggadablues/relik-proto-library/user-proto/pb"
)

// UserServer implements the gRPC UserServiceServer interface
type UserServer struct {
	pb.UnimplementedUserServiceServer
	businessLogic businesslogic.IUserBusinessLogic
}

// NewUserServer creates a new gRPC server with business logic
func NewUserServer(bl businesslogic.IUserBusinessLogic) pb.UserServiceServer {
	return &UserServer{
		businessLogic: bl,
	}
}

// CreateUser handles the CreateUser gRPC call
func (s *UserServer) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateBasicUserResponse, error) {
	return s.businessLogic.CreateUser(ctx, req)
}

// ListUsers handles the ListUsers gRPC call
func (s *UserServer) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	return s.businessLogic.ListUsers(ctx, req)
}

// GetUserById handles the GetUserById gRPC call
func (s *UserServer) GetUserById(ctx context.Context, req *pb.GetUserByIdRequest) (*pb.GetUserByIdResponse, error) {
	return s.businessLogic.GetUserById(ctx, req)
}

// GetUser handles the GetUser gRPC call
func (s *UserServer) GetUser(ctx context.Context, req *pb.Empty) (*pb.AuthResponse, error) {
	return s.businessLogic.GetUser(ctx, req)
}

// UpdateUser handles the UpdateUser gRPC call
func (s *UserServer) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	return s.businessLogic.UpdateUser(ctx, req)
}

// DeleteUser handles the DeleteUser gRPC call
func (s *UserServer) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	return s.businessLogic.DeleteUser(ctx, req)
}

// Register handles the Register gRPC call
func (s *UserServer) LoginUser(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	return s.businessLogic.LoginUser(ctx, req)
}

func (s *UserServer) RefreshToken(ctx context.Context, req *pb.RefreshRequest) (*pb.LoginResponse, error) {
	return s.businessLogic.RefreshToken(ctx, req)
}

func (s *UserServer) UserVerification(ctx context.Context, req *pb.UserVerificationRequest) (*pb.UserVerificationResponse, error) {
	return s.businessLogic.UserVerification(ctx, req)
}
