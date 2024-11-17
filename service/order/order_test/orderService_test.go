package order_test

import (
	"context"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	orderspb "pdmicro/proto/order"
	userpb "pdmicro/proto/user"
	"pdmicro/service/order"
	"sync"
	"testing"
	"time"
)

// Mock for the UserServiceClient
type MockUserServiceClient struct {
	mock.Mock
}

func (m *MockUserServiceClient) GetUserDetails(ctx context.Context, in *userpb.GetUserRequest, opts ...grpc.CallOption) (*userpb.GetUserResponse, error) {
	args := m.Called(ctx, in) // Only match the first two arguments (context and request)
	if resp, ok := args.Get(0).(*userpb.GetUserResponse); ok {
		return resp, args.Error(1)
	}
	return nil, args.Error(1)
}

func TestOrderService_CreateOrder(t *testing.T) {
	mockUserClient := new(MockUserServiceClient)

	// Mock success response
	mockUserClient.On("GetUserDetails", mock.Anything, &userpb.GetUserRequest{UserId: "user123"}).
		Return(&userpb.GetUserResponse{Name: "John Doe", Email: "john.doe@example.com"}, nil)

	// Mock error response
	mockUserClient.On("GetUserDetails", mock.Anything, &userpb.GetUserRequest{UserId: "invalid_user"}).
		Return(nil, status.Errorf(codes.NotFound, "user not found"))

	server := &order.OrderServiceServer{UserClient: mockUserClient}

	// Test valid order creation
	req := &orderspb.CreateOrderRequest{UserId: "user123"}
	resp, err := server.CreateOrder(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, "Order Created Successfully", resp.Status)

	// Test error scenario
	req = &orderspb.CreateOrderRequest{UserId: "invalid_user"}
	_, err = server.CreateOrder(context.Background(), req)
	require.Error(t, err)

	st, _ := status.FromError(err)
	require.Equal(t, codes.NotFound, st.Code()) // Expecting NotFound instead of Internal
	require.Equal(t, "failed to fetch user details: user not found", st.Message())
}

func TestOrderService_CreateOrder_Concurrency(t *testing.T) {
	mockUserClient := new(MockUserServiceClient)

	// Mock responses
	mockUserClient.On("GetUserDetails", mock.Anything, &userpb.GetUserRequest{UserId: "user123"}).
		Return(&userpb.GetUserResponse{Name: "John Doe", Email: "john.doe@example.com"}, nil)

	server := &order.OrderServiceServer{UserClient: mockUserClient}

	var wg sync.WaitGroup
	wg.Add(2)

	// Simulate two concurrent order creation requests
	go func() {
		defer wg.Done()
		req := &orderspb.CreateOrderRequest{UserId: "user123"}
		resp, err := server.CreateOrder(context.Background(), req)
		require.NoError(t, err)
		require.Equal(t, "Order Created Successfully", resp.Status)
	}()

	go func() {
		defer wg.Done()
		req := &orderspb.CreateOrderRequest{UserId: "user123"}
		resp, err := server.CreateOrder(context.Background(), req)
		require.NoError(t, err)
		require.Equal(t, "Order Created Successfully", resp.Status)
	}()

	wg.Wait()
}

func TestOrderService_CreateOrder_Performance(t *testing.T) {
	// Mock the UserServiceClient
	mockUserClient := new(MockUserServiceClient)

	// Mock user details response for the valid user ID
	mockUserClient.On("GetUserDetails", mock.Anything, &userpb.GetUserRequest{UserId: "user123"}).
		Return(&userpb.GetUserResponse{Name: "John Doe", Email: "john.doe@example.com"}, nil)

	// Mock user details response for the invalid user ID
	mockUserClient.On("GetUserDetails", mock.Anything, &userpb.GetUserRequest{UserId: "invalid_user"}).
		Return(nil, status.Errorf(codes.NotFound, "user not found"))

	// Create the OrderService server
	server := &order.OrderServiceServer{UserClient: mockUserClient}

	// Number of concurrent requests to simulate
	concurrentRequests := 1000

	// WaitGroup to wait for all goroutines to finish
	var wg sync.WaitGroup
	wg.Add(concurrentRequests)

	// Start measuring the performance
	startTime := time.Now()

	// Simulate concurrent CreateOrder requests
	for i := 0; i < concurrentRequests; i++ {
		go func() {
			defer wg.Done()

			// Simulate order creation with a valid user ID
			req := &orderspb.CreateOrderRequest{UserId: "user123"}
			_, err := server.CreateOrder(context.Background(), req)

			// Check for errors (for simplicity, we assume no error in valid cases)
			require.NoError(t, err)
		}()
	}

	// Wait for all requests to complete
	wg.Wait()

	// Measure the time taken for all concurrent requests
	duration := time.Since(startTime)

	// Log the performance result
	t.Logf("Processed %d concurrent CreateOrder requests in %s", concurrentRequests, duration)

	// Optionally, set a performance threshold, e.g., all requests should complete in under 5 seconds
	require.Less(t, duration.Seconds(), 5.0, "Request handling time should be under 5 seconds for %d concurrent requests", concurrentRequests)
}
