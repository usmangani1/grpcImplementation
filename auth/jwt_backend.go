package auth

import (
	"bufio"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/api-server/settings"
	"github.com/api-server/types"
	"github.com/api-server/util"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

type JWTAuthenticationBackend struct {
	privateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
}

var (
	InvalidTokenCleanerQuitChans = make(chan bool)
)

func init() {
	glog.Info("init JWT auth backend...")
	authBackendInstance = &JWTAuthenticationBackend{
		privateKey: getPrivateKey(),
		PublicKey:  getPublicKey(),
	}

	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for {
			select {
			case <-ticker.C:
				glog.Info("InvalidTokenCleaner working...")
				logoutMap.Range(func(key, value interface{}) bool {
					now := time.Now()
					expired := value.(time.Time)
					if expired.Before(now) {
						logoutMap.Delete(key)
					}
					return true
				})
			case <-InvalidTokenCleanerQuitChans:
				ticker.Stop()
				glog.Info("InvalidTokenCleaner stopped")
				return
			}
		}
	}()
}

var authBackendInstance *JWTAuthenticationBackend = nil
var logoutMap = &sync.Map{}

func JwtAuthenticationBackend() *JWTAuthenticationBackend {
	return authBackendInstance
}

func (backend *JWTAuthenticationBackend) GenerateToken(user *types.User) (string, error) {
	token := jwt.New(jwt.SigningMethodRS512)
	exp := time.Now().Add(time.Hour * time.Duration(settings.Get().JWTExpirationDelta)).Unix()
	claims := token.Claims.(jwt.MapClaims)
	util.BuildTokenClaims(claims, exp, *user)
	tokenString, err := token.SignedString(backend.privateKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (backend *JWTAuthenticationBackend) Logout(tokenId float64, exp float64) error {
	logoutMap.Store(tokenId, time.Unix(int64(exp), 0))
	return nil
}

func (backend *JWTAuthenticationBackend) IsInBlacklist(token float64) bool {
	_, ok := logoutMap.Load(token)
	if !ok {
		return false
	}

	return true
}

func getPrivateKey() *rsa.PrivateKey {
	pkPath := settings.Get().PrivateKeyPath
	glog.Infof("PK path: %s", pkPath)
	privateKeyFile, err := os.Open(pkPath)
	if err != nil {
		panic(err)
	}

	pemFileInfo, _ := privateKeyFile.Stat()
	var size = pemFileInfo.Size()
	pemBytes := make([]byte, size)

	buffer := bufio.NewReader(privateKeyFile)
	_, err = buffer.Read(pemBytes)

	data, _ := pem.Decode([]byte(pemBytes))

	privateKeyFile.Close()

	privateKeyImported, err := x509.ParsePKCS1PrivateKey(data.Bytes)

	if err != nil {
		panic(err)
	}

	return privateKeyImported
}

func getPublicKey() *rsa.PublicKey {
	publicKeyFile, err := os.Open(settings.Get().PublicKeyPath)
	if err != nil {
		panic(err)
	}

	pemFileInfo, _ := publicKeyFile.Stat()
	var size = pemFileInfo.Size()
	pemBytes := make([]byte, size)

	buffer := bufio.NewReader(publicKeyFile)
	_, err = buffer.Read(pemBytes)

	data, _ := pem.Decode([]byte(pemBytes))

	publicKeyFile.Close()

	publicKeyImported, err := x509.ParsePKIXPublicKey(data.Bytes)

	if err != nil {
		panic(err)
	}

	rsaPub, ok := publicKeyImported.(*rsa.PublicKey)

	if !ok {
		panic(err)
	}

	return rsaPub
}

func Permission() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := util.GetCurrentUser(c)
		var accessible = false
		if user.IsAdmin {
			accessible = true
		} else {
			start := time.Now()
			uri := c.Request.RequestURI
			permList := settings.Get().Permissions
			determineFinished := false
			for _, p := range permList {
				//1. judge action
				action := strings.ToLower(c.Request.Method)
				if action != p.Action && p.Action != "any" {
					// action not match, continue
					continue
				}

				//2. judge url
				rgx, err := regexp.Compile(p.UrlRgx)
				if err != nil {
					glog.Errorln("permission urlRegex configuration[" + p.UrlRgx + "] is invalid regex")
					// if there error, just forbidden user to access
					continue
				}
				if rgx.MatchString(uri) {
					switch p.Type {
					case "role":
						if contains(user.Roles, p.Name) {
							// current user has the role
							accessible = "allow" == p.Effect
							determineFinished = true
						}
					case "user":
						if user.Email == p.Name {
							// current user is the user
							accessible = "allow" == p.Effect
							determineFinished = true
						}
					}
				}
				if determineFinished {
					break
				}
			}
			glog.Infof("%s access [%s], accessible: %t\n", user.Email, uri, accessible)
			glog.Infof("permission determine time is: %s\n", time.Since(start).String())
		}

		if !accessible {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "Forbidden access"})
			return
		}

		c.Next()
	}
}

func Required() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := RestoreToken(c)
		valid := err == nil && token.Valid
		if valid {
			claims := token.Claims.(jwt.MapClaims)
			if IsValidToken(claims["jti"].(float64)) {
				c.Set("goVueClaims", claims)
				util.RestoreContextWithUser(c, claims)
			} else {
				valid = false
			}
		}

		if !valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "invalid access token"})
			return
		}
		c.Next()
	}
}

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func IsValidToken(tokenId float64) bool {
	var backend = JwtAuthenticationBackend()
	return !backend.IsInBlacklist(tokenId)
}

func RestoreToken(c *gin.Context) (*jwt.Token, error) {
	var backend = JwtAuthenticationBackend()
	tokenStr := c.GetHeader("Authorization")
	tokenType := "Bearer "
	if tokenStr == "" {
		tokenStr = c.Query("token")
		tokenType = ""
	}

	if tokenStr == "" {
		tokenStr = c.PostForm("Authorization")
	}

	if tokenStr == "" {
		tokenStr = c.PostForm("authorization")
	}

	if tokenStr == "" {
		return nil, errors.New("invalid access token")
	}
	tokenStr = tokenStr[len(tokenType):]
	return jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		} else {
			return backend.PublicKey, nil
		}
	})
}

func SignToken(requestUser *types.User) (string, error) {
	authBackend := JwtAuthenticationBackend()
	return authBackend.GenerateToken(requestUser)
}

func Logout(c *gin.Context) bool {
	authBackend := JwtAuthenticationBackend()
	goVueClaims, ok := c.Get("goVueClaims")
	if !ok {
		return false
	}

	claims, ok := goVueClaims.(jwt.MapClaims)
	if !ok {
		return false
	}
	err := authBackend.Logout(claims["jti"].(float64), claims["exp"].(float64))
	if err != nil {
		glog.Errorln(err)
		return false
	}
	return true
}

func contains(slice []string, item string) bool {
	if len(slice) == 0 {
		return false
	}

	for _, v := range slice {
		if v == item {
			return true
		}
	}

	return false
}
