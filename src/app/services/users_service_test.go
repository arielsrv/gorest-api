package services_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/iskaypetcom/digital/oms/api-core/gorest-api/src/app/model"
	"gitlab.com/iskaypetcom/digital/oms/api-core/gorest-api/src/app/services"
	"gitlab.com/iskaypetcom/digital/oms/api-core/gorest-api/src/resources/mocks/src/app/clients"
)

func TestService_GetUsers(t *testing.T) {
	userClient := &clients.MockIUserClient{}

	userClient.On("GetUsers").Return([]model.UserResponse{{ID: 1}, {ID: 2}}, nil)

	userClient.On("GetUser", 1).Return(&model.UserResponse{ID: 1}, nil)
	userClient.On("GetPosts", 1).Return([]model.PostResponse{{ID: 1, UserID: 1, Title: "post1"}}, nil)
	userClient.On("GetComments", 1).Return([]model.CommentResponse{{ID: 1, PostID: 1, Name: "comment1"}}, nil)
	userClient.On("GetTodos", 1).Return([]model.TodoResponse{{ID: 1, UserID: 1, Title: "todo1"}}, nil)

	userClient.On("GetUser", 2).Return(&model.UserResponse{ID: 2}, nil)
	userClient.On("GetPosts", 2).Return([]model.PostResponse{{ID: 2, UserID: 2, Title: "post2"}}, nil)
	userClient.On("GetComments", 2).Return([]model.CommentResponse{{ID: 2, PostID: 2, Name: "comment2"}}, nil)
	userClient.On("GetTodos", 2).Return([]model.TodoResponse{{ID: 2, UserID: 2, Title: "todo2"}}, nil)

	userService := services.NewUserService(userClient)

	users, err := userService.GetUsers()

	require.NoError(t, err)
	assert.NotNil(t, users)
	assert.Len(t, users, 2)

	assert.Len(t, users[0].Posts, 1)
	assert.Len(t, users[0].Todos, 1)

	assert.Len(t, users[1].Posts, 1)
	assert.Len(t, users[1].Todos, 1)
}
