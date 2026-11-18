package repository

import (
	"context"
	"errors"
	"time"

	"relik-user-service/internal/models"
	"github.com/ranggadablues/relik-proto-library/user-proto/pb"

	"github.com/ranggadablues/gosok/common"
	"github.com/ranggadablues/gosok/constants"
	"github.com/ranggadablues/gosok/db"
	"github.com/ranggadablues/gosok/db/ref"
	"github.com/ranggadablues/gosok/logger"
	"github.com/ranggadablues/gosok/security"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type IUserRepo interface {
	CreateUser(ctx context.Context, userDoc *models.User, userDetailsDoc *models.UserDetails) (*pb.User, *pb.UserDetails, error)
	ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error)
	GetUserById(ctx context.Context, req *pb.GetUserByIdRequest) (*pb.GetUserByIdResponse, error)
	UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error)
	ActivateUser(ctx context.Context, id string, updatedBy string) error
	DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error)
	LoginUser(ctx context.Context, req *pb.LoginRequest) (*pb.User, *models.UserDetails, error)
	GetUserByEmail(ctx context.Context, email string) (*pb.User, *pb.UserDetails, error)
	GetNotificationIntegrationsById(ctx context.Context, req *pb.NotificationIntegrationsRequest) (models.NotificationIntegrations, error)
	GetNotificationEventsByName(ctx context.Context, req *pb.NotificationEventsRequest) (*pb.NotificationEvents, error)
}

type UserRepo struct {
	db     func(opts ...db.MongoOption) (db.IMongoLib, error)
	logger logger.ILogLevel
}

func NewUserRepo() IUserRepo {
	return &UserRepo{
		db:     db.NewMongo,
		logger: logger.NewLogger(),
	}
}

func (r *UserRepo) checkEmailDupe(ctx context.Context, user *models.User) (bool, error) {
	mongo, err := r.db()
	if err != nil {
		return false, common.ErrorInternal(constants.FailedToConnectDatabase, err)
	}
	count, err := mongo.Count(ctx, UserCollection(), bson.M{"email": user.Email})
	if err != nil {
		return false, common.ErrorFind(err)
	}
	if count > 0 {
		return true, nil
	}
	return false, nil
}

func (r *UserRepo) CreateUser(ctx context.Context, userDoc *models.User, userDetailsDoc *models.UserDetails) (*pb.User, *pb.UserDetails, error) {
	dupe, err := r.checkEmailDupe(ctx, userDoc)
	if err != nil {
		return nil, nil, err
	}

	if dupe {
		return nil, nil, common.ErrorAlreadyExists("email already exists", nil)
	}

	mongo, err := r.db()
	if err != nil {
		return nil, nil, common.ErrorInternal(constants.FailedToConnectDatabase, err)
	}
	err = mongo.WithTransaction(ctx, func(txctx context.Context) error {
		_, err = mongo.InsertOne(txctx, UserDetailsCollection(), userDetailsDoc)
		if err != nil {
			return common.ErrorInsert(err)
		}

		_, err = mongo.InsertOne(txctx, UserCollection(), userDoc)
		if err != nil {
			return common.ErrorInsert(err)
		}
		return nil
	})
	if err != nil {
		return nil, nil, common.ErrorInternal(constants.FailedToConnectDatabase, err)
	}

	user := userDoc.ToProto()
	userDetails := userDetailsDoc.ToProto()
	return user, userDetails, nil
}

func (r *UserRepo) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	// get from database
	mongo, err := r.db()
	if err != nil {
		return nil, common.ErrorInternal(constants.FailedToConnectDatabase, err)
	}

	usersOutput := make([]bson.M, 0)
	optSort := ref.WithSort(bson.D{{Key: "firstname", Value: 1}})
	err = mongo.Find(ctx, &usersOutput, bson.M{}, UserCollection(), optSort)
	if err != nil {
		return nil, common.ErrorFind(err)
	}

	pbusers := make([]*pb.User, 0)
	for _, user := range usersOutput {
		userModel := models.User{}
		userModel.Id = common.ParseObjectID(user["_id"])
		userModel.Email = common.ParseString(user["email"])
		userModel.Isactive = common.ParseBool(user["isactive"])
		userModel.CreatedAt = common.ParseTime(user["createdat"])
		userModel.UpdatedAt = common.ParseTime(user["updatedat"])
		userModel.CreateBy = common.ParseString(user["createby"])
		userModel.UpdateBy = common.ParseString(user["updateby"])
		pbUser := userModel.ToProto()
		pbusers = append(pbusers, pbUser)
	}
	return &pb.ListUsersResponse{Users: pbusers}, nil
}

func (r *UserRepo) GetUserById(ctx context.Context, req *pb.GetUserByIdRequest) (*pb.GetUserByIdResponse, error) {
	// Convert string ID to MongoDB ObjectID using v2 driver
	objectId, err := bson.ObjectIDFromHex(req.Id)
	if err != nil {
		return nil, common.Error(constants.InvalidUserIDFormat, err)
	}

	// Get from database with proper filter
	mongo, err := r.db()
	if err != nil {
		return nil, common.ErrorInternal(constants.FailedToConnectDatabase, err)
	}

	var userMap bson.M
	filter := bson.M{"_id": objectId}
	if err = mongo.FindOne(ctx, &userMap, filter, UserCollection()); err != nil {
		return nil, common.ErrorFind(err)
	}

	// Map database result to protobuf User struct
	user := &pb.User{
		Id:       common.ParseString(userMap["_id"]),
		Email:    common.ParseString(userMap["email"]),
		Isactive: common.ParseBool(userMap["isactive"]),
	}

	return &pb.GetUserByIdResponse{
		User: user,
	}, nil
}

func (r *UserRepo) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	objectId, err := bson.ObjectIDFromHex(req.Id)
	if err != nil {
		return nil, common.Error(constants.InvalidUserIDFormat, err)
	}

	user := req.GetUser()
	if user == nil {
		return nil, common.ErrorInvalidArgument("missing user update payload", nil)
	}

	mongo, err := r.db()
	if err != nil {
		return nil, common.ErrorInternal(constants.FailedToConnectDatabase, err)
	}

	now := time.Now()
	userDoc := bson.M{
		"isactive":  user.Isactive,
		"updatedat": now,
	}
	if user.Email != "" {
		userDoc["email"] = user.Email
	}
	if user.Updatedby != "" {
		userDoc["updateby"] = user.Updatedby
	}

	filter := bson.M{"_id": objectId}
	if err = mongo.UpdateOneSet(ctx, UserCollection(), filter, userDoc); err != nil {
		return nil, common.ErrorUpdate(err)
	}

	if user.Password != "" {
		if len(user.Password) < 8 {
			return nil, common.ErrorInvalidArgument("password must be at least 8 characters", nil)
		}
		passwordHash, err := security.PleaseHash(user.Password, security.DefaultParams())
		if err != nil {
			return nil, common.Error("failed to hash password", err)
		}

		var existingUser models.User
		if err := mongo.FindOne(ctx, &existingUser, filter, UserCollection()); err == nil && len(existingUser.UserDetailsIds) > 0 {
			detailsFilter := bson.M{"_id": bson.M{"$in": existingUser.UserDetailsIds}, "isdefault": true}
			detailsUpdate := bson.M{"password": passwordHash, "updatedat": now}
			if user.Updatedby != "" {
				detailsUpdate["updateby"] = user.Updatedby
			}
			_ = mongo.UpdateOneSet(ctx, UserDetailsCollection(), detailsFilter, detailsUpdate)
		}
	}

	return &pb.UpdateUserResponse{
		Success: true,
	}, nil
}

func (r *UserRepo) ActivateUser(ctx context.Context, id string, updatedBy string) error {
	objectId, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return common.Error(constants.InvalidUserIDFormat, err)
	}

	mongo, err := r.db()
	if err != nil {
		return common.ErrorInternal(constants.FailedToConnectDatabase, err)
	}

	filter := bson.M{"_id": objectId}
	update := bson.M{
		"isactive":  true,
		"updateby":  updatedBy,
		"updatedat": time.Now(),
	}
	if err := mongo.UpdateOneSet(ctx, UserCollection(), filter, update); err != nil {
		return common.ErrorUpdate(err)
	}
	return nil
}

func (r *UserRepo) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	// delete from database
	// Convert string ID to MongoDB ObjectID using v2 driver
	objectId, err := bson.ObjectIDFromHex(req.Id)
	if err != nil {
		return nil, common.Error(constants.InvalidUserIDFormat, err)
	}
	filter := bson.M{"_id": objectId}
	mongo, err := r.db()
	if err != nil {
		return nil, common.ErrorInternal(constants.FailedToConnectDatabase, err)
	}

	if err = mongo.DeleteOne(ctx, UserCollection(), filter); err != nil {
		return nil, common.ErrorDelete(err)
	}

	return &pb.DeleteUserResponse{
		Success: true,
	}, nil
}

func (r *UserRepo) LoginUser(ctx context.Context, req *pb.LoginRequest) (*pb.User, *models.UserDetails, error) {
	// get from database
	mongo, err := r.db()
	if err != nil {
		r.logger.Error(constants.FailedToConnectDatabase, err)
		return nil, nil, common.ErrorInternal(constants.FailedToConnectDatabase, err)
	}

	userDoc := bson.M{"email": req.Email}
	userdoc := &models.User{}
	err = mongo.FindOne(ctx, userdoc, userDoc, UserCollection())
	if err != nil {
		return nil, nil, common.Error("invalid email or password", err)
	}

	if !userdoc.Isactive {
		return nil, nil, common.Error("account is not active, please verify your email first", nil)
	}

	userdetailsDoc := &models.UserDetails{}
	if err = mongo.FindOne(ctx, userdetailsDoc, bson.M{"_id": bson.M{"$in": userdoc.UserDetailsIds}, "isdefault": true}, UserDetailsCollection()); err != nil {
		return nil, nil, common.Error("invalid email or password", err)
	}

	pbUser := userdoc.ToProto()
	verified, err := security.Verify(req.Password, userdetailsDoc.Password)
	if err != nil {
		return nil, nil, common.Error("failed to verify password", err)
	}
	if !verified {
		r.logger.Error("msg", "invalid password")
		return nil, nil, common.Error("", errors.New(constants.FailedToLoginUser))
	}

	return pbUser, userdetailsDoc, nil
}

func (r *UserRepo) GetUserByEmail(ctx context.Context, email string) (*pb.User, *pb.UserDetails, error) {
	mongo, err := r.db()
	if err != nil {
		return nil, nil, common.ErrorInternal(constants.FailedToConnectDatabase, err)
	}

	userDoc := bson.M{"email": email}
	usermodel := &models.User{}
	if err = mongo.FindOne(ctx, usermodel, userDoc, UserCollection()); err != nil {
		return nil, nil, common.Error("unknown email", err)
	}

	userdetailsmodel := &models.UserDetails{}
	if err := mongo.FindOne(ctx, userdetailsmodel, bson.M{"_id": bson.M{"$in": usermodel.UserDetailsIds}, "isdefault": true}, UserDetailsCollection()); err != nil {
		return nil, nil, common.Error("unknown email", err)
	}

	return usermodel.ToProto(), userdetailsmodel.ToProto(), nil
}

func (r *UserRepo) GetNotificationIntegrationsById(ctx context.Context, req *pb.NotificationIntegrationsRequest) (models.NotificationIntegrations, error) {
	notifIntegrationId := common.ParseObjectID(req.Id)
	notifIntegration := models.NotificationIntegrations{}
	mongo, err := r.db()
	if err != nil {
		return notifIntegration, common.ErrorInternal(constants.FailedToConnectDatabase, err)
	}

	if err = mongo.FindOne(ctx, &notifIntegration, bson.M{"_id": notifIntegrationId}, NotificationIntegrationsCollection()); err != nil {
		return notifIntegration, common.Error("unknown notification integration", err)
	}

	return notifIntegration, nil
}

func (r *UserRepo) GetNotificationEventsByName(ctx context.Context, req *pb.NotificationEventsRequest) (*pb.NotificationEvents, error) {
	notifEvent := models.NotificationEvents{}
	mongo, err := r.db()
	if err != nil {
		return nil, common.ErrorInternal(constants.FailedToConnectDatabase, err)
	}

	if err = mongo.FindOne(ctx, &notifEvent, bson.M{"name": req.Name}, NotificationEventsCollection()); err != nil {
		return nil, common.Error("unknown notification event", err)
	}
	return notifEvent.ToProto(), nil
}
