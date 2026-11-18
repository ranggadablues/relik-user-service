package models

import (
	"github.com/ranggadablues/gosok/common"
	notificationv1 "github.com/ranggadablues/relik-proto-library/gen/go/relik/notification/v1"
	"github.com/ranggadablues/relik-proto-library/user-proto/pb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type NotificationEvents struct {
	Id      bson.ObjectID `bson:"_id" json:"id"`
	Type    string        `bson:"type" json:"type"`
	Name    string        `bson:"name" json:"name"`
	Subject string        `bson:"subject" json:"subject"`
	Body    string        `bson:"body" json:"body"`

	AuditTrack `bson:",inline"`
}

func (i *NotificationEvents) ToProto() *pb.NotificationEvents {
	return &pb.NotificationEvents{
		Id:        i.Id.Hex(),
		Type:      i.Type,
		Name:      i.Name,
		Subject:   i.Subject,
		Body:      i.Body,
		Createby:  i.CreateBy,
		Createdat: timestamppb.New(i.CreatedAt),
		Updateby:  i.UpdateBy,
		Updatedat: timestamppb.New(i.UpdatedAt),
	}
}

func (i *NotificationEvents) FromProto(integration *pb.NotificationEvents) {
	i.Id = common.ParseObjectID(integration.Id)
	i.Type = integration.Type
	i.Name = integration.Name
	i.Subject = integration.Subject
	i.Body = integration.Body
	i.CreateBy = integration.Createby
	if integration.Createdat != nil {
		i.CreatedAt = integration.Createdat.AsTime()
	}
	i.UpdateBy = integration.Updateby
	if integration.Updatedat != nil {
		i.UpdatedAt = integration.Updatedat.AsTime()
	}
}

func (i *NotificationEvents) ToProtoV1() *notificationv1.NotificationEvent {
	return &notificationv1.NotificationEvent{
		Id:        i.Id.Hex(),
		Type:      i.Type,
		Name:      i.Name,
		Subject:   i.Subject,
		Body:      i.Body,
		CreatedBy: i.CreateBy,
		CreatedAt: timestamppb.New(i.CreatedAt),
		UpdatedBy: i.UpdateBy,
		UpdatedAt: timestamppb.New(i.UpdatedAt),
	}
}

func (i *NotificationEvents) FromProtoV1(integration *notificationv1.NotificationEvent) {
	i.Id = common.ParseObjectID(integration.Id)
	i.Type = integration.Type
	i.Name = integration.Name
	i.Subject = integration.Subject
	i.Body = integration.Body
	i.CreateBy = integration.CreatedBy
	if integration.CreatedAt != nil {
		i.CreatedAt = integration.CreatedAt.AsTime()
	}
	i.UpdateBy = integration.UpdatedBy
	if integration.UpdatedAt != nil {
		i.UpdatedAt = integration.UpdatedAt.AsTime()
	}
}

func (i *NotificationEvents) Collection() string {
	return "mt_notificationevents"
}
