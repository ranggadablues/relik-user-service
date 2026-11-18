package test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"relik-user-service/internal/businesslogic"

	pb "github.com/ranggadablues/relik-proto-library/user-proto/pb"
)

// Mock repository
type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) CreateUser(req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	args := m.Called(req)
	return args.Get(0).(*pb.CreateUserResponse), args.Error(1)
}
func (m *MockRepo) ListUsers(req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	args := m.Called(req)
	return args.Get(0).(*pb.ListUsersResponse), args.Error(1)
}
func (m *MockRepo) GetUserById(req *pb.GetUserByIdRequest) (*pb.GetUserByIdResponse, error) {
	args := m.Called(req)
	return args.Get(0).(*pb.GetUserByIdResponse), args.Error(1)
}
func (m *MockRepo) UpdateUser(req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	args := m.Called(req)
	return args.Get(0).(*pb.UpdateUserResponse), args.Error(1)
}
func (m *MockRepo) DeleteUser(req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	args := m.Called(req)
	return args.Get(0).(*pb.DeleteUserResponse), args.Error(1)
}
func (m *MockRepo) LoginUser(req *pb.LoginRequest) (*pb.User, error) {
	args := m.Called(req)
	return args.Get(0).(*pb.User), args.Error(1)
}
func (m *MockRepo) RefreshToken(req *pb.RefreshRequest) (*pb.LoginResponse, error) {
	args := m.Called(req)
	return args.Get(0).(*pb.LoginResponse), args.Error(1)
}

func TestUserService(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	mockRepo := new(MockRepo)
	svc := businesslogic.NewUserBusinessLogic()

	// --- CreateUser
	const emailTest = "johndoe@email.com"
	reqCreate := &pb.CreateUserRequest{User: &pb.User{Email: emailTest}}
	mockRepo.On("CreateUser", reqCreate).Return(&pb.CreateUserResponse{User: &pb.User{Id: "1", Email: emailTest}}, nil)
	resCreate, err := svc.CreateUser(ctx, reqCreate)
	assert.NoError(t, err)
	assert.Equal(t, emailTest, resCreate.GetLink())

	// --- ListUsers
	reqList := &pb.ListUsersRequest{}
	users := []*pb.User{{Id: "1", Email: emailTest}}
	mockRepo.On("ListUsers", reqList).Return(&pb.ListUsersResponse{Users: users}, nil)
	resList, err := svc.ListUsers(ctx, reqList)
	assert.NoError(t, err)
	assert.Len(t, resList.Users, 1)

	// --- GetUserById
	reqGet := &pb.GetUserByIdRequest{Id: "1"}
	user := &pb.User{Id: "1", Email: emailTest}
	mockRepo.On("GetUserById", reqGet).Return(&pb.GetUserByIdResponse{User: user}, nil)
	resGet, err := svc.GetUserById(ctx, reqGet)
	assert.NoError(t, err)
	assert.Equal(t, "1", resGet.User.Id)

	// --- UpdateUser
	reqUpdate := &pb.UpdateUserRequest{Id: "1", User: &pb.UserUpdate{Email: "john@doe.com", Password: "password", Isactive: true}}
	mockRepo.On("UpdateUser", reqUpdate).Return(&pb.UpdateUserResponse{Success: true}, nil)
	resUpdate, err := svc.UpdateUser(ctx, reqUpdate)
	assert.NoError(t, err)
	assert.Equal(t, true, resUpdate.Success)

	// --- DeleteUser
	reqDelete := &pb.DeleteUserRequest{Id: "1"}
	mockRepo.On("DeleteUser", reqDelete).Return(&pb.DeleteUserResponse{Success: true}, nil)
	resDelete, err := svc.DeleteUser(ctx, reqDelete)
	assert.NoError(t, err)
	assert.Equal(t, true, resDelete.Success)

	// --- LoginUser
	reqRegister := &pb.LoginRequest{Email: "john@doe.com", Password: "password"}
	resRegister, err := svc.LoginUser(ctx, reqRegister)
	assert.NoError(t, err)
	// make sure token is not empty
	assert.NotEmpty(t, resRegister.Token, "token should not be empty")
	// optionally ensure it has a Bearer prefix
	assert.True(t, len(resRegister.Token) > 10, "token seems too short")

	// --- RefreshToken
	reqRefresh := &pb.RefreshRequest{RefreshToken: "refreshToken"}
	mockRepo.On("RefreshToken", reqRefresh).Return(&pb.LoginResponse{Token: "token", RefreshToken: "refreshToken"}, nil)
	resRefresh, err := svc.RefreshToken(ctx, reqRefresh)
	assert.NoError(t, err)
	assert.Equal(t, "token", resRefresh.Token)
	assert.Equal(t, "refreshToken", resRefresh.RefreshToken)

	mockRepo.AssertExpectations(t)
}
