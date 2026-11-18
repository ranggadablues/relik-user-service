package models

import (
	"github.com/ranggadablues/gosok/common"
	userv1 "github.com/ranggadablues/relik-proto-library/gen/go/relik/user/v1"
	"github.com/ranggadablues/relik-proto-library/user-proto/pb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type User struct {
	Id             bson.ObjectID   `bson:"_id" json:"id"`
	Email          string          `bson:"email" json:"email"`
	Isactive       bool            `bson:"isactive" json:"isactive"`
	CreatedSource  string          `bson:"createdsource" json:"createdsource"` // basic, google, facebook, etc
	UserDetailsIds []bson.ObjectID `bson:"userdetailsid" json:"userdetailsid"`
	AuditTrack     `bson:",inline"`
}

func (u *User) ToProto() *pb.User {
	var userDetailsIds []string
	for _, userDetailId := range u.UserDetailsIds {
		userDetailsIds = append(userDetailsIds, userDetailId.Hex())
	}
	return &pb.User{
		Id:             u.Id.Hex(),
		Email:          u.Email,
		Isactive:       u.Isactive,
		Createdsource:  u.CreatedSource,
		Userdetailsids: userDetailsIds,
		Createby:       u.CreateBy,
		Createdat:      timestamppb.New(u.CreatedAt),
		Updateby:       u.UpdateBy,
		Updatedat:      timestamppb.New(u.UpdatedAt),
	}
}

func (u *User) FromProto(user *pb.User) {
	u.Id = common.ParseObjectID(user.Id)
	u.Email = user.Email
	u.Isactive = user.Isactive
	u.CreatedSource = user.Createdsource
	for _, userdetailid := range user.Userdetailsids {
		userDetailId, _ := bson.ObjectIDFromHex(userdetailid)
		u.UserDetailsIds = append(u.UserDetailsIds, userDetailId)
	}
	u.CreateBy = user.Createby
	if user.Createdat != nil {
		u.CreatedAt = user.Createdat.AsTime()
	}
	u.UpdateBy = user.Updateby
	if user.Updatedat != nil {
		u.UpdatedAt = user.Updatedat.AsTime()
	}
}

func (u *User) ToProtoV1() *userv1.User {
	var userDetailsIds []string
	for _, userDetailId := range u.UserDetailsIds {
		userDetailsIds = append(userDetailsIds, userDetailId.Hex())
	}
	return &userv1.User{
		Id:             u.Id.Hex(),
		Email:          u.Email,
		IsActive:       u.Isactive,
		CreatedSource:  u.CreatedSource,
		UserDetailsIds: userDetailsIds,
		CreatedBy:      u.CreateBy,
		CreatedAt:      timestamppb.New(u.CreatedAt),
		UpdatedBy:      u.UpdateBy,
		UpdatedAt:      timestamppb.New(u.UpdatedAt),
	}
}

func (u *User) FromProtoV1(user *userv1.User) {
	u.Id = common.ParseObjectID(user.Id)
	u.Email = user.Email
	u.Isactive = user.IsActive
	u.CreatedSource = user.CreatedSource
	for _, userdetailid := range user.UserDetailsIds {
		userDetailId, _ := bson.ObjectIDFromHex(userdetailid)
		u.UserDetailsIds = append(u.UserDetailsIds, userDetailId)
	}
	u.CreateBy = user.CreatedBy
	if user.CreatedAt != nil {
		u.CreatedAt = user.CreatedAt.AsTime()
	}
	u.UpdateBy = user.UpdatedBy
	if user.UpdatedAt != nil {
		u.UpdatedAt = user.UpdatedAt.AsTime()
	}
}

func (u *User) Collection() string {
	return "cl_users"
}
