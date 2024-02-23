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
	userClient := clients.NewMockIUserClient(t)
	userClient.EXPECT().GetUsers().Return([]model.UserResponse{{ID: 1}}, nil)

	userClient.EXPECT().GetUser(1).Return(&model.UserResponse{ID: 1}, nil)
	userClient.EXPECT().GetPosts(1).Return([]model.PostResponse{{ID: 1, UserID: 1, Title: "Hello world!", Body: "From Golang"}}, nil)
	userClient.EXPECT().GetTodos(1).Return([]model.TodoResponse{{ID: 1, UserID: 1, Title: "Hello world!"}}, nil)
	userClient.EXPECT().GetComments(1).Return([]model.CommentResponse{{ID: 1, Name: "Hello world!"}}, nil)

	userService := services.NewUserService(userClient)

	actual, err := userService.GetUsers()
	require.NoError(t, err)
	assert.NotNil(t, actual)
	assert.Len(t, actual, 1)
}
