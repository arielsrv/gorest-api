package services

import (
	"github.com/sourcegraph/conc/pool"
	"runtime"
	"slices"
	"sync"

	"github.com/alitto/pond"
	"gitlab.com/iskaypetcom/digital/oms/api-core/gorest-api/src/app/clients"
	"gitlab.com/iskaypetcom/digital/oms/api-core/gorest-api/src/app/model"
	"go.uber.org/multierr"
)

type IUsersService interface {
	GetUsers() ([]model.UserDTO, error)
	GetUsers2() ([]model.UserDTO, error)
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

	uChan := make(chan Task[*model.UserResponse], len(userResponses))
	pChan := make(chan Task[[]model.PostResponse], len(userResponses))
	tChan := make(chan Task[[]model.TodoResponse], len(userResponses))

	var aggErr error
	var users []model.UserResponse
	posts := make(map[int][]model.PostResponse)
	todos := make(map[int][]model.TodoResponse)

	var mtx sync.RWMutex
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

				commentsPool := pond.New(runtime.NumCPU()-1, 100)
				cChan := make(chan Task[[]model.CommentResponse], len(userResponses))

				for k := 0; k < len(response); k++ {
					postID := response[k].ID
					commentsPool.Submit(func() {
						cChan <- ToTask[[]model.CommentResponse](func() ([]model.CommentResponse, error) {
							return r.userClient.GetComments(postID)
						})
					})
				}

				commentsPool.Submit(func() {
					for k := 0; k < len(response); k++ {
						cTask := <-cChan
						if cTask.Err != nil {
							aggErr = multierr.Append(aggErr, cTask.Err)
							continue
						}
						for j := 0; j < len(cTask.Result); j++ {
							postID := cTask.Result[j].PostID
							mtx.Lock()
							comments[postID] = append(comments[postID], cTask.Result[j])
							mtx.Unlock()
						}
					}
				})

				commentsPool.StopAndWait()

				return response, nil
			})
		})

		pool.Submit(func() {
			tChan <- ToTask[[]model.TodoResponse](func() ([]model.TodoResponse, error) {
				return r.userClient.GetTodos(userID)
			})
		})
	}

	pool.Submit(func() {
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
	})

	pool.StopAndWait()

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

func (r *UsersService) GetUsers2() ([]model.UserDTO, error) {
	var users []model.UserDTO
	posts := make(map[int][]model.PostDTO)
	todos := make(map[int][]model.TodoDTO)

	userResponses, err := r.userClient.GetUsers()
	if err != nil {
		return nil, err
	}

	uChan := make(chan Task[*model.UserResponse])
	pChan := make(chan Task[[]model.PostResponse])
	tChan := make(chan Task[[]model.TodoResponse])

	var aggErr error

	consume := pool.New()

	consume.Go(func() {
		for userTask := range uChan {
			if userTask.Err != nil {
				aggErr = multierr.Append(aggErr, userTask.Err)
				continue
			}

			userDTO := model.UserDTO{
				Posts:  make([]model.PostDTO, 0),
				Todos:  make([]model.TodoDTO, 0),
				ID:     userTask.Result.ID,
				Name:   userTask.Result.Name,
				Email:  userTask.Result.Email,
				Gender: userTask.Result.Gender,
				Status: userTask.Result.Status,
			}

			users = append(users, userDTO)
		}
	})

	consume.Go(func() {
		for postTask := range pChan {
			if postTask.Err != nil {
				aggErr = multierr.Append(aggErr, postTask.Err)
				continue
			}

			for i := 0; i < len(postTask.Result); i++ {
				userID := postTask.Result[i].UserID

				postDTO := model.PostDTO{
					Comments: make([]model.CommentDTO, 0),
					ID:       postTask.Result[i].ID,
					Title:    postTask.Result[i].Title,
					Body:     postTask.Result[i].Body,
				}

				posts[userID] = append(posts[userID], postDTO)
			}
		}
	})

	consume.Go(func() {
		for todoTask := range tChan {
			if todoTask.Err != nil {
				aggErr = multierr.Append(aggErr, todoTask.Err)
				continue
			}
			for i := 0; i < len(todoTask.Result); i++ {
				userID := todoTask.Result[i].UserID

				todoDTO := model.TodoDTO{
					ID:     todoTask.Result[i].ID,
					Title:  todoTask.Result[i].Title,
					DueOn:  todoTask.Result[i].DueOn,
					Status: todoTask.Result[i].Status,
				}

				todos[userID] = append(todos[userID], todoDTO)
			}
		}
	})

	produce := pool.New()
	for i := 0; i < len(userResponses); i++ {
		produce.Go(func() {
			uChan <- ToTask[*model.UserResponse](func() (*model.UserResponse, error) {
				return r.userClient.GetUser(userResponses[i].ID)
			})
		})
		produce.Go(func() {
			pChan <- ToTask[[]model.PostResponse](func() ([]model.PostResponse, error) {
				return r.userClient.GetPosts(userResponses[i].ID)
			})
		})
		produce.Go(func() {
			tChan <- ToTask[[]model.TodoResponse](func() ([]model.TodoResponse, error) {
				return r.userClient.GetTodos(userResponses[i].ID)
			})
		})
	}

	produce.Wait()

	if aggErr != nil {
		return nil, aggErr
	}

	for i := 0; i < len(users); i++ {
		if posts[users[i].ID] != nil {
			users[i].Posts = append(users[i].Posts, posts[users[i].ID]...)
		}
		if todos[users[i].ID] != nil {
			users[i].Todos = append(users[i].Todos, todos[users[i].ID]...)
		}
	}

	return users, nil
}
