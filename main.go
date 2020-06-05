package main 

import(
	"context"
	"log"
	"net"
	"os"
	"sync"
	"google.golang.org/grpc"
	proto "github.com/TwinkleMehta/chatserver/proto"
	glog "google.golang.org/grpc/grpclog"
)

var grpcLog glog.LoggerV2

func init() {
	grpcLog = glog.NewLoggerV2(os.Stdout, os.Stdout, os.Stdout)
}

type Connection struct {
	stream proto.Broadcast_CreateStreamServer 
	id string
	active bool 
	error chan error // goroutine requires this to be a channel
}

type Server struct {
	Connection []*Connection 
}

// CreateStream() adds each client connection to server and creates stream object
func (s *Server) CreateStream(pconn *proto.Connect, stream proto.Broadcast_CreateStreamServer) error {
	conn := &Connection{
		stream: stream, 
		id: pconn.User.Id, 
		active: true, 
		error: make(chan error),
	}

	s.Connection = append(s.Connection, conn)
	return <- conn.error
}

// BroadcastMessage() sends message to each client connection from
func (s *Server) BroadcastMessage(ctx context.Context, ms *proto.Message) (*proto.Close, error){
	 wait := sync.WaitGroup{}
	 done := make(chan int)
	 
	 for _, conn := range s.Connection {
		 wait.Add(1)

		 go func(msg *proto.Message, conn *Connection){
			defer wait.Done()
			if conn.active{
				err := conn.stream.Send(msg)
				grpcLog.Info("Sending message to: ", conn.stream)

				if err != nil {
					grpcLog.Errorf("Error with Stream: %s - Error: %v", conn.stream, err)
					conn.active = false 
					conn.error <- err
				}
			}
		 }(ms,conn)
	 }

	 go func() { // waits for all other goroutines to finish
		 wait.Wait()
		 close(done)
	 }()

	 <- done 
	 return &proto.Close{}, nil
 }

func main() {
	var connections []*Connection 
	server := &Server{connections}

	grpcServer := grpc.NewServer()
	listener, err := net.Listen("tcp",":8080")
	if err != nil{
		log.Fatalf("error creating the server %v", err)
	}

	grpcLog.Info("Starting server at port: 8080")

	proto.RegisterBroadcastServer(grpcServer, server)
	grpcServer.Serve(listener)
    
}
