package types

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type File struct {
	Id          primitive.ObjectID `json:"id"    bson:"_id,omitempty"`
	Owner       primitive.ObjectID `owner:"id"    bson:"owner,omitempty"`
	Name        string             `json:"name" bson:"name"`
	Type        string             `json:"type"`
	SubType     string             `json:"subType,omitempty" bson:"subType, omitempty"`
	Size        int64              `json:"size,omitempty" bson:"size, omitempty"`
	Path        string             `json:"path,omitempty" bson:"path, omitempty"`
	PreviewPath string             `json:"previewPath,omitempty" bson:"previewPath, omitempty"`
	DiskPath    string             `json:"diskPath,omitempty" bson:"diskPath, omitempty"`
	EntryDate   time.Time          `json:"entryDate,omitempty" bson:"entryDate, omitempty"`
}
