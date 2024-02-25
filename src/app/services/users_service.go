package services

import (
	"runtime"
	"slices"

	"github.com/sourcegraph/conc/pool"

	"github.com/sourcegraph/conc/iter"

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
	userResponses, aggErr := r.userClient.GetUsers()
	if aggErr != nil {
		return nil, aggErr
	}

	var users []model.UserDTO

	parent := pool.New().WithMaxGoroutines(3)
	parent.Go(func() {
		child := pool.New().WithMaxGoroutines(2)
		uChan := make(chan Task[*model.UserResponse], len(userResponses))
		child.Go(func() {
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
				}

				users = append(users, *userDTO)
			}
		})

		child.Go(func() {
			maxGoroutines := runtime.NumCPU()
			if len(userResponses) < maxGoroutines {
				maxGoroutines = len(userResponses)
			}

			it := iter.Iterator[model.UserResponse]{
				MaxGoroutines: maxGoroutines,
			}

			it.ForEach(userResponses, func(userResponse *model.UserResponse) {
				uChan <- ToTask[*model.UserResponse](func() (*model.UserResponse, error) {
					return r.userClient.GetUser(userResponse.ID)
				})
			})
		})

		child.Wait()
	})

	var posts []model.PostDTO
	parent.Go(func() {
		child := pool.New().WithMaxGoroutines(2)
		pChan := make(chan Task[[]model.PostResponse], len(userResponses))
		child.Go(func() {
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
					posts = append(posts, postDTO)
				}
			}

			if aggErr != nil {
				return
			}

			var comments []model.CommentDTO
			cChan := make(chan Task[[]model.CommentResponse], len(posts))
			cChild := pool.New()
			cChild.Go(func() {
				for i := 0; i < len(posts); i++ {
					commentTask := <-cChan
					if commentTask.Err != nil {
						aggErr = multierr.Append(aggErr, commentTask.Err)
						return
					}
					for k := 0; k < len(commentTask.Result); k++ {
						commentDTO := &model.CommentDTO{
							ID:     commentTask.Result[k].ID,
							PostID: commentTask.Result[k].PostID,
							Name:   commentTask.Result[k].Name,
							Email:  commentTask.Result[k].Email,
							Body:   commentTask.Result[k].Body,
						}
						comments = append(comments, *commentDTO)
					}
				}
			})

			cChild.Go(func() {
				maxGoroutines := runtime.NumCPU()
				if len(userResponses) < maxGoroutines {
					maxGoroutines = len(posts)
				}

				it := iter.Iterator[model.PostDTO]{
					MaxGoroutines: maxGoroutines,
				}

				it.ForEach(posts, func(postDTO *model.PostDTO) {
					cChan <- ToTask[[]model.CommentResponse](func() ([]model.CommentResponse, error) {
						return r.userClient.GetComments(postDTO.ID)
					})
				})
			})

			cChild.Wait()

			if aggErr != nil {
				return
			}

			for i := 0; i < len(posts); i++ {
				post := &posts[i]
				for k := 0; k < len(comments); k++ {
					if post.ID == comments[k].PostID {
						post.Comments = append(post.Comments, comments[k])
					}
				}

				slices.SortFunc(post.Comments, func(a, b model.CommentDTO) int {
					return a.ID - b.ID
				})
			}
		})

		child.Go(func() {
			maxGoroutines := runtime.NumCPU()
			if len(userResponses) < maxGoroutines {
				maxGoroutines = len(userResponses)
			}

			it := iter.Iterator[model.UserResponse]{
				MaxGoroutines: maxGoroutines,
			}

			it.ForEach(userResponses, func(userResponse *model.UserResponse) {
				pChan <- ToTask[[]model.PostResponse](func() ([]model.PostResponse, error) {
					return r.userClient.GetPosts(userResponse.ID)
				})
			})
		})

		child.Wait()
	})

	var todos []model.TodoDTO
	parent.Go(func() {
		tChan := make(chan Task[[]model.TodoResponse], len(userResponses))
		child := pool.New().WithMaxGoroutines(2)
		child.Go(func() {
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

		child.Go(func() {
			maxGoroutines := runtime.NumCPU()
			if len(userResponses) < maxGoroutines {
				maxGoroutines = len(userResponses)
			}

			it := iter.Iterator[model.UserResponse]{
				MaxGoroutines: maxGoroutines,
			}

			it.ForEach(userResponses, func(userResponse *model.UserResponse) {
				tChan <- ToTask[[]model.TodoResponse](func() ([]model.TodoResponse, error) {
					return r.userClient.GetTodos(userResponse.ID)
				})
			})
		})

		child.Wait()
	})

	parent.Wait()

	if aggErr != nil {
		return nil, aggErr
	}

	for i := 0; i < len(users); i++ {
		userDTO := &users[i]

		userDTO.Posts = make([]model.PostDTO, 0)
		for k := 0; k < len(posts); k++ {
			postDTO := posts[k]
			if postDTO.UserID == userDTO.ID {
				userDTO.Posts = append(userDTO.Posts, postDTO)
			}
		}

		userDTO.Todos = make([]model.TodoDTO, 0)
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
