package models

import "time"

type AuditTrack struct {
	CreateBy  string    `bson:"createby" json:"createby"`
	CreatedAt time.Time `bson:"createdat" json:"createdat"`
	UpdateBy  string    `bson:"updateby" json:"updateby"`
	UpdatedAt time.Time `bson:"updatedat" json:"updatedat"`
}

