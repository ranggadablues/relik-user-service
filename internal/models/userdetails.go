package models

import (
	"github.com/ranggadablues/gosok/common"
	userv1 "github.com/ranggadablues/relik-proto-library/gen/go/relik/user/v1"
	"github.com/ranggadablues/relik-proto-library/user-proto/pb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserDetails struct {
	Id                            bson.ObjectID `bson:"_id" json:"id"`
	IsDefault                     bool          `bson:"isdefault" json:"isdefault"`
	Name                          string        `bson:"name" json:"name"`
	Password                      string        `bson:"password" json:"-"`
	Gender                        string        `bson:"gender" json:"gender"`
	Address                       string        `bson:"address" json:"address"`
	Zipcode                       string        `bson:"zipcode" json:"zipcode"`
	Phone                         string        `bson:"phone" json:"phone"`
	City                          string        `bson:"city" json:"city"`
	State                         string        `bson:"state" json:"state"`
	Country                       string        `bson:"country" json:"country"`
	DOB                           string        `bson:"dob" json:"dob"`
	TermOfServiceAndPrivacyPolicy bool          `bson:"termofserviceandprivacypolicy" json:"termofserviceandprivacypolicy"`
	Subscribenewsletter           bool          `bson:"subscribenewsletter" json:"subscribenewsletter"`
	AuditTrack                    `bson:",inline"`
}

func (ud *UserDetails) ToProto() *pb.UserDetails {
	return &pb.UserDetails{
		Id:                            ud.Id.Hex(),
		Isdefault:                     ud.IsDefault,
		Name:                          ud.Name,
		Gender:                        ud.Gender,
		Address:                       ud.Address,
		Zipcode:                       ud.Zipcode,
		Phone:                         ud.Phone,
		City:                          ud.City,
		State:                         ud.State,
		Country:                       ud.Country,
		Dob:                           ud.DOB,
		Termofserviceandprivacypolicy: ud.TermOfServiceAndPrivacyPolicy,
		Subscribenewsletter:           ud.Subscribenewsletter,
		Createby:                      ud.CreateBy,
		Createdat:                     timestamppb.New(ud.CreatedAt),
		Updateby:                      ud.UpdateBy,
		Updatedat:                     timestamppb.New(ud.UpdatedAt),
	}
}

func (ud *UserDetails) FromProto(user *pb.UserDetails) {
	ud.Id = common.ParseObjectID(user.Id)
	ud.IsDefault = user.Isdefault
	ud.Name = user.Name
	ud.Password = user.Password
	ud.Gender = user.Gender
	ud.Address = user.Address
	ud.Zipcode = user.Zipcode
	ud.Phone = user.Phone
	ud.City = user.City
	ud.State = user.State
	ud.Country = user.Country
	ud.DOB = user.Dob
	ud.TermOfServiceAndPrivacyPolicy = user.Termofserviceandprivacypolicy
	ud.Subscribenewsletter = user.Subscribenewsletter
	ud.CreateBy = user.Createby
	if user.Createdat != nil {
		ud.CreatedAt = user.Createdat.AsTime()
	}
	ud.UpdateBy = user.Updateby
	if user.Updatedat != nil {
		ud.UpdatedAt = user.Updatedat.AsTime()
	}
}

func (ud *UserDetails) ToProtoV1() *userv1.UserDetails {
	return &userv1.UserDetails{
		Id:                         ud.Id.Hex(),
		IsDefault:                  ud.IsDefault,
		Name:                       ud.Name,
		Gender:                     ud.Gender,
		Address:                    ud.Address,
		Zipcode:                    ud.Zipcode,
		Phone:                      ud.Phone,
		City:                       ud.City,
		State:                      ud.State,
		Country:                    ud.Country,
		Dob:                        ud.DOB,
		TermsAndPrivacyAccepted:    ud.TermOfServiceAndPrivacyPolicy,
		SubscribeNewsletter:        ud.Subscribenewsletter,
		CreatedBy:                  ud.CreateBy,
		CreatedAt:                  timestamppb.New(ud.CreatedAt),
		UpdatedBy:                  ud.UpdateBy,
		UpdatedAt:                  timestamppb.New(ud.UpdatedAt),
	}
}

func (ud *UserDetails) FromProtoV1(user *userv1.UserDetails) {
	ud.Id = common.ParseObjectID(user.Id)
	ud.IsDefault = user.IsDefault
	ud.Name = user.Name
	ud.Password = user.Password
	ud.Gender = user.Gender
	ud.Address = user.Address
	ud.Zipcode = user.Zipcode
	ud.Phone = user.Phone
	ud.City = user.City
	ud.State = user.State
	ud.Country = user.Country
	ud.DOB = user.Dob
	ud.TermOfServiceAndPrivacyPolicy = user.TermsAndPrivacyAccepted
	ud.Subscribenewsletter = user.SubscribeNewsletter
	ud.CreateBy = user.CreatedBy
	if user.CreatedAt != nil {
		ud.CreatedAt = user.CreatedAt.AsTime()
	}
	ud.UpdateBy = user.UpdatedBy
	if user.UpdatedAt != nil {
		ud.UpdatedAt = user.UpdatedAt.AsTime()
	}
}

func (ud *UserDetails) Collection() string {
	return "cl_userdetails"
}
