package businesslogic

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"relik-user-service/internal/repository"
	"strings"
	"time"

	models "relik-user-service/internal/models"

	"relik-user-service/internal/auth"

	"github.com/ranggadablues/gosok/auth/jwt"
	"github.com/ranggadablues/gosok/common"
	"github.com/ranggadablues/gosok/constants"
	"github.com/ranggadablues/gosok/email"
	"github.com/ranggadablues/gosok/security"
	"github.com/ranggadablues/relik-proto-library/user-proto/pb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type IUserBusinessLogic interface {
	CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateBasicUserResponse, error)
	ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error)
	GetUserById(ctx context.Context, req *pb.GetUserByIdRequest) (*pb.GetUserByIdResponse, error)
	GetUser(ctx context.Context, req *pb.Empty) (*pb.AuthResponse, error)
	UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error)
	DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error)
	LoginUser(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error)
	RefreshToken(ctx context.Context, req *pb.RefreshRequest) (*pb.LoginResponse, error)
	FindOrCreateUser(ctx context.Context, email, name, createdSource string) (*pb.User, *pb.UserDetails, error)
	UserVerification(ctx context.Context, req *pb.UserVerificationRequest) (*pb.UserVerificationResponse, error)
}

type UserBusinessLogic struct {
	userRepo func() repository.IUserRepo
}

func NewUserBusinessLogic() IUserBusinessLogic {
	return &UserBusinessLogic{userRepo: repository.NewUserRepo}
}

func (u *UserBusinessLogic) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateBasicUserResponse, error) {
	// validate user
	if err := createUserValidation(req); err != nil {
		return nil, err
	}

	user, userdetails, err := reqUserFromProto(req)
	if err != nil {
		return nil, err
	}

	// generate token
	timeStamp := common.ParseString(time.Now().UnixNano())
	encryptedLink, err := common.Encrypt(timeStamp, os.Getenv("ENCRYPTED_KEY"), "createaccount")
	if err != nil {
		return nil, common.ErrorInvalidArgument("", err)
	}

	// send activation email
	if err := u.snedActivationEmail(ctx, req, user); err != nil {
		return nil, err
	}

	// create user
	_, _, err = u.userRepo().CreateUser(ctx, user, userdetails)
	if err != nil {
		return nil, err
	}

	return &pb.CreateBasicUserResponse{Link: encryptedLink}, nil
}

func reqUserFromProto(req *pb.CreateUserRequest) (*models.User, *models.UserDetails, error) {
	userDetailsDoc := &models.UserDetails{}
	userDoc := &models.User{}
	email := strings.ToLower(req.User.Email)
	timeNow := timestamppb.New(time.Now())

	// Convert protobuf to BSON document
	req.Userdetails.Id = bson.NewObjectID().Hex()
	req.Userdetails.Name = common.Proper(req.Userdetails.Name)

	password, err := security.PleaseHash(req.Userdetails.Password, security.DefaultParams())
	if err != nil {
		return nil, nil, common.ErrorInternal("failed to hash password", err)
	}
	req.Userdetails.Password = password
	req.Userdetails.Createby = email
	req.Userdetails.Createdat = timeNow
	req.Userdetails.Updateby = email
	req.Userdetails.Updatedat = timeNow
	userDetailsDoc.FromProto(req.Userdetails)

	req.User.Id = bson.NewObjectID().Hex()
	req.User.Createdsource = constants.BasicAuth
	req.User.Userdetailsids = []string{userDetailsDoc.Id.Hex()}
	req.User.Email = email
	req.User.Createby = email
	req.User.Createdat = timeNow
	req.User.Updateby = email
	req.User.Updatedat = timeNow
	userDoc.FromProto(req.User)

	return userDoc, userDetailsDoc, nil
}

func (u *UserBusinessLogic) snedActivationEmail(ctx context.Context, req *pb.CreateUserRequest, user *models.User) error {
	// load email config
	notifintegrationid := os.Getenv("NOTIF_INTEGRATION_ID")
	notifPb := &pb.NotificationIntegrationsRequest{Id: notifintegrationid}
	notifResp, err := u.userRepo().GetNotificationIntegrationsById(ctx, notifPb)
	if err != nil {
		return err
	}

	// Create email sender
	cfg, err := email.LoadConfig(notifResp.Provider, notifResp)
	if err != nil {
		return err
	}
	sender, err := email.NewEmailSender(cfg)
	if err != nil {
		return common.ErrorInternal("failed to create email sender", err)
	}

	// load notification events
	notifEventPb := &pb.NotificationEventsRequest{Name: "user_activation"}
	notifEventResp, err := u.userRepo().GetNotificationEventsByName(ctx, notifEventPb)
	if err != nil {
		return err
	}

	// Send email
	// Create message
	msg := &email.Message{
		To:      []string{req.User.Email},
		Subject: notifEventResp.Subject,
		Body:    notifEventResp.Body,
	}

	// Handle signature inline image
	signaturePath := os.Getenv("SIGNATURE_IMAGE")
	var signatureFilename string
	if signaturePath != "" {
		if err := msg.AddInline(signaturePath); err != nil {
			// Log error but proceed? Or fail?
			// Given it's a signature, maybe we can proceed without it or fail.
			// Let's fail for now to be safe as per previous strictness.
			return common.ErrorInternal("failed to add inline signature", err)
		}
		signatureFilename = filepath.Base(signaturePath)
	}

	// generate activation link
	activationLink := fmt.Sprintf("%s-%s", user.Id.Hex(), common.ParseString(time.Now().UnixNano()))
	encryptedLink, err := common.Encrypt(activationLink, os.Getenv("ENCRYPTED_KEY"), "useractivation")
	if err != nil {
		return common.ErrorInternal("", err)
	}

	placeholders := map[string]string{
		"{{username}}":       req.Userdetails.Name,
		"{{activationlink}}": fmt.Sprintf("%s%s", os.Getenv("ACTIVATION_LINK"), encryptedLink),
		"{{signaturelogo}}":  fmt.Sprintf("<img src='cid:%s' />", signatureFilename),
	}

	msg.ParseTemplate(placeholders)

	err = sender.Send(ctx, msg)
	if err != nil {
		return common.ErrorInternal("failed to send email", err)
	}
	return nil
}

func createUserValidation(req *pb.CreateUserRequest) error {
	if req.Userdetails.Name == "" {
		return common.Error("name is required", nil)
	}

	if req.User.Email == "" {
		return common.Error("email is required", nil)
	}

	if !common.IsValidEmail(req.User.Email) {
		return common.Error("invalid email format", nil)
	}

	if req.Userdetails.Dob == "" {
		return common.Error("date of birth is required", nil)
	}

	if req.Userdetails.Password == "" {
		return common.Error("password is required", nil)
	}

	if len(req.Userdetails.Password) < 8 {
		return common.Error("password must be at least 8 characters", nil)
	}

	if req.Userdetails.Confirmpassword == "" {
		return common.Error("confirm password is required", nil)
	}

	if len(req.Userdetails.Confirmpassword) < 8 {
		return common.Error("confirm password must be at least 8 characters", nil)
	}

	if req.Userdetails.Password != req.Userdetails.Confirmpassword {
		return common.Error("password and confirm password must be same", nil)
	}

	if !req.Userdetails.Termofserviceandprivacypolicy {
		return common.Error("term of service and privacy policy must be accepted", nil)
	}

	return nil
}

func (u *UserBusinessLogic) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	return u.userRepo().ListUsers(ctx, req)
}

func (u *UserBusinessLogic) GetUserById(ctx context.Context, req *pb.GetUserByIdRequest) (*pb.GetUserByIdResponse, error) {
	return u.userRepo().GetUserById(ctx, req)
}

func (u *UserBusinessLogic) GetUser(ctx context.Context, req *pb.Empty) (*pb.AuthResponse, error) {
	authResponse := &pb.AuthResponse{}
	if err := jwt.IncomingContext(ctx, authResponse); err != nil {
		return nil, err
	}
	return authResponse, nil
}

func (u *UserBusinessLogic) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	return u.userRepo().UpdateUser(ctx, req)
}

func (u *UserBusinessLogic) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	return u.userRepo().DeleteUser(ctx, req)
}

func (u *UserBusinessLogic) LoginUser(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	// Validate user
	user, userdetails, err := u.userRepo().LoginUser(ctx, req)
	if err != nil {
		return nil, err
	}

	// the userInfo should be from the database
	userInfo := jwt.UserClaims{
		UserID:   user.Id,
		Email:    user.Email,
		Name:     userdetails.Name,
		Provider: "password",
	}
	token, refreshToken, err := auth.Manager().GenerateTokenPair(ctx, userInfo)
	if err != nil {
		return nil, err
	}
	return &pb.LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
	}, nil
}

func (u *UserBusinessLogic) RefreshToken(ctx context.Context, req *pb.RefreshRequest) (*pb.LoginResponse, error) {
	newAccess, newRefresh, err := auth.Manager().RefreshTokenPair(ctx, req.RefreshToken)
	if err != nil {
		return nil, err
	}

	return &pb.LoginResponse{
		Token:        newAccess,
		RefreshToken: newRefresh,
	}, nil
}

func (u *UserBusinessLogic) FindOrCreateUser(ctx context.Context, email, name, createdSource string) (*pb.User, *pb.UserDetails, error) {
	// find user by email
	usermodel, userdetailsmodel, err := u.userRepo().GetUserByEmail(ctx, email)
	if err != nil {
		if strings.Contains(err.Error(), "unknown email") {
			// Treat as user not found and proceed to create a new user
			err = nil
		} else {
			return nil, nil, err
		}
	}

	if err == nil && usermodel != nil && userdetailsmodel != nil && usermodel.Id != "" && userdetailsmodel.Id != "" {
		return usermodel, userdetailsmodel, nil
	}

	// if user not found, create user with proper fields for SSO
	timeNow := time.Now()
	userDetailsId := bson.NewObjectID()
	userId := bson.NewObjectID()

	userDetailsDoc := &models.UserDetails{
		Id:        userDetailsId,
		IsDefault: true,
		Name:      name,
	}
	userDetailsDoc.CreateBy = email
	userDetailsDoc.CreatedAt = timeNow
	userDetailsDoc.UpdateBy = email
	userDetailsDoc.UpdatedAt = timeNow

	userDoc := &models.User{
		Id:             userId,
		Email:          email,
		Isactive:       true,
		CreatedSource:  createdSource,
		UserDetailsIds: []bson.ObjectID{userDetailsId},
	}
	userDoc.CreateBy = email
	userDoc.CreatedAt = timeNow
	userDoc.UpdateBy = email
	userDoc.UpdatedAt = timeNow

	user, userdetails, err := u.userRepo().CreateUser(ctx, userDoc, userDetailsDoc)
	if err != nil {
		return nil, nil, err
	}

	return user, userdetails, nil
}

func (u *UserBusinessLogic) UserVerification(ctx context.Context, req *pb.UserVerificationRequest) (*pb.UserVerificationResponse, error) {
	decryptedString, err := common.Decrypt(req.Token, os.Getenv("ENCRYPTED_KEY"), req.Module)
	if err != nil {
		return nil, common.ErrorInvalidArgument("", err)
	}

	resp := &pb.UserVerificationResponse{}
	splitDecryptedString := strings.Split(decryptedString, "-")
	if len(splitDecryptedString) == 1 { // create account
		timestamp := splitDecryptedString[0]
		resp.Success = validateToken(timestamp)
		return resp, nil
	}

	if len(splitDecryptedString) == 2 { // account activation
		userId := splitDecryptedString[0]
		userid := common.ParseObjectID(userId)
		if userid.IsZero() {
			resp.Success = false
			return resp, common.ErrorInvalidArgument("Something went wrong!", nil)
		}
		timestamp := splitDecryptedString[1]
		if !validateToken(timestamp) {
			resp.Success = false
			return resp, nil
		}

		userResp, err := u.userRepo().GetUserById(ctx, &pb.GetUserByIdRequest{Id: userId})
		if err != nil || userResp == nil || userResp.User == nil {
			resp.Success = false
			return resp, common.ErrorInvalidArgument("Something went wrong!", err)
		}

		// Prevent replay if account is already activated
		if userResp.User.Isactive {
			resp.Success = false
			return resp, nil
		}

		// Update user status to active
		if err := u.userRepo().ActivateUser(ctx, userId, userResp.User.Email); err != nil {
			return nil, err
		}

		resp.Success = true
		return resp, nil
	}

	return &pb.UserVerificationResponse{Success: false}, common.ErrorInvalidArgument("Something went wrong!", nil)
}

func validateToken(timestamp string) bool {
	timeStamp := int64(common.ParseInt(timestamp))
	if timeStamp <= 0 {
		return false
	}
	userActivationTimestamp := time.Unix(0, timeStamp)
	if userActivationTimestamp.IsZero() {
		return false
	}

	diff := time.Since(userActivationTimestamp)
	if diff < 0 || diff > 15*time.Minute {
		return false
	}

	return true
}
