package main

import (
	"flag"
	"github.com/api-server/auth"
	"github.com/api-server/controllers"
	"github.com/api-server/core/mongo"
	"github.com/api-server/settings"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	flag.Parse()
	os.Setenv("GIN_MODE", "release")
	workPath := strings.Split(os.Getenv("GOPATH"), ";")[0]
	env := os.Getenv("GO_ENV")
	if env == "production" {
		workPath, _ = path.Split(os.Args[0])
	}
	glog.Infoln("work dir: ", workPath)
	mongo.InitMongo()
	gin.DisableConsoleColor()
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.MaxMultipartMemory = 8 << 2
	userFileController := controllers.UserFileController{}
	userController := controllers.UserController{}
	router.Use(static.Serve("/", static.LocalFile(filepath.Join(settings.Get().UploadDir, "web"), false)))
	router.Use(auth.Cors())
	router.POST("/signup", userController.SignUp)
	router.POST("/signin", userController.SignIn)
	// need auth information[access token]
	authGroup := router.Group("/")
	authGroup.Use(auth.Required())
	{
		router.POST("/signout", userController.SignOut)
		router.POST("/refresh-token", userController.RefreshToken)
	}

	v1 := router.Group("/api/v1/users")
	v1.Use(auth.Required())
	v1.Use(auth.Permission())
	{
		v1.GET("/", userController.ListUsers)
		v1.GET("/:uid", userController.FetchUser)
		v1.PUT("/:uid", userController.ModifyUserProfile)
		v1.GET("/:uid/files/", userFileController.List)
		v1.GET("/:uid/files/:fid", userFileController.Read)
		v1.POST("/:uid/files/", userFileController.Add)
		v1.PUT("/:uid/files/:fid", userFileController.Modify)
		v1.DELETE("/:uid/files/:fid", userFileController.Remove)
	}

	//router.NoRoute(func(c *gin.Context) {
	//	c.Redirect(http.StatusFound, "/")
	//})
	s := &http.Server{
		Addr:           ":8080",
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	glog.Errorln(s.ListenAndServe())
}
