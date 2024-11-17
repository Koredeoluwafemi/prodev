package user

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	userpb "pdmicro/proto/user"
)

type UserServiceServer struct {
	userpb.UnimplementedUserServiceServer
	UserClient userpb.UserServiceClient
}

func (s *UserServiceServer) GetUserDetails(ctx context.Context, req *userpb.GetUserRequest) (*userpb.GetUserResponse, error) {
	// Simulate database lookup
	userData := map[string]*userpb.GetUserResponse{
		"user123": {Name: "John Doe", Email: "john.doe@example.com"},
	}

	// Simulate other potential errors (e.g., database failure)
	if req.UserId == "error" {
		return nil, status.Errorf(codes.Internal, "internal server error occurred")
	}

	user, exists := userData[req.UserId]
	if !exists {
		// Return a NOT_FOUND error if the user does not exist
		return nil, status.Errorf(codes.NotFound, "user with ID %s not found", req.UserId)
	}

	return user, nil
}
