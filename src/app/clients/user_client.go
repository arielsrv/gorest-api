package clients

import (
	"fmt"
	"net/http"
	"strconv"

	"gitlab.com/iskaypetcom/digital/oms/api-core/gorest-api/src/app/model/paging"

	"gitlab.com/iskaypetcom/digital/oms/api-core/gorest-api/src/app/model"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/rest"
)

type IUserClient interface {
	GetUsers(page int, perPage int) (*paging.PagedResultResponse[model.UserResponse], error)
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

func (c *UserClient) GetUsers(page int, perPage int) (*paging.PagedResultResponse[model.UserResponse], error) {
	apiURL := "/users"

	if page > 0 {
		apiURL += "?page=" + strconv.Itoa(page)
	}

	if perPage > 0 {
		apiURL += "&per_page=" + strconv.Itoa(perPage)
	}

	response := c.rb.Get(apiURL)
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

	limit, err := strconv.Atoi(response.Header.Get("X-Pagination-Limit"))
	if err != nil {
		return nil, err
	}

	pageNumber, err := strconv.Atoi(response.Header.Get("X-Pagination-Page"))
	if err != nil {
		return nil, err
	}

	pages, err := strconv.Atoi(response.Header.Get("X-Pagination-Pages"))
	if err != nil {
		return nil, err
	}

	total, err := strconv.Atoi(response.Header.Get("X-Pagination-Total"))
	if err != nil {
		return nil, err
	}

	pagedResult := &paging.PagedResultResponse[model.UserResponse]{
		Limit:   limit,
		Page:    pageNumber,
		Pages:   pages,
		Total:   total,
		Results: userResponses,
	}

	return pagedResult, nil
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
