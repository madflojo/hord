package app

import (
	"context"
	"hord/config"
	"net"
	"time"

	"github.com/madflojo/hord/databases"
	pb "github.com/madflojo/hord/proto/client"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// server is used to implement the client protobuf server interface
type Server struct{}

// Listen will start the grpc server listening on the defined port
func Listen() error {
	lis, err := net.Listen("tcp", Config.Listen+":"+Config.GRPCPort)
	if err != nil {
		return err
	}
	srv := grpc.NewServer()
	pb.RegisterHordServer(srv, &Server{})
	err = srv.Serve(lis)
	if err != nil {
		return err
	}
	return nil
}

// Get will retrieve requested information from the datastore and return it
func (s *Server) Get(ctx context.Context, msg *pb.GetRequest) (*pb.GetResponse, error) {
	// Define reply message
	r := &pb.GetResponse{
		Status: &pb.Status{
			Code:        0,
			Description: "Success",
		},
	}

	// Check key length
	if len(msg.Key) == 0 {
		log.Trace(config.GrpcKeyNotDefine)
		r.Status.Code = 4
		r.Status.Description = config.GrpcKeyNotDefine
		return r, nil
	}

	// Fetch data using key
	d, err := db.Get(msg.Key)
	if err != nil {
		log.WithFields(logrus.Fields{"key": msg.Key, "error": err}).Tracef(config.GrpcFailedToFetchData+" for key - %s", err)
		r.Status.Code = 5
		r.Status.Description = config.GrpcFailedToFetchData
		return r, nil
	}

	// Return data to client
	r.Key = msg.Key
	r.Data = d.Data
	r.LastUpdated = d.LastUpdated
	return r, nil
}

// Set will take the supplied data and store it within the datastore returning success or failure
func (s *Server) Set(ctx context.Context, msg *pb.SetRequest) (*pb.SetResponse, error) {
	// Define reply message
	r := &pb.SetResponse{
		Status: &pb.Status{
			Code:        0,
			Description: "Success",
		},
	}

	// Check key length
	if len(msg.Key) == 0 {
		log.Trace(config.GrpcKeyNotDefine)
		r.Status.Code = 4
		r.Status.Description = config.GrpcKeyNotDefine
		return r, nil
	}

	// Create data item for insertion
	d := &databases.Data{}
	d.Data = msg.Data
	d.LastUpdated = time.Now().UnixNano()

	// Insert data into datastore
	err := db.Set(msg.Key, d)
	if err != nil {
		log.WithFields(logrus.Fields{"key": msg.Key, "error": err}).Tracef(config.GrpcFailedToStoreData+" for key - %s", err)
		r.Status.Code = 5
		r.Status.Description = config.GrpcFailedToStoreData
		return r, nil
	}

	r.Key = msg.Key
	return r, nil
}

// Delete will remove the specified key from the datastore and return success or failure
func (s *Server) Delete(ctx context.Context, msg *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	// Define reply message
	r := &pb.DeleteResponse{
		Status: &pb.Status{
			Code:        0,
			Description: "Success",
		},
	}

	// Check key length
	if len(msg.Key) == 0 {
		log.Trace(config.GrpcKeyNotDefine)
		r.Status.Code = 4
		r.Status.Description = config.GrpcKeyNotDefine
		return r, nil
	}

	// Delete data from datastore
	err := db.Delete(msg.Key)
	if err != nil {
		log.WithFields(logrus.Fields{"key": msg.Key, "error": err}).Tracef(config.GrpcFailedToDeleteData+" for key - %s", err)
		r.Status.Code = 5
		r.Status.Description = config.GrpcFailedToDeleteData
		return r, nil
	}

	r.Key = msg.Key
	return r, nil
}
