package services

import (
	"slices"

	"github.com/sourcegraph/conc/pool"
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

func (r *UsersService) GetComments(posts []model.PostDTO) ([]model.CommentDTO, error) {
	var comments []model.CommentDTO

	cChan := make(chan Task[[]model.CommentResponse])

	var aggErr error
	consume := pool.New()
	consume.Go(func() {
		for commentTask := range cChan {
			if commentTask.Err != nil {
				aggErr = multierr.Append(aggErr, commentTask.Err)
				continue
			}

			for i := 0; i < len(commentTask.Result); i++ {
				commentDTO := model.CommentDTO{
					ID:     commentTask.Result[i].ID,
					PostID: commentTask.Result[i].PostID,
					Name:   commentTask.Result[i].Name,
					Email:  commentTask.Result[i].Email,
					Body:   commentTask.Result[i].Body,
				}
				comments = append(comments, commentDTO)
			}
		}
	})

	produce := pool.New()
	for i := 0; i < len(posts); i++ {
		produce.Go(func() {
			cChan <- ToTask[[]model.CommentResponse](func() ([]model.CommentResponse, error) {
				return r.userClient.GetComments(posts[i].ID)
			})
		})
	}

	produce.Wait()

	if aggErr != nil {
		return nil, aggErr
	}

	return comments, nil
}

func (r *UsersService) GetUsers() ([]model.UserDTO, error) {
	var users []model.UserDTO

	userResponses, aggErr := r.userClient.GetUsers()
	if aggErr != nil {
		return nil, aggErr
	}

	uChan := make(chan Task[*model.UserResponse])
	pChan := make(chan Task[[]model.PostResponse])
	tChan := make(chan Task[[]model.TodoResponse])

	posts := make(map[int][]model.PostDTO)
	todos := make(map[int][]model.TodoDTO)

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

				commentsResponse, err := r.userClient.GetComments(postDTO.ID)
				if err != nil {
					aggErr = multierr.Append(aggErr, err)
					continue
				}

				for k := 0; k < len(commentsResponse); k++ {
					commentDTO := model.CommentDTO{
						ID:     commentsResponse[k].ID,
						PostID: commentsResponse[k].PostID,
						Name:   commentsResponse[k].Name,
						Email:  commentsResponse[k].Email,
						Body:   commentsResponse[k].Body,
					}
					postDTO.Comments = append(postDTO.Comments, commentDTO)
				}

				slices.SortFunc(postDTO.Comments, func(a, b model.CommentDTO) int {
					return a.ID - b.ID
				})

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

	slices.SortFunc(users, func(user1, user2 model.UserDTO) int {
		return user1.ID - user2.ID
	})

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
