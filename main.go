package main

import (
	"log"
	"net"
	"pdmicro/service/order"
	"pdmicro/service/user"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	//"path/to/grpc-microservices/internal/user_service"
	orderpb "pdmicro/proto/order"
	userpb "pdmicro/proto/user"
)

func main() {
	// Start User Service
	go func() {
		listener, err := net.Listen("tcp", ":50051")
		if err != nil {
			log.Fatalf("Failed to listen: %v", err)
		}

		grpcServer := grpc.NewServer()
		userpb.RegisterUserServiceServer(grpcServer, &user.UserServiceServer{})
		log.Println("User Service running on :50051")
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Start Order Service
	go func() {
		listener, err := net.Listen("tcp", ":50052")
		if err != nil {
			log.Fatalf("Failed to listen: %v", err)
		}

		// Connect to User Service
		conn, err := grpc.Dial(":50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("Failed to connect to User Service: %v", err)
		}
		defer conn.Close()

		userClient := userpb.NewUserServiceClient(conn)
		grpcServer := grpc.NewServer()
		orderpb.RegisterOrderServiceServer(grpcServer, &order.OrderServiceServer{UserClient: userClient})
		log.Println("Order Service running on :50052")
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Block the main function
	select {}
}
