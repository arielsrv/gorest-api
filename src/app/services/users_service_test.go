package services_test

import (
	"errors"
	"testing"

	"gitlab.com/iskaypetcom/digital/oms/api-core/gorest-api/src/app/model/paging"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/iskaypetcom/digital/oms/api-core/gorest-api/src/app/model"
	"gitlab.com/iskaypetcom/digital/oms/api-core/gorest-api/src/app/services"
	"gitlab.com/iskaypetcom/digital/oms/api-core/gorest-api/src/resources/mocks/src/app/clients"
)

func TestService_GetUsers(t *testing.T) {
	userClient := clients.NewMockIUserClient(t)

	userClient.EXPECT().GetUsers(1, 10).Return(&paging.PagedResultResponse[model.UserResponse]{
		Results: []model.UserResponse{{ID: 1}, {ID: 2}},
	}, nil)

	userClient.EXPECT().GetUser(1).Return(&model.UserResponse{ID: 1}, nil)
	userClient.EXPECT().GetPosts(1).Return([]model.PostResponse{{ID: 1, UserID: 1, Title: "post1"}}, nil)
	userClient.EXPECT().GetComments(1).Return([]model.CommentResponse{{ID: 1, PostID: 1, Name: "comment1"}, {ID: 2, PostID: 1, Name: "comment2"}}, nil)
	userClient.EXPECT().GetTodos(1).Return([]model.TodoResponse{{ID: 1, UserID: 1, Title: "todo1"}}, nil)

	userClient.EXPECT().GetUser(2).Return(&model.UserResponse{ID: 2}, nil)
	userClient.EXPECT().GetPosts(2).Return([]model.PostResponse{{ID: 2, UserID: 2, Title: "post2"}}, nil)
	userClient.EXPECT().GetComments(2).Return([]model.CommentResponse{{ID: 3, PostID: 2, Name: "comment2"}}, nil)
	userClient.EXPECT().GetTodos(2).Return([]model.TodoResponse{{ID: 2, UserID: 2, Title: "todo2"}}, nil)

	userService := services.NewUserService(userClient)

	pagedResult, err := userService.GetUsers(1, 10)

	require.NoError(t, err)
	assert.NotNil(t, pagedResult)
	assert.Len(t, pagedResult.Results, 2)
	assert.Equal(t, 1, pagedResult.Results[0].ID)
	assert.Equal(t, 2, pagedResult.Results[1].ID)

	assert.Len(t, pagedResult.Results[0].Posts, 1)
	assert.Equal(t, 1, pagedResult.Results[0].Posts[0].ID)
	assert.Len(t, pagedResult.Results[0].Posts[0].Comments, 2)
	assert.Equal(t, 1, pagedResult.Results[0].Posts[0].Comments[0].ID)
	assert.Equal(t, 2, pagedResult.Results[0].Posts[0].Comments[1].ID)
	assert.Len(t, pagedResult.Results[0].Todos, 1)

	assert.Len(t, pagedResult.Results[1].Posts, 1)
	assert.Len(t, pagedResult.Results[1].Posts[0].Comments, 1)
	assert.Len(t, pagedResult.Results[1].Todos, 1)
}

func TestService_GetUsers_Err(t *testing.T) {
	userClient := clients.NewMockIUserClient(t)

	userClient.EXPECT().GetUsers(1, 10).Return(nil, errors.New("some error"))

	actual, err := services.NewUserService(userClient).GetUsers(1, 10)

	require.Error(t, err)
	assert.Nil(t, actual)
}

func TestService_GetUsers_Err_UserPool(t *testing.T) {
	userClient := clients.NewMockIUserClient(t)

	userClient.EXPECT().GetUsers(1, 10).Return(&paging.PagedResultResponse[model.UserResponse]{
		Results: []model.UserResponse{{ID: 1}, {ID: 2}},
	}, nil)

	userClient.EXPECT().GetUser(1).Return(&model.UserResponse{ID: 1}, nil)
	userClient.EXPECT().GetPosts(1).Return([]model.PostResponse{{ID: 1, UserID: 1, Title: "post1"}}, nil)
	userClient.EXPECT().GetComments(1).Return([]model.CommentResponse{{ID: 1, PostID: 1, Name: "comment1"}, {ID: 2, PostID: 1, Name: "comment2"}}, nil)
	userClient.EXPECT().GetTodos(1).Return([]model.TodoResponse{{ID: 1, UserID: 1, Title: "todo1"}}, nil)

	userClient.EXPECT().GetUser(2).Return(nil, errors.New("some error"))
	userClient.EXPECT().GetPosts(2).Return([]model.PostResponse{{ID: 2, UserID: 2, Title: "post2"}}, nil)
	userClient.EXPECT().GetComments(2).Return([]model.CommentResponse{{ID: 3, PostID: 2, Name: "comment2"}}, nil)
	userClient.EXPECT().GetTodos(2).Return([]model.TodoResponse{{ID: 2, UserID: 2, Title: "todo2"}}, nil)

	actual, err := services.NewUserService(userClient).GetUsers(1, 10)

	require.Error(t, err)
	assert.Nil(t, actual)
}

func TestService_GetUsers_Todo_Err(t *testing.T) {
	userClient := clients.NewMockIUserClient(t)

	userClient.EXPECT().GetUsers(1, 10).Return(&paging.PagedResultResponse[model.UserResponse]{
		Results: []model.UserResponse{{ID: 1}, {ID: 2}},
	}, nil)

	userClient.EXPECT().GetUser(1).Return(&model.UserResponse{ID: 1}, nil)
	userClient.EXPECT().GetPosts(1).Return([]model.PostResponse{{ID: 1, UserID: 1, Title: "post1"}}, nil)
	userClient.EXPECT().GetComments(1).Return([]model.CommentResponse{{ID: 1, PostID: 1, Name: "comment1"}, {ID: 2, PostID: 1, Name: "comment2"}}, nil)
	userClient.EXPECT().GetTodos(1).Return(nil, errors.New("some error"))

	userClient.EXPECT().GetUser(2).Return(&model.UserResponse{ID: 2}, nil)
	userClient.EXPECT().GetPosts(2).Return([]model.PostResponse{{ID: 2, UserID: 2, Title: "post2"}}, nil)
	userClient.EXPECT().GetComments(2).Return([]model.CommentResponse{{ID: 3, PostID: 2, Name: "comment2"}}, nil)
	userClient.EXPECT().GetTodos(2).Return([]model.TodoResponse{{ID: 2, UserID: 2, Title: "todo2"}}, nil)

	actual, err := services.NewUserService(userClient).GetUsers(1, 10)

	require.Error(t, err)
	assert.Nil(t, actual)
}

func TestService_GetUsers_Post_Err(t *testing.T) {
	userClient := clients.NewMockIUserClient(t)

	userClient.EXPECT().GetUsers(1, 10).Return(&paging.PagedResultResponse[model.UserResponse]{
		Results: []model.UserResponse{{ID: 1}, {ID: 2}},
	}, nil)

	userClient.EXPECT().GetUser(1).Return(&model.UserResponse{ID: 1}, nil)
	userClient.EXPECT().GetPosts(1).Return(nil, errors.New("some error"))
	userClient.EXPECT().GetTodos(1).Return([]model.TodoResponse{{ID: 1, UserID: 1, Title: "todo1"}}, nil)

	userClient.EXPECT().GetUser(2).Return(&model.UserResponse{ID: 2}, nil)
	userClient.EXPECT().GetPosts(2).Return([]model.PostResponse{{ID: 2, UserID: 2, Title: "post2"}}, nil)
	userClient.EXPECT().GetComments(2).Return([]model.CommentResponse{{ID: 3, PostID: 2, Name: "comment2"}}, nil)
	userClient.EXPECT().GetTodos(2).Return([]model.TodoResponse{{ID: 2, UserID: 2, Title: "todo2"}}, nil)

	actual, err := services.NewUserService(userClient).GetUsers(1, 10)

	require.Error(t, err)
	assert.Nil(t, actual)
}

func TestService_GetUsers_Comments_Err(t *testing.T) {
	userClient := clients.NewMockIUserClient(t)

	userClient.EXPECT().GetUsers(1, 10).Return(&paging.PagedResultResponse[model.UserResponse]{
		Results: []model.UserResponse{{ID: 1}, {ID: 2}},
	}, nil)

	userClient.EXPECT().GetUser(1).Return(&model.UserResponse{ID: 1}, nil)
	userClient.EXPECT().GetPosts(1).Return([]model.PostResponse{{ID: 1, UserID: 1, Title: "post1"}}, nil)
	userClient.EXPECT().GetComments(1).Return(nil, errors.New("some error"))
	userClient.EXPECT().GetTodos(1).Return([]model.TodoResponse{{ID: 1, UserID: 1, Title: "todo1"}}, nil)

	userClient.EXPECT().GetUser(2).Return(&model.UserResponse{ID: 2}, nil)
	userClient.EXPECT().GetPosts(2).Return([]model.PostResponse{{ID: 2, UserID: 2, Title: "post2"}}, nil)
	userClient.EXPECT().GetComments(2).Return([]model.CommentResponse{{ID: 3, PostID: 2, Name: "comment2"}}, nil)
	userClient.EXPECT().GetTodos(2).Return([]model.TodoResponse{{ID: 2, UserID: 2, Title: "todo2"}}, nil)

	actual, err := services.NewUserService(userClient).GetUsers(1, 10)

	require.Error(t, err)
	assert.Nil(t, actual)
}
