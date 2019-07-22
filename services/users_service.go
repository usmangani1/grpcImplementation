package services

import (
	"context"
	"errors"
	"github.com/api-server/core/mongo"
	"github.com/api-server/settings"
	"github.com/api-server/types"
	"github.com/api-server/util"
	"github.com/golang/glog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	m2 "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strings"
	"time"
)

var timeout = time.Duration(settings.Get().Mongo.Timeout) * time.Second

func LoadAllUsers(page types.Page, query bson.D) *[]types.User {
	c := getUsersCollection()
	glog.Info(c)
	var users []types.User
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	// exclude password
	cursor, err := c.Find(ctx, query, options.Find().SetProjection(bson.M{"password": 0}))
	if err != nil {
		glog.Error(err)
		return nil
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var u types.User
		err = cursor.Decode(&u)
		if err != nil {
			glog.Errorln(err)
			continue
		}
		users = append(users, u)
	}
	if err = cursor.Err(); err != nil {
		glog.Errorln(err)
	}
	return &users
}


func getUsersCollection() *m2.Collection {
	return mongo.GetClient().Database("api-server").Collection("users")
}

func getFilesCollection() *m2.Collection {
	return mongo.GetClient().Database("api-server").Collection("files")
}

func CreateUser(user types.User) (types.User, error) {
	c := getUsersCollection()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	// default is user
	user.Roles = []string{"user"}
	r, err := c.InsertOne(ctx, user)
	if err != nil {
		return user, err
	}
	user.Id = r.InsertedID.(primitive.ObjectID)
	return user, nil
}

//UpdateUser will only update profile information except email, password, isAdmin
func UpdateUser(user types.User) (types.User, error) {
	c := getUsersCollection()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	update := bson.M{}

	if len(strings.TrimSpace(user.FirstName)) > 0 {
		update["firstName"] = user.FirstName
	}
	if len(strings.TrimSpace(user.MiddleName)) > 0 {
		update["middleName"] = user.MiddleName
	}
	if len(strings.TrimSpace(user.LastName)) > 0 {
		update["lastName"] = user.LastName
	}
	if len(strings.TrimSpace(user.Avatar)) > 0 {
		update["avatar"] = user.Avatar
	}
	update["birthday"] = user.Birthday
	if len(strings.TrimSpace(user.VideoUrlOfYoutube)) > 0 {
		update["videoUrlOfYoutube"] = user.VideoUrlOfYoutube
	}

	if len(user.OtherImages) > 0 {
		update["otherImages"] = user.OtherImages
	}
	_, err := c.UpdateOne(ctx, bson.M{"_id": user.Id}, bson.M{"$set": update})
	if err != nil {
		return user, err
	}
	return GetUserById(user.Id)
}

func DeleteUser(id string) bool {
	c := getUsersCollection()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		glog.Errorln("convert hex id error:", id)
		return false
	}
	_, err = c.DeleteOne(ctx, bson.M{"_id": oid})
	if err != nil {
		glog.Error(err)
	}
	return err == nil
}

func GetUserById(userId primitive.ObjectID) (types.User, error) {
	c := getUsersCollection()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	var user types.User
	err := c.FindOne(ctx, bson.D{{"_id", userId}}, options.FindOne().SetProjection(bson.M{"password": 0})).Decode(&user)
	if err != nil {
		return user, err
	}
	return user, nil
}

func Login(email, password string) (types.User, error) {
	c := getUsersCollection()
	var user types.User
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	err := c.FindOne(ctx, bson.M{"email": email, "password": password}, options.FindOne().SetProjection(bson.M{"password": 0})).Decode(&user)
	return user, err

}

func GetByUserName(userName string) (types.User, error) {
	c := getUsersCollection()
	var user types.User
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	err := c.FindOne(ctx, bson.M{"email": userName}, options.FindOne().SetProjection(bson.M{"password": 0})).Decode(&user)
	return user, err
}

// ##############files##############

func ListAllFiles(userId primitive.ObjectID, page types.Page) ([]types.File, error) {
	c := getFilesCollection()
	var files []types.File
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cursor, err := c.Find(ctx, bson.M{"owner": userId}, options.Find().SetSkip((page.Page-1)*page.Limit).SetLimit(page.Limit))
	if err != nil {
		glog.Error(err)
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var f types.File
		err = cursor.Decode(&f)
		if err != nil {
			glog.Errorln(err)
			continue
		}
		files = append(files, f)
	}
	if err = cursor.Err(); err != nil {
		glog.Errorln(err)
	}
	return files, nil
}

func CreateFile(file types.File) (types.File, error) {
	c := getFilesCollection()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	r, err := c.InsertOne(ctx, file)
	if err != nil {
		return file, err
	}
	file.Id = r.InsertedID.(primitive.ObjectID)
	return file, nil
}

func GetFileById(fileId primitive.ObjectID) (file types.File, err error) {
	c := getFilesCollection()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	err = c.FindOne(ctx, bson.M{"_id": fileId}).Decode(&file)
	if err != nil {
		return
	}
	return
}

func DeleteFile(userId primitive.ObjectID, fileId primitive.ObjectID) error {
	c := getFilesCollection()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	f, err := GetFileById(fileId)
	if err != nil {
		glog.Errorln(err)
		return err
	}
	r, err := c.DeleteOne(ctx, bson.M{"_id": fileId, "owner": userId})
	if err != nil {
		return err
	}
	if r.DeletedCount != 1 {
		return errors.New("not deleted any file")
	}
	if len(f.DiskPath) > 0 {
		lastDot := strings.LastIndex(f.DiskPath, ".")
		preViewPath := f.DiskPath[0:lastDot] + "-pre" + f.DiskPath[lastDot:]
		util.DeleteFile(f.DiskPath)
		util.DeleteFile(preViewPath)
	}
	return nil
}
