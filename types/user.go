package types

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type User struct {
	Id                primitive.ObjectID `json:"id"    bson:"_id,omitempty"`
	Email             string             `json:"email" bson:"email"`
	Password          string             `json:"password"`
	FirstName         string             `json:"firstName,omitempty" bson:"firstName, omitempty"`
	MiddleName        string             `json:"middleName,omitempty" bson:"middleName, omitempty"`
	LastName          string             `json:"lastName,omitempty" bson:"lastName, omitempty"`
	Birthday          time.Time          `json:"birthday,omitempty" bson:"birthday, omitempty"` // dd/mm/yyyy
	Avatar            string             `json:"avatar,omitempty" bson:"avatar, omitempty"`
	OtherImages       []string           `json:"otherImages,omitempty" bson:"otherImages, omitempty""`
	VideoUrlOfYoutube string             `json:"videoUrlOfYoutube,omitempty" bson:"videoUrlOfYoutube, emitempty"`
	IsAdmin           bool               `json:"isAdmin,omitempty" bson:"isAdmin, emitempty"`
	Roles             []string           `json:"roles,omitempty" bson:"roles, emitempty"`
}
