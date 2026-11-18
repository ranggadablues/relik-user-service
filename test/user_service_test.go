package test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"relik-user-service/internal/auth"
	"relik-user-service/internal/businesslogic"
	"relik-user-service/internal/models"

	pb "github.com/ranggadablues/relik-proto-library/user-proto/pb"
)

// Mock repository implementing repository.IUserRepo
type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) CreateUser(ctx context.Context, userDoc *models.User, userDetailsDoc *models.UserDetails) (*pb.User, *pb.UserDetails, error) {
	args := m.Called(ctx, userDoc, userDetailsDoc)
	return args.Get(0).(*pb.User), args.Get(1).(*pb.UserDetails), args.Error(2)
}
func (m *MockRepo) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*pb.ListUsersResponse), args.Error(1)
}
func (m *MockRepo) GetUserById(ctx context.Context, req *pb.GetUserByIdRequest) (*pb.GetUserByIdResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*pb.GetUserByIdResponse), args.Error(1)
}
func (m *MockRepo) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*pb.UpdateUserResponse), args.Error(1)
}
func (m *MockRepo) ActivateUser(ctx context.Context, id string, updatedBy string) error {
	args := m.Called(ctx, id, updatedBy)
	return args.Error(0)
}
func (m *MockRepo) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*pb.DeleteUserResponse), args.Error(1)
}
func (m *MockRepo) LoginUser(ctx context.Context, req *pb.LoginRequest) (*pb.User, *models.UserDetails, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*pb.User), args.Get(1).(*models.UserDetails), args.Error(2)
}
func (m *MockRepo) GetUserByEmail(ctx context.Context, email string) (*pb.User, *pb.UserDetails, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*pb.User), args.Get(1).(*pb.UserDetails), args.Error(2)
}
func (m *MockRepo) GetNotificationIntegrationsById(ctx context.Context, req *pb.NotificationIntegrationsRequest) (models.NotificationIntegrations, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(models.NotificationIntegrations), args.Error(1)
}
func (m *MockRepo) GetNotificationEventsByName(ctx context.Context, req *pb.NotificationEventsRequest) (*pb.NotificationEvents, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*pb.NotificationEvents), args.Error(1)
}

func TestUserService(t *testing.T) {
	t.Setenv("ACCESS_SECRET", "test-access-secret-key-at-least-32-chars")
	t.Setenv("REFRESH_SECRET", "test-refresh-secret-key-at-least-32-chars")
	assert.NoError(t, auth.Init())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mockRepo := new(MockRepo)
	svc := businesslogic.NewUserBusinessLogicWithRepo(mockRepo)

	const emailTest = "johndoe@email.com"

	// --- ListUsers
	reqList := &pb.ListUsersRequest{}
	users := []*pb.User{{Id: "1", Email: emailTest}}
	mockRepo.On("ListUsers", ctx, reqList).Return(&pb.ListUsersResponse{Users: users}, nil)
	resList, err := svc.ListUsers(ctx, reqList)
	assert.NoError(t, err)
	assert.Len(t, resList.Users, 1)

	// --- GetUserById
	reqGet := &pb.GetUserByIdRequest{Id: "1"}
	user := &pb.User{Id: "1", Email: emailTest}
	mockRepo.On("GetUserById", ctx, reqGet).Return(&pb.GetUserByIdResponse{User: user}, nil)
	resGet, err := svc.GetUserById(ctx, reqGet)
	assert.NoError(t, err)
	assert.Equal(t, "1", resGet.User.Id)

	// --- UpdateUser
	reqUpdate := &pb.UpdateUserRequest{Id: "1", User: &pb.UserUpdate{Email: "john@doe.com", Password: "password", Isactive: true}}
	mockRepo.On("UpdateUser", ctx, reqUpdate).Return(&pb.UpdateUserResponse{Success: true}, nil)
	resUpdate, err := svc.UpdateUser(ctx, reqUpdate)
	assert.NoError(t, err)
	assert.Equal(t, true, resUpdate.Success)

	// --- DeleteUser
	reqDelete := &pb.DeleteUserRequest{Id: "1"}
	mockRepo.On("DeleteUser", ctx, reqDelete).Return(&pb.DeleteUserResponse{Success: true}, nil)
	resDelete, err := svc.DeleteUser(ctx, reqDelete)
	assert.NoError(t, err)
	assert.Equal(t, true, resDelete.Success)

	// --- LoginUser
	reqLogin := &pb.LoginRequest{Email: "john@doe.com", Password: "password"}
	mockRepo.On("LoginUser", ctx, reqLogin).Return(
		&pb.User{Id: "1", Email: "john@doe.com"},
		&models.UserDetails{Name: "John Doe"},
		nil,
	)
	resLogin, err := svc.LoginUser(ctx, reqLogin)
	assert.NoError(t, err)
	assert.NotEmpty(t, resLogin.Token, "token should not be empty")
	assert.NotEmpty(t, resLogin.RefreshToken, "refreshToken should not be empty")

	// --- RefreshToken
	reqRefresh := &pb.RefreshRequest{RefreshToken: resLogin.RefreshToken}
	resRefresh, err := svc.RefreshToken(ctx, reqRefresh)
	assert.NoError(t, err)
	assert.NotEmpty(t, resRefresh.Token)
	assert.NotEmpty(t, resRefresh.RefreshToken)

	mockRepo.AssertExpectations(t)
}
