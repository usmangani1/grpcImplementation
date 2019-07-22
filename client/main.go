// A grpc Client which accepts the userName and the email as the parameters from the browser
// and sends the request to the grpc server.
// built using grpc and proto3

package main

import (
	"fmt"
	proto "github.com/api-server/proto"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"github.com/golang/glog"
	"net/http"
)


func main(){
	// opening a grpc connection with port 4040
	conn,err := grpc.Dial("localhost:4040",grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	client := proto.NewAddServiceClient(conn)
	// using gin framework which helps the client to pick up the requests.
	g := gin.Default()
	// Method which is used to create the user
	g.GET("/createuser/:name/:email",func(ctx *gin.Context){

		name := ctx.Param("name")
		email := ctx.Param("email")

		req := &proto.Request{Name : name, Email : email}

		if response ,err := client.CreateUser(ctx,req); err == nil {
			ctx.JSON(http.StatusOK,gin.H{
				"Created UserID : ":fmt.Sprint(response.UserId),
			})
		}else{
			ctx.JSON(http.StatusInternalServerError,gin.H{"error":err.Error()})
		}

	})
	// method which is used to fetch the user details for the given userID.
	g.GET("/fetchuser/:userid",func(ctx *gin.Context){

		userID := ctx.Param("userid")
		glog.Info("Data",userID)

		req := &proto.Response{UserId : userID}

		if response ,err := client.FetchUser(ctx,req); err == nil {
			ctx.JSON(http.StatusOK,gin.H{
				"Fetched User Details UserName,Email":fmt.Sprint(response.Name,response.Email),
			})
		}else{
			ctx.JSON(http.StatusInternalServerError,gin.H{"error":err.Error()})
		}

	})
	if err := g.Run(":8080");err != nil {
		glog.Fatalf("failed to run server %v",err)
	}
}