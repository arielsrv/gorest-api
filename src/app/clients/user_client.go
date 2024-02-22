package clients

import (
	"fmt"
	"net/http"

	"gitlab.com/iskaypetcom/digital/oms/api-core/gorest-api/src/app/model"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/rest"
)

type IUserClient interface {
	GetUsers() ([]model.UserResponse, error)
	GetUser(userID int) (*model.UserResponse, error)
	GetPosts(userID int) ([]model.PostResponse, error)
	GetTodos(userID int) ([]model.TodoResponse, error)
	GetComments(postID int) ([]model.CommentResponse, error)
}

type UserClient struct {
	rb rest.IRequestBuilder
}

func NewUserClient(rb rest.IRequestBuilder) *UserClient {
	return &UserClient{
		rb: rb,
	}
}

func (c *UserClient) GetUsers() ([]model.UserResponse, error) {
	response := c.rb.Get("/users")
	if response.Err != nil {
		return nil, response.Err
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	var userResponses []model.UserResponse
	if err := response.FillUp(&userResponses); err != nil {
		return nil, err
	}

	return userResponses, nil
}

func (c *UserClient) GetUser(userID int) (*model.UserResponse, error) {
	apiURL := fmt.Sprintf("/users/%d", userID)
	response := c.rb.Get(apiURL)

	if response.Err != nil {
		return nil, response.Err
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	userResponse := new(model.UserResponse)
	if err := response.FillUp(userResponse); err != nil {
		return nil, err
	}

	return userResponse, nil
}

func (c *UserClient) GetPosts(userID int) ([]model.PostResponse, error) {
	apiURL := fmt.Sprintf("/users/%d/posts", userID)
	response := c.rb.Get(apiURL)

	if response.Err != nil {
		return nil, response.Err
	}

	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusNotFound {
		return nil, fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	var postResponses []model.PostResponse
	if err := response.FillUp(&postResponses); err != nil {
		return nil, err
	}

	return postResponses, nil
}

func (c *UserClient) GetComments(postID int) ([]model.CommentResponse, error) {
	apiURL := fmt.Sprintf("/posts/%d/comments", postID)
	response := c.rb.Get(apiURL)

	if response.Err != nil {
		return nil, response.Err
	}

	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusNotFound {
		return nil, fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	var commentResponses []model.CommentResponse
	if err := response.FillUp(&commentResponses); err != nil {
		return nil, err
	}

	return commentResponses, nil
}

func (c *UserClient) GetTodos(userID int) ([]model.TodoResponse, error) {
	apiURL := fmt.Sprintf("/users/%d/todos", userID)
	response := c.rb.Get(apiURL)

	if response.Err != nil {
		return nil, response.Err
	}

	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusNotFound {
		return nil, fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	var todoResponses []model.TodoResponse
	if err := response.FillUp(&todoResponses); err != nil {
		return nil, err
	}

	return todoResponses, nil
}
