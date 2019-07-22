package mongo

import (
	"context"
	"fmt"
	"github.com/api-server/settings"
	"github.com/golang/glog"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"strings"
	"time"
)

var client *mongo.Client

//InitMongo connect to mongodb
func InitMongo() {
	dialInfo := settings.Get().Mongo
	var err error
	mongoUrl := fmt.Sprintf(dialInfo.Url, dialInfo.Username, dialInfo.Password)
	printUrl := strings.Replace(mongoUrl, dialInfo.Password, "******", 1)
	glog.Infoln("connection url:", printUrl)
	client, err = mongo.NewClient(options.Client().ApplyURI(mongoUrl))
	if err != nil {
		panic(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(dialInfo.Timeout)*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		panic(err.Error())
	}

	go func(c *mongo.Client) {
		/*		ticker := time.NewTicker(10 * time.Second)
				for _ = range ticker.C {*/
		var fr readpref.ReadPref
		glog.Errorln(client.Ping(nil, &fr))
		if &fr != nil {
			glog.Infoln("ping ...")
			glog.Infoln(fr.Mode())
			glog.Infoln(fr.MaxStaleness())
		}
		//}

	}(client)
	glog.Infoln("connect mongodb successfully")
}

func GetClient() *mongo.Client {
	return client
}

func IsDupErr(err error) bool {
	return strings.Contains(err.Error(), "E11000 duplicate key error")
}

func IsNotFoundError(err error) bool {
	return strings.Contains(err.Error(), "mongo: no documents in result")
}
