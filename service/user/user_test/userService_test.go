package user_test

import (
	"context"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	userpb "pdmicro/proto/user"
	"pdmicro/service/user"
	"sync"
	"testing"
	"time"
)

func TestUserService_GetUserDetails(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		expectedErr error
		expectedRes *userpb.GetUserResponse
	}{
		{
			name:        "User exists",
			userID:      "user123",
			expectedErr: nil,
			expectedRes: &userpb.GetUserResponse{Name: "John Doe", Email: "john.doe@example.com"},
		},
		{
			name:        "User not found",
			userID:      "user999",
			expectedErr: status.Errorf(codes.NotFound, "user with ID user999 not found"),
			expectedRes: nil,
		},
		{
			name:        "Internal server error",
			userID:      "error",
			expectedErr: status.Errorf(codes.Internal, "internal server error occurred"),
			expectedRes: nil,
		},
	}

	// Initialize the UserServiceServer
	server := &user.UserServiceServer{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the request with the userID from the test case
			req := &userpb.GetUserRequest{UserId: tt.userID}
			// Call the GetUserDetails method
			resp, err := server.GetUserDetails(context.Background(), req)

			// Assertions
			if tt.expectedErr != nil {
				require.Error(t, err)
				require.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedRes, resp)
			}
		})
	}
}

func TestUserService_GetUserDetails_Performance(t *testing.T) {
	// Setup your server
	server := &user.UserServiceServer{}

	// Number of concurrent requests you want to simulate
	concurrentRequests := 1000
	var wg sync.WaitGroup
	startTime := time.Now()

	// Simulate multiple concurrent gRPC requests
	for i := 0; i < concurrentRequests; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			// Create a request with a userID, can vary the userID or use a fixed one
			req := &userpb.GetUserRequest{UserId: "user123"}

			// Call GetUserDetails for this request
			_, err := server.GetUserDetails(context.Background(), req)

			// Ensure no error occurred
			require.NoError(t, err)
		}(i)
	}

	// Wait for all requests to complete
	wg.Wait()

	// Measure the time taken for all requests
	duration := time.Since(startTime)
	t.Logf("Processed %d concurrent requests in %s", concurrentRequests, duration)

	// Optionally, you can also assert on response times, etc.
	require.Less(t, duration.Seconds(), 5.0, "Request handling time should be under 5 seconds")
}
