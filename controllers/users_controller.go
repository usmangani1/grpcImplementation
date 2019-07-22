package controllers

import (
	"fmt"
	"github.com/api-server/core/mongo"
	"github.com/api-server/services"
	"github.com/api-server/settings"
	"github.com/api-server/types"
	"github.com/api-server/util"
	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	jsoniter "github.com/json-iterator/go"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type UserController struct {
}

func (uc UserController) RefreshToken(c *gin.Context) {
	//user := util.GetCurrentUser(c)
	//token, err := auth.SignToken(user)
	//if err != nil {
	//	glog.Errorln(err)
	//	c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
	//	return
	//}
	c.JSON(http.StatusOK, gin.H{"token": ""})
}

func (uc UserController) ListUsers(c *gin.Context) {
	page := util.ExactPage(c)
	users := services.LoadAllUsers(page, nil)
	c.JSON(http.StatusOK, users)
}

func ListUsersdef(){
	page := types.Page{Page: int64(0),
		Limit: int64(10),
	}
	users := services.LoadAllUsers(page, nil)
	fmt.Println(users)
}

func (uc UserController) SignOut(c *gin.Context) {
	//auth.Logout(c)
	c.JSON(http.StatusOK, gin.H{"message": "logout successfully"})
}

func (uc UserController) SignIn(c *gin.Context) {
	var user types.User

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Bad json found"})
		return
	}
	u, err := services.Login(user.Email, user.Password)

	if err != nil {
		if mongo.IsNotFoundError(err) {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "user or password incorrect"})
		} else {
			glog.Errorln(err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "database error"})
		}
		return
	}

	//token, err := auth.SignToken(&u)
	//if err != nil {
	//	glog.Errorln(err)
	//	c.JSON(http.StatusInternalServerError, gin.H{"message": "sign access token error: " + err.Error()})
	//	return
	//}
	c.JSON(http.StatusOK, gin.H{"user": u, "token": ""})
}

func (uc UserController) SignUp(c *gin.Context) {
	var user types.User

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Bad json found"})
		return
	}

	_, err := services.GetByUserName(user.Email)
	if err == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "user exists"})
		glog.Errorln(err)
		return
	} else {
		if !mongo.IsNotFoundError(err) {
			glog.Errorln(err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "database error"})
			return
		}
	}

	u, err := services.CreateUser(user)

	if err != nil {
		if mongo.IsDupErr(err) {
			c.JSON(http.StatusBadRequest, gin.H{"message": "email exists"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "database error"})
		}
		glog.Errorln(err)
		return
	}

	//token, err := auth.SignToken(&u)
	//if err != nil {
	//	glog.Errorln(err)
	//	c.JSON(http.StatusInternalServerError, gin.H{"message": "sign access token error: " + err.Error()})
	//	return
	//}
	c.JSON(http.StatusCreated, gin.H{"user": u, "token": ""})
}

func (uc UserController) FetchUser(c *gin.Context) {
	hexId := c.Param("uid")
	uid, err := primitive.ObjectIDFromHex(hexId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid user id found[" + hexId + "]"})
		return
	}
	user, err := services.GetUserById(uid)
	if err != nil {
		if mongo.IsNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("user[id=%s] not found", hexId)})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "database error"})
		}
		return
	}
	c.JSON(http.StatusOK, user)
}

func (uc UserController) ModifyUserProfile(c *gin.Context) {
	var currentUser = util.GetCurrentUser(c)
	var user types.User
	/*
		type UserProfileForm struct {
			Avatar      *multipart.FileHeader `form:"avatar" binding:"-"`
			OtherImages []*multipart.FileHeader `form:"otherImages" binding:"-"`
			User        string                `form:"user" binding:"required"`
		}
		form := UserProfileForm{}
		if err := c.ShouldBind(&form); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Bad form data found"})
			return
		}*/

	userJson := c.PostForm("user")
	if strings.TrimSpace(userJson) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Bad form data field 'user' found"})
		return
	}
	mForm, err := c.MultipartForm()
	if err != nil {
		glog.Errorln(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	avatarFiles := mForm.File["avatar"]
	otherImages := mForm.File["otherImages"]
	err = jsoniter.Unmarshal([]byte(userJson), &user)
	if err != nil {
		glog.Errorln(err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Bad form data field 'user' found"})
		return
	}

	user.Id = currentUser.Id
	userId := currentUser.Id.Hex()
	if len(avatarFiles) > 0 {
		avatar := avatarFiles[0]
		// process avatar
		avatarDir := filepath.Join(settings.Get().UploadDir, "web/public", "avatar")
		util.CreateDirIfMissing(avatarDir)
		avatarName := userId + filepath.Ext(avatar.Filename)
		targetFile := filepath.Join(avatarDir, avatarName)
		glog.Infoln("save avatar to", targetFile)
		file, err := avatar.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "read avatar file error: " + err.Error()})
			return
		}
		img, err := imaging.Decode(file)
		file.Close()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid avatar file"})
			return
		}
		w, h := util.CalcSize(img, 48)
		img = imaging.Resize(img, w, h, imaging.Lanczos)
		imaging.Save(img, targetFile)
		//c.SaveUploadedFile(avatar, targetFile)
		user.Avatar = "/public/avatar/" + avatarName
	}

	if len(otherImages) > 0 {
		// process profile images
		profileDir := filepath.Join(settings.Get().UploadDir, "web/public", "profile", userId)
		//clear the profile image for current user
		os.RemoveAll(profileDir)
		util.CreateDirIfMissing(profileDir)
		profileFileFormat := "/public/profile/" + userId + "/%s"
		var profileImages []string
		for _, file := range otherImages {
			//img := otherImages[index]
			name := strconv.Itoa(int(time.Now().UnixNano()))
			ext := filepath.Ext(file.Filename)
			imgName := name + ext
			imagePath := filepath.Join(profileDir, imgName)
			c.SaveUploadedFile(file, imagePath)
			imgUrl := fmt.Sprintf(profileFileFormat, imgName)
			profileImages = append(profileImages, imgUrl)
		}
		user.OtherImages = profileImages
	}
	u, err := services.UpdateUser(user)
	if err != nil {
		if mongo.IsNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("user[id=%s] not found", user.Id.Hex())})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "database server"})
		}
		glog.Error("Failed insert person: ", err)
		return
	}
	c.JSON(http.StatusOK, u)
}

func (uc UserController) DeleteUser(c *gin.Context) {
	userId := c.Param("uid")
	deleted := services.DeleteUser(userId)
	if deleted {
		c.JSON(http.StatusOK, gin.H{"message": "delete user successfully"})
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "delete user failed"})
	}
}

func (uc UserController) GetUserById(c *gin.Context) {
	userId := c.Param("uid")
	uid, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid user id found[" + userId + "]"})
		return
	}
	user, err := services.GetUserById(uid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "user not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}
