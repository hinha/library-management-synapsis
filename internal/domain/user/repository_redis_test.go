package user

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRedisClient is a mock implementation of RedisClientInterface
type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	args := m.Called(ctx, key, value, expiration)
	return args.Get(0).(*redis.StatusCmd)
}

func (m *MockRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.StringCmd)
}

// Ping mocks the Ping method of the Redis client
func (m *MockRedisClient) Ping(ctx context.Context) *redis.StatusCmd {
	args := m.Called(ctx)
	return args.Get(0).(*redis.StatusCmd)
}

func TestNewCacheRepository(t *testing.T) {
	// Create a real Redis client for this test
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // This won't be used as we're not making actual calls
	})

	// Call the constructor
	repo := NewCacheRepository(client)

	// Assert that the returned repository is not nil and implements ICacheRepository
	assert.NotNil(t, repo)
	_, ok := repo.(ICacheRepository)
	assert.True(t, ok, "Repository should implement ICacheRepository")
}

func TestCacheRepository_Ping(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name          string
		setup         func(mockClient *MockRedisClient)
		expectedError error
	}{
		{
			name: "Success",
			setup: func(mockClient *MockRedisClient) {
				cmd := redis.NewStatusCmd(context.Background())
				cmd.SetVal("PONG")
				mockClient.On("Ping", mock.Anything).Return(cmd)
			},
			expectedError: nil,
		},
		{
			name: "Redis Error",
			setup: func(mockClient *MockRedisClient) {
				cmd := redis.NewStatusCmd(context.Background())
				cmd.SetErr(errors.New("connection refused"))
				mockClient.On("Ping", mock.Anything).Return(cmd)
			},
			expectedError: errors.New("connection refused"),
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mock client
			mockClient := new(MockRedisClient)
			tc.setup(mockClient)

			// Create repository with mock client
			repo := &CacheRepository{
				client: mockClient,
			}

			// Execute the method being tested
			ctx := context.Background()
			err := repo.Ping(ctx)

			// Assert the results
			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			// Verify that all expected calls were made
			mockClient.AssertExpectations(t)
		})
	}
}

func TestCacheRepository_SaveUser(t *testing.T) {
	testCases := []struct {
		name          string
		setup         func(mockClient *MockRedisClient)
		user          *User
		expectedError error
	}{
		{
			name: "Success",
			setup: func(mockClient *MockRedisClient) {
				cmd := redis.NewStatusCmd(context.Background())
				cmd.SetVal("OK")
				mockClient.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(cmd)
			},
			user:          &User{ID: 1, Name: "testuser"},
			expectedError: nil,
		},
		{
			name: "Redis Error",
			setup: func(mockClient *MockRedisClient) {
				cmd := redis.NewStatusCmd(context.Background())
				cmd.SetErr(errors.New("redis set error"))
				mockClient.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(cmd)
			},
			user:          &User{ID: 2, Name: "failuser"},
			expectedError: errors.New("redis set error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := new(MockRedisClient)
			tc.setup(mockClient)

			repo := &CacheRepository{
				client: mockClient,
			}

			ctx := context.Background()
			err := repo.SaveUser(ctx, tc.user)

			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestCacheRepository_GetUser(t *testing.T) {
	testCases := []struct {
		name          string
		setup         func(mockClient *MockRedisClient)
		id            uint
		expectedUser  *User
		expectedError error
	}{
		{
			name: "Success",
			setup: func(mockClient *MockRedisClient) {
				user := &User{ID: 1, Name: "testuser"}
				data, _ := json.Marshal(user)
				cmd := redis.NewStringCmd(context.Background())
				cmd.SetVal(string(data))
				mockClient.On("Get", mock.Anything, mock.Anything).Return(cmd)
			},
			id:            1,
			expectedUser:  &User{ID: 1, Name: "testuser"},
			expectedError: nil,
		},
		{
			name: "User Not Found",
			setup: func(mockClient *MockRedisClient) {
				cmd := redis.NewStringCmd(context.Background())
				cmd.SetErr(redis.Nil)
				mockClient.On("Get", mock.Anything, mock.Anything).Return(cmd)
			},
			id:            2,
			expectedUser:  nil,
			expectedError: ErrUserNotFound,
		},
		{
			name: "Redis Error",
			setup: func(mockClient *MockRedisClient) {
				cmd := redis.NewStringCmd(context.Background())
				cmd.SetErr(errors.New("redis get error"))
				mockClient.On("Get", mock.Anything, mock.Anything).Return(cmd)
			},
			id:            3,
			expectedUser:  nil,
			expectedError: errors.New("redis get error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := new(MockRedisClient)
			tc.setup(mockClient)

			repo := &CacheRepository{
				client: mockClient,
			}

			ctx := context.Background()
			user, err := repo.GetUser(ctx, tc.id)

			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError.Error(), err.Error())
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedUser, user)
			}

			mockClient.AssertExpectations(t)
		})
	}
}
