package settings

import (
	"encoding/json"
	"flag"
	"github.com/api-server/util"
	"github.com/golang/glog"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
)

var environments = map[string]string{
	"production":    "settings/prod.json",
	"preproduction": "settings/pre.json",
	"tests":         "../../settings/tests.json",
}

type MongoDialInfo struct {
	Url      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
	Timeout  int    `json:"timeout"`
}

type Permission struct {
	Type   string `json:"type"`
	Name   string `json:"name"`
	Action string `json:"action"`
	UrlRgx string `json:"urlRgx"`
	Effect string `json:"effect"`
}

type Settings struct {
	PrivateKeyPath     string
	PublicKeyPath      string
	WorkDir            string
	UploadDir          string
	JWTExpirationDelta int
	Mongo              *MongoDialInfo
	Permissions        []Permission
}

var settings Settings = Settings{}
var env = "preproduction"

func init() {
	flag.Parse()
	flag.Lookup("logtostderr").Value.Set("true")
	env = os.Getenv("GO_ENV")
	if env == "" {
		glog.Info("Warning: Setting preproduction environment due to lack of GO_ENV value")
		env = "preproduction"
	}
	LoadSettingsByEnv(env)
}

func LoadSettingsByEnv(env string) {
	//workpath := strings.Split(os.Getenv("GOPATH"), ";")[0]
	workpath, err := os.Getwd()
	glog.Infoln("current dir", workpath)
	if err != nil {
		log.Println(err)
		return
	}
	configFile := path.Join(workpath, environments[env])
	glog.Infoln("read config from", configFile)
	content, err := ioutil.ReadFile(configFile)
	if err != nil {
		glog.Info("error while reading config file", err)
	}
	settings = Settings{}
	jsonErr := json.Unmarshal(content, &settings)
	if jsonErr != nil {
		glog.Info("error while parsing config file", jsonErr)
	}
	if env != "production" {
		settings.PrivateKeyPath = path.Join(workpath, settings.PrivateKeyPath)
		settings.PublicKeyPath = path.Join(workpath, settings.PublicKeyPath)
	}
	settings.WorkDir = workpath
	publicFilePath := filepath.Join(settings.UploadDir, "public")
	util.CreateDirIfMissing(publicFilePath)
}

func GetEnvironment() string {
	return env
}

func Get() Settings {
	return settings
}

func IsTestEnvironment() bool {
	return env == "tests"
}
