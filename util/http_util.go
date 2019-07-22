package util

import (
	"github.com/api-server/types"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"image"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	CurrentUser = "currentUser"
)

func ExactPage(c *gin.Context) types.Page {
	startPage, _ := strconv.Atoi(c.Query("page"))
	pageCount, _ := strconv.Atoi(c.Query("limit"))
	if pageCount <= 0 {
		pageCount = 10
	} else if pageCount > 50 {
		pageCount = 50
	}
	if startPage == 0 {
		startPage = 1
	}
	page := types.Page{Page: int64(startPage),
		Limit: int64(pageCount),
	}
	return page
}

func BuildTokenClaims(claims jwt.MapClaims, exp int64, user types.User) {
	claims["jti"] = time.Now().UnixNano()
	claims["exp"] = exp
	claims["iat"] = time.Now().Unix()
	claims["iss"] = "GoVue"
	claims["sub"] = user.Id.Hex()
	claims["firstName"] = user.FirstName
	claims["middleName"] = user.MiddleName
	claims["lastName"] = user.LastName
	claims["email"] = user.Email
	claims["roles"] = strings.Join(user.Roles, ",")
	claims["isAdmin"] = user.IsAdmin
}

func RestoreContextWithUser(c *gin.Context, claims jwt.MapClaims) {
	id, _ := primitive.ObjectIDFromHex(claims["sub"].(string))
	user := types.User{
		Email:      claims["email"].(string),
		Id:         id,
		IsAdmin:    claims["isAdmin"].(bool),
		FirstName:  claims["firstName"].(string),
		MiddleName: claims["middleName"].(string),
		LastName:   claims["lastName"].(string),
		Roles:      strings.Split(claims["roles"].(string), ","),
	}
	c.Set(CurrentUser, &user)
}

func GetCurrentUser(c *gin.Context) *types.User {
	obj, ok := c.Get(CurrentUser)
	if ok {
		return obj.(*types.User)
	}
	return nil
}

func CreateDirIfMissing(dir string) {
	_, err := os.Stat(dir)
	if err != nil {
		os.MkdirAll(dir, 0777)
	}
}

func DeleteFile(targetPath string) {
	glog.Infoln("delete file: ", targetPath)
	os.Remove(targetPath)
}

func CalcSize(img image.Image, max int) (w, h int) {
	b := img.Bounds()
	w = b.Max.X
	h = b.Max.Y
	if w > h {
		w = max
		h = 0
	} else {
		w = 0
		h = max
	}
	return
}
