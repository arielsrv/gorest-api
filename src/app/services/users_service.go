package services

import (
	"github.com/alitto/pond"
	"runtime"
	"slices"

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
	var users []model.UserDTO

	userResponses, aggErr := r.userClient.GetUsers()
	if aggErr != nil {
		return nil, aggErr
	}

	uChan := make(chan Task[*model.UserResponse], len(userResponses))
	pChan := make(chan Task[[]model.PostResponse], len(userResponses))
	tChan := make(chan Task[[]model.TodoResponse], len(userResponses))

	var posts []model.PostDTO
	var todos []model.TodoDTO

	produce := pond.New(runtime.NumCPU()-1, len(userResponses)*3)

	produce.Submit(func() {
		for i := 0; i < len(userResponses); i++ {
			userTask := <-uChan
			if userTask.Err != nil {
				aggErr = multierr.Append(aggErr, userTask.Err)
				return
			}

			userDTO := &model.UserDTO{
				ID:     userTask.Result.ID,
				Name:   userTask.Result.Name,
				Email:  userTask.Result.Email,
				Gender: userTask.Result.Gender,
				Status: userTask.Result.Status,
				Posts:  make([]model.PostDTO, 0),
				Todos:  make([]model.TodoDTO, 0),
			}

			users = append(users, *userDTO)
		}
	})

	produce.Submit(func() {
		for i := 0; i < len(userResponses); i++ {
			postTask := <-pChan
			if postTask.Err != nil {
				aggErr = multierr.Append(aggErr, postTask.Err)
				return
			}

			for k := 0; k < len(postTask.Result); k++ {
				postDTO := model.PostDTO{
					Comments: make([]model.CommentDTO, 0),
					ID:       postTask.Result[k].ID,
					UserID:   postTask.Result[k].UserID,
					Title:    postTask.Result[k].Title,
					Body:     postTask.Result[k].Body,
				}

				commentsResponse, err := r.userClient.GetComments(postDTO.ID)
				if err != nil {
					aggErr = multierr.Append(aggErr, err)
					continue
				}

				for j := 0; j < len(commentsResponse); j++ {
					commentDTO := &model.CommentDTO{
						ID:     commentsResponse[j].ID,
						PostID: commentsResponse[j].PostID,
						Name:   commentsResponse[j].Name,
						Email:  commentsResponse[j].Email,
						Body:   commentsResponse[j].Body,
					}
					postDTO.Comments = append(postDTO.Comments, *commentDTO)
				}

				slices.SortFunc(postDTO.Comments, func(a, b model.CommentDTO) int {
					return a.ID - b.ID
				})

				posts = append(posts, postDTO)
			}
		}
	})

	produce.Submit(func() {
		for i := 0; i < len(userResponses); i++ {
			todoTask := <-tChan
			if todoTask.Err != nil {
				aggErr = multierr.Append(aggErr, todoTask.Err)
				return
			}
			for k := 0; k < len(todoTask.Result); k++ {
				todoDTO := model.TodoDTO{
					ID:     todoTask.Result[k].ID,
					UserID: todoTask.Result[k].UserID,
					Title:  todoTask.Result[k].Title,
					DueOn:  todoTask.Result[k].DueOn,
					Status: todoTask.Result[k].Status,
				}

				todos = append(todos, todoDTO)
			}
		}
	})

	produce.Submit(func() {
		for i := 0; i < len(userResponses); i++ {
			uChan <- ToTask[*model.UserResponse](func() (*model.UserResponse, error) {
				return r.userClient.GetUser(userResponses[i].ID)
			})
		}
	})

	produce.Submit(func() {
		for i := 0; i < len(userResponses); i++ {
			pChan <- ToTask[[]model.PostResponse](func() ([]model.PostResponse, error) {
				return r.userClient.GetPosts(userResponses[i].ID)
			})
		}
	})

	produce.Submit(func() {
		for i := 0; i < len(userResponses); i++ {
			tChan <- ToTask[[]model.TodoResponse](func() ([]model.TodoResponse, error) {
				return r.userClient.GetTodos(userResponses[i].ID)
			})
		}
	})

	produce.StopAndWait()

	if aggErr != nil {
		return nil, aggErr
	}

	for i := 0; i < len(users); i++ {
		userDTO := &users[i]
		for k := 0; k < len(posts); k++ {
			postDTO := posts[k]
			if postDTO.UserID == userDTO.ID {
				userDTO.Posts = append(userDTO.Posts, postDTO)
			}
		}
		for k := 0; k < len(todos); k++ {
			todoDTO := todos[k]
			if todoDTO.UserID == userDTO.ID {
				userDTO.Todos = append(userDTO.Todos, todoDTO)
			}
		}
	}

	slices.SortFunc(users, func(a, b model.UserDTO) int {
		return a.ID - b.ID
	})

	return users, nil
}
