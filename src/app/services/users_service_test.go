package services_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/iskaypetcom/digital/oms/api-core/gorest-api/src/app/model"
	"gitlab.com/iskaypetcom/digital/oms/api-core/gorest-api/src/app/services"
	"gitlab.com/iskaypetcom/digital/oms/api-core/gorest-api/src/resources/mocks/src/app/clients"
)

func TestService_GetUsers(t *testing.T) {
	userClient := clients.NewMockIUserClient(t)
	userClient.EXPECT().GetUsers().Return([]model.UserResponse{{ID: 1}, {ID: 2}}, nil)

	userClient.EXPECT().GetUser(1).Return(&model.UserResponse{ID: 1}, nil)
	userClient.EXPECT().GetPosts(1).Return([]model.PostResponse{{ID: 123, UserID: 1, Title: "Hello world!", Body: "From Golang"}}, nil)
	userClient.EXPECT().GetTodos(1).Return([]model.TodoResponse{{ID: 123, UserID: 1, Title: "Hello world!"}}, nil)
	userClient.EXPECT().GetComments(123).Return([]model.CommentResponse{{ID: 123, Name: "Hello world!"}}, nil)

	userClient.EXPECT().GetUser(2).Return(&model.UserResponse{ID: 2}, nil)
	userClient.EXPECT().GetPosts(2).Return([]model.PostResponse{{ID: 124, UserID: 1, Title: "Hello world!", Body: "From Golang"}}, nil)
	userClient.EXPECT().GetTodos(2).Return([]model.TodoResponse{{ID: 123, UserID: 2, Title: "Hello world!"}}, nil)
	userClient.EXPECT().GetComments(124).Return([]model.CommentResponse{{ID: 124, Name: "Hello world!"}}, nil)

	userService := services.NewUserService(userClient)

	actual, err := userService.GetUsers()
	require.NoError(t, err)
	assert.NotNil(t, actual)
	assert.Len(t, actual, 2)
}

func TestService_GetUsers_Err(t *testing.T) {
	userClient := clients.NewMockIUserClient(t)
	userClient.EXPECT().GetUsers().Return(nil, errors.New("error on blocking GetUsers"))

	userService := services.NewUserService(userClient)

	actual, err := userService.GetUsers()
	require.Error(t, err)
	assert.Nil(t, actual)
}

func TestService_GetUsers_SomeErr(t *testing.T) {
	userClient := clients.NewMockIUserClient(t)
	userClient.EXPECT().GetUsers().Return([]model.UserResponse{{ID: 1}, {ID: 2}}, nil)
	userClient.EXPECT().GetUser(1).Return(&model.UserResponse{ID: 1}, nil)
	userClient.EXPECT().GetPosts(1).Return([]model.PostResponse{}, nil)
	userClient.EXPECT().GetTodos(1).Return([]model.TodoResponse{}, nil)
	userClient.EXPECT().GetUser(2).Return(nil, errors.New("some error"))
	userClient.EXPECT().GetPosts(2).Return([]model.PostResponse{}, nil)
	userClient.EXPECT().GetTodos(2).Return([]model.TodoResponse{}, nil)

	userService := services.NewUserService(userClient)

	actual, err := userService.GetUsers()
	require.Error(t, err)
	assert.Nil(t, actual)
}

func TestService_GetUsers_Post_SomeErr(t *testing.T) {
	userClient := clients.NewMockIUserClient(t)
	userClient.EXPECT().GetUsers().Return([]model.UserResponse{{ID: 1}, {ID: 2}}, nil)
	userClient.EXPECT().GetUser(1).Return(&model.UserResponse{ID: 1}, nil)
	userClient.EXPECT().GetPosts(1).Return(nil, errors.New("some error"))
	userClient.EXPECT().GetTodos(1).Return([]model.TodoResponse{}, nil)
	userClient.EXPECT().GetUser(2).Return(&model.UserResponse{ID: 2}, nil)
	userClient.EXPECT().GetPosts(2).Return(nil, errors.New("some error"))
	userClient.EXPECT().GetTodos(2).Return([]model.TodoResponse{}, nil)

	userService := services.NewUserService(userClient)

	actual, err := userService.GetUsers()
	require.Error(t, err)
	assert.Nil(t, actual)
}

func TestService_GetUsers_Todo_SomeErr(t *testing.T) {
	userClient := clients.NewMockIUserClient(t)
	userClient.EXPECT().GetUsers().Return([]model.UserResponse{{ID: 1}}, nil)
	userClient.EXPECT().GetUser(1).Return(&model.UserResponse{ID: 1}, nil)
	userClient.EXPECT().GetPosts(1).Return([]model.PostResponse{}, nil)
	userClient.EXPECT().GetTodos(1).Return(nil, errors.New("some error"))

	userService := services.NewUserService(userClient)

	actual, err := userService.GetUsers()
	require.Error(t, err)
	assert.Nil(t, actual)
}
