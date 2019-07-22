// A grpc server which implements the Create user and the fetch user functionality
// the inputs will be given by the grpc client
// Built using grpc and proto3

package main

import (
	"context"
	"github.com/api-server/core/mongo"
	proto "github.com/api-server/proto"
	"github.com/api-server/services"
	"github.com/api-server/types"
	"github.com/golang/glog"
	"github.com/spf13/cast"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
)

type server struct {}


func main(){
	// listening to the requests on port 4040
	listener,err := net.Listen("tcp",":4040")
	if err != nil {
		panic (err)
	}
	//creating a new grpc server
	srv := grpc.NewServer()
	mongo.InitMongo()
	glog.Info("DEBUG: SERVER STARTED")
	proto.RegisterAddServiceServer(srv,&server{})
	reflection.Register(srv)
	if e := srv.Serve(listener);e!=nil{
		panic(e)
	}
}

// CreateUser function implements the creation of the user functionality for the given user name
// and the email and returns the UserID as the response.
// the user data will be stored in the given mongoDB.
func(s *server)CreateUser(ctx context.Context,request *proto.Request)(*proto.Response,error){

	// get the Name the Email passed by the Client using the proto request
	name,email := request.GetName(),request.GetEmail()
	var user types.User
	user.FirstName = name
	user.Email = email
	//creating the user //using the existing mongo functionality
	u, err := services.CreateUser(user)
	if err != nil {
		panic(err)
		return nil,err
	}
	glog.Info("USER SUCESSFULLY CREATED :",u)
	return &proto.Response{UserId:cast.ToString(u.Id)},nil
}

// FetchUser function accepts the userID given by the client and gives the user data such as the user name and the user email.
// this function also fetches the data from the given mongo connection.
func (s*server) FetchUser (ctx context.Context,request *proto.Response)(*proto.Request,error){

	// get the userID passed by the client using the proto request.
	b := request.GetUserId()
	oid, err := primitive.ObjectIDFromHex(b)
	if err != nil {
		glog.Errorln("convert hex id error")
		panic(err)
		return nil,err
	}
	// fetching the userInformation using the existing mongo functinality
	userInfo, err := services.GetUserById(oid)
	if err != nil {
		glog.Errorln("error while fetching the user details.")
		panic(err)
	}
	glog.Info(userInfo,userInfo.FirstName,userInfo.Email)
	return &proto.Request{Name :userInfo.FirstName,Email:userInfo.Email},nil
}
