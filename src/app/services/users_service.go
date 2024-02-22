package services

import (
	"runtime"
	"slices"

	"github.com/alitto/pond"

	"gitlab.com/iskaypetcom/digital/oms/api-core/gorest-api/src/app/clients"
	"gitlab.com/iskaypetcom/digital/oms/api-core/gorest-api/src/app/model"
	"go.uber.org/multierr"
)

type IUsersService interface {
	GetUsers() ([]model.UserDTO, error)
}

type UsersService struct {
	userClient clients.IUserClient
}

func NewUserService(userClient clients.IUserClient) *UsersService {
	return &UsersService{
		userClient: userClient,
	}
}

type Task[T any] struct {
	Result T
	Err    error
}

func ToTask[T any](f func() (T, error)) Task[T] {
	result, err := f()

	return Task[T]{
		Result: result,
		Err:    err,
	}
}

func (r *UsersService) GetUsers() ([]model.UserDTO, error) {
	userResponses, err := r.userClient.GetUsers()
	if err != nil {
		return nil, err
	}

	uChan := make(chan Task[*model.UserResponse])
	pChan := make(chan Task[[]model.PostResponse])
	tChan := make(chan Task[[]model.TodoResponse])
	cChan := make(chan Task[[]model.CommentResponse], 100)

	var aggErr error
	var users []model.UserResponse
	posts := make(map[int][]model.PostResponse)
	todos := make(map[int][]model.TodoResponse)
	comments := make(map[int][]model.CommentResponse)

	pool := pond.New(runtime.NumCPU()-1, 100)

	for i := 0; i < len(userResponses); i++ {
		userID := userResponses[i].ID

		pool.Submit(func() {
			uChan <- ToTask[*model.UserResponse](func() (*model.UserResponse, error) {
				return r.userClient.GetUser(userID)
			})
		})

		pool.Submit(func() {
			pChan <- ToTask[[]model.PostResponse](func() ([]model.PostResponse, error) {
				response, pErr := r.userClient.GetPosts(userID)
				if pErr != nil {
					return nil, pErr
				}

				for k := 0; k < len(response); k++ {
					postID := response[k].ID
					pool.Submit(func() {
						cChan <- ToTask[[]model.CommentResponse](func() ([]model.CommentResponse, error) {
							return r.userClient.GetComments(postID)
						})
					})
				}

				return response, nil
			})
		})

		pool.Submit(func() {
			tChan <- ToTask[[]model.TodoResponse](func() ([]model.TodoResponse, error) {
				return r.userClient.GetTodos(userID)
			})
		})
	}

	for i := 0; i < len(userResponses)*3; i++ {
		select {
		case uTask := <-uChan:
			if uTask.Err != nil {
				aggErr = multierr.Append(aggErr, uTask.Err)
				continue
			}
			users = append(users, *uTask.Result)
		case pTask := <-pChan:
			if pTask.Err != nil {
				aggErr = multierr.Append(aggErr, pTask.Err)
				continue
			}
			for k := 0; k < len(pTask.Result); k++ {
				userID := pTask.Result[k].UserID
				posts[userID] = append(posts[userID], pTask.Result[k])
			}
		case tTask := <-tChan:
			if tTask.Err != nil {
				aggErr = multierr.Append(aggErr, tTask.Err)
				continue
			}
			for k := 0; k < len(tTask.Result); k++ {
				userID := tTask.Result[k].UserID
				todos[userID] = append(todos[userID], tTask.Result[k])
			}
		}
	}

	pool.StopAndWait()

	close(uChan)
	close(pChan)
	close(tChan)
	close(cChan)

	for cTask := range cChan {
		if cTask.Err != nil {
			aggErr = multierr.Append(aggErr, cTask.Err)
			continue
		}
		for k := 0; k < len(cTask.Result); k++ {
			postID := cTask.Result[k].PostID
			comments[postID] = append(comments[postID], cTask.Result[k])
		}
	}

	if aggErr != nil {
		return nil, aggErr
	}

	var result []model.UserDTO
	for i := 0; i < len(users); i++ {
		userResponse := &users[i]
		userDTO := new(model.UserDTO)
		userDTO.Posts = make([]model.PostDTO, 0)
		userDTO.Todos = make([]model.TodoDTO, 0)
		userDTO.ID = userResponse.ID
		userDTO.Name = userResponse.Name
		userDTO.Email = userResponse.Email
		userDTO.Gender = userResponse.Gender
		userDTO.Status = userResponse.Status

		if posts[userResponse.ID] != nil {
			for k := 0; k < len(posts[userResponse.ID]); k++ {
				postResponse := &posts[userResponse.ID][k]
				postDTO := new(model.PostDTO)
				postDTO.Comments = make([]model.CommentDTO, 0)
				postDTO.ID = postResponse.ID
				postDTO.Title = postResponse.Title
				postDTO.Body = postResponse.Body

				if comments[postDTO.ID] != nil {
					for j := 0; j < len(comments[postDTO.ID]); j++ {
						commentResponse := &comments[postDTO.ID][j]
						commentDTO := new(model.CommentDTO)
						commentDTO.ID = commentResponse.ID
						commentDTO.Name = commentResponse.Name
						commentDTO.Email = commentResponse.Email
						commentDTO.Body = commentResponse.Body
						postDTO.Comments = append(postDTO.Comments, *commentDTO)
					}
				}

				userDTO.Posts = append(userDTO.Posts, *postDTO)
			}
		}

		if todos[userResponse.ID] != nil {
			for k := 0; k < len(todos[userResponse.ID]); k++ {
				todoResponse := &todos[userResponse.ID][k]
				todoDTO := new(model.TodoDTO)
				todoDTO.ID = todoResponse.ID
				todoDTO.Title = todoResponse.Title
				todoDTO.DueOn = todoResponse.DueOn
				todoDTO.Status = todoResponse.Status
				userDTO.Todos = append(userDTO.Todos, *todoDTO)
			}
		}

		result = append(result, *userDTO)
	}

	slices.SortFunc(result, func(a, b model.UserDTO) int {
		return a.ID - b.ID
	})

	return result, nil
}
