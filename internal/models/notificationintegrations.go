package models

import (
	"github.com/ranggadablues/gosok/common"
	notificationv1 "github.com/ranggadablues/relik-proto-library/gen/go/relik/notification/v1"
	"github.com/ranggadablues/relik-proto-library/user-proto/pb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type NotificationIntegrations struct {
	Id          bson.ObjectID `bson:"_id" json:"id"`
	Type        string        `bson:"type" json:"type"`
	Provider    string        `bson:"provider,omitempty" json:"provider,omitempty"`
	SenderName  string        `bson:"sendername,omitempty" json:"sendername,omitempty"`
	SenderEmail string        `bson:"senderemail,omitempty" json:"senderemail,omitempty"`

	// OAuth
	OAuthClientID     string `bson:"oauthclientid,omitempty" json:"oauthclientid,omitempty"`
	OAuthClientSecret string `bson:"oauthclientsecret,omitempty" json:"-"`
	OAuthRefreshToken string `bson:"oauthrefreshtoken,omitempty" json:"oauthrefreshtoken,omitempty"`
	OAuthRedirectURL  string `bson:"oauthredirecturl,omitempty" json:"oauthredirecturl,omitempty"`

	// SMTP
	SMTPHost     string `bson:"smtphost,omitempty" json:"smtphost,omitempty"`
	SMTPPort     int    `bson:"smtpport,omitempty" json:"smtpport,omitempty"`
	SMTPUsername string `bson:"smtpusername,omitempty" json:"smtpusername,omitempty"`
	SMTPPassword string `bson:"smtppassword,omitempty" json:"-"`

	AuditTrack `bson:",inline"`
}

func (i *NotificationIntegrations) ToProto() *pb.NotificationIntegrations {
	return &pb.NotificationIntegrations{
		Id:           i.Id.Hex(),
		Type:         i.Type,
		Provider:     i.Provider,
		Clientid:     i.OAuthClientID,
		Clientsecret: i.OAuthClientSecret,
		Refreshtoken: i.OAuthRefreshToken,
		Redirecturl:  i.OAuthRedirectURL,
		Smtphost:     i.SMTPHost,
		Smtpport:     int32(i.SMTPPort),
		Smtpusername: i.SMTPUsername,
		Smtppassword: i.SMTPPassword,
		Createby:     i.CreateBy,
		Createdat:    timestamppb.New(i.CreatedAt),
		Updateby:     i.UpdateBy,
		Updatedat:    timestamppb.New(i.UpdatedAt),
	}
}

func (i *NotificationIntegrations) FromProto(integration *pb.NotificationIntegrations) {
	i.Id = common.ParseObjectID(integration.Id)
	i.Type = integration.Type
	i.Provider = integration.Provider
	i.OAuthClientID = integration.Clientid
	i.OAuthClientSecret = integration.Clientsecret
	i.OAuthRefreshToken = integration.Refreshtoken
	i.OAuthRedirectURL = integration.Redirecturl
	i.SMTPHost = integration.Smtphost
	i.SMTPPort = int(integration.Smtpport)
	i.SMTPUsername = integration.Smtpusername
	i.SMTPPassword = integration.Smtppassword
	i.CreateBy = integration.Createby
	if integration.Createdat != nil {
		i.CreatedAt = integration.Createdat.AsTime()
	}
	i.UpdateBy = integration.Updateby
	if integration.Updatedat != nil {
		i.UpdatedAt = integration.Updatedat.AsTime()
	}
}

func (i *NotificationIntegrations) ToProtoV1() *notificationv1.NotificationIntegration {
	return &notificationv1.NotificationIntegration{
		Id:           i.Id.Hex(),
		Type:         i.Type,
		Provider:     i.Provider,
		SenderName:   i.SenderName,
		SenderEmail:  i.SenderEmail,
		ClientId:     i.OAuthClientID,
		ClientSecret: i.OAuthClientSecret,
		RefreshToken: i.OAuthRefreshToken,
		RedirectUrl:  i.OAuthRedirectURL,
		SmtpHost:     i.SMTPHost,
		SmtpPort:     int32(i.SMTPPort),
		SmtpUsername: i.SMTPUsername,
		SmtpPassword: i.SMTPPassword,
		CreatedBy:    i.CreateBy,
		CreatedAt:    timestamppb.New(i.CreatedAt),
		UpdatedBy:    i.UpdateBy,
		UpdatedAt:    timestamppb.New(i.UpdatedAt),
	}
}

func (i *NotificationIntegrations) FromProtoV1(integration *notificationv1.NotificationIntegration) {
	i.Id = common.ParseObjectID(integration.Id)
	i.Type = integration.Type
	i.Provider = integration.Provider
	i.SenderName = integration.SenderName
	i.SenderEmail = integration.SenderEmail
	i.OAuthClientID = integration.ClientId
	i.OAuthClientSecret = integration.ClientSecret
	i.OAuthRefreshToken = integration.RefreshToken
	i.OAuthRedirectURL = integration.RedirectUrl
	i.SMTPHost = integration.SmtpHost
	i.SMTPPort = int(integration.SmtpPort)
	i.SMTPUsername = integration.SmtpUsername
	i.SMTPPassword = integration.SmtpPassword
	i.CreateBy = integration.CreatedBy
	if integration.CreatedAt != nil {
		i.CreatedAt = integration.CreatedAt.AsTime()
	}
	i.UpdateBy = integration.UpdatedBy
	if integration.UpdatedAt != nil {
		i.UpdatedAt = integration.UpdatedAt.AsTime()
	}
}

func (i *NotificationIntegrations) Collection() string {
	return "mt_notificationintegrations"
}
