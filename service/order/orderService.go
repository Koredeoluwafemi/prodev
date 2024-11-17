package order

import (
	"context"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	pb "pdmicro/proto/order"
	userpb "pdmicro/proto/user"
)

type OrderServiceServer struct {
	pb.UnimplementedOrderServiceServer
	UserClient userpb.UserServiceClient
}

func (s *OrderServiceServer) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	// Channels for results and errors
	userRespChan := make(chan *userpb.GetUserResponse, 1)
	userErrChan := make(chan error, 1)

	// Fetch user details concurrently
	go func() {
		userResp, err := s.UserClient.GetUserDetails(ctx, &userpb.GetUserRequest{UserId: req.UserId})
		if err != nil {
			userErrChan <- err
			return
		}
		userRespChan <- userResp
	}()

	// Handle results from goroutine
	select {
	case userErr := <-userErrChan:
		st, ok := status.FromError(userErr)
		if ok {
			log.Printf("User service error: %v", st.Message())
			//return nil, userErr // Preserve the original error
			return nil, status.Errorf(codes.NotFound, "failed to fetch user details: %v", st.Message())
		}
		return nil, fmt.Errorf("unknown error: %v", userErr)

	case userResp := <-userRespChan:
		if userResp == nil {
			return nil, status.Error(codes.Internal, "failed to retrieve user details")
		}

		// Successful user response
		return &pb.CreateOrderResponse{
			Status:    "Order Created Successfully",
			UserName:  userResp.Name,
			UserEmail: userResp.Email,
		}, nil
	}
}
