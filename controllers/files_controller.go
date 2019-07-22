package controllers

import (
	"fmt"
	"github.com/api-server/services"
	"github.com/api-server/settings"
	"github.com/api-server/types"
	"github.com/api-server/util"
	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type UserFileController struct {
}

func (f UserFileController) List(c *gin.Context) {
	currentUser := util.GetCurrentUser(c)
	page := util.ExactPage(c)
	files, err := services.ListAllFiles(currentUser.Id, page)
	if err != nil {
		glog.Errorln(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, files)
}

func (f UserFileController) Read(c *gin.Context) {
	name := c.Param("fid")
	isFile := strings.Index(name, ".") > 0
	currentUser := util.GetCurrentUser(c)
	uid := c.Param("uid")

	if uid != currentUser.Id.Hex() {
		// can't access other user's file
		c.JSON(http.StatusForbidden, gin.H{"message": "forbidden to access"})
		return
	}

	if isFile {
		// not a hex id, but a file request
		file := filepath.Join(settings.Get().UploadDir, "files", uid, name)
		c.File(file)
		return
	}

	fid, err := primitive.ObjectIDFromHex(name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid file id found[" + name + "]"})
		return
	}
	file, err := services.GetFileById(fid)
	if err != nil {
		glog.Errorln(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	if uid != file.Owner.Hex() {
		// can't access other user's file
		c.JSON(http.StatusForbidden, gin.H{"message": "forbidden to access"})
		return
	}

	c.JSON(http.StatusOK, file)
}

func (f UserFileController) Add(c *gin.Context) {
	mForm, err := c.MultipartForm()
	if err != nil {
		glog.Errorln(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	files := mForm.File["files"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "no file(s) found"})
		return
	}
	fileType := c.PostForm("type")
	fileSubType := c.PostForm("subType")

	if fileType == "" {
		fileType = "image"
	}
	currentUser := util.GetCurrentUser(c)
	uid := currentUser.Id.Hex()
	fileDir := filepath.Join(settings.Get().UploadDir, "files", uid)
	util.CreateDirIfMissing(fileDir)
	var savedFiles []types.File
	pathFormat := "/api/v1/users/" + uid + "/files/%s"
	for _, file := range files {
		toSave := types.File{}
		toSave.Owner = currentUser.Id
		toSave.Type = fileType
		if fileSubType == "" {
			fileSubType = filepath.Ext(file.Filename)[1:]
		}
		toSave.SubType = fileSubType
		toSave.Name = file.Filename
		toSave.Size = file.Size
		toSave.EntryDate = time.Now()
		name := strconv.Itoa(int(time.Now().UnixNano()))
		ext := filepath.Ext(file.Filename)
		toSaveFileName := name + ext
		toSave.Path = fmt.Sprintf(pathFormat, toSaveFileName)
		diskPath := filepath.Join(fileDir, toSaveFileName)
		toSave.DiskPath = diskPath
		toSaveFileName = name + "-pre" + ext
		toSave.PreviewPath = fmt.Sprintf(pathFormat, toSaveFileName)
		c.SaveUploadedFile(file, diskPath)
		f, err := services.CreateFile(toSave)
		if err != nil {
			glog.Errorln(err)
			break
		}
		img, err := imaging.Open(diskPath)
		w, h := util.CalcSize(img, 80)
		img = imaging.Resize(img, w, h, imaging.Linear)
		diskPath = filepath.Join(fileDir, toSaveFileName)
		imaging.Save(img, diskPath)
		savedFiles = append(savedFiles, f)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "save file error failed"})
		return
	}

	c.JSON(http.StatusOK, savedFiles)
}

func (f UserFileController) Modify(c *gin.Context) {

}

func (f UserFileController) Remove(c *gin.Context) {
	fileHexId := c.Param("fid")
	fid, err := primitive.ObjectIDFromHex(fileHexId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid file id found"})
		return
	}
	currentUser := util.GetCurrentUser(c)
	err = services.DeleteFile(currentUser.Id, fid)
	if err != nil {
		glog.Errorln(err)
		if strings.Contains(err.Error(), "not deleted any file") {
			c.JSON(http.StatusNotFound, gin.H{"message": "not found the file for the user"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "server error: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted file successfully"})
}
