package services

import (
	"github.com/sourcegraph/conc/pool"
	"gitlab.com/iskaypetcom/digital/oms/api-core/gorest-api/src/app/clients"
	"gitlab.com/iskaypetcom/digital/oms/api-core/gorest-api/src/app/model"
	"go.uber.org/multierr"
	"slices"
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

	uChan := make(chan Task[*model.UserResponse])
	pChan := make(chan Task[[]model.PostResponse])
	tChan := make(chan Task[[]model.TodoResponse])

	posts := make(map[int][]model.PostDTO)
	todos := make(map[int][]model.TodoDTO)
	comments := make(map[int][]model.CommentDTO)

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

	cChan := make(chan Task[[]model.CommentResponse])

	consume.Go(func() {
		for postTask := range pChan {
			if postTask.Err != nil {
				aggErr = multierr.Append(aggErr, postTask.Err)
				continue
			}

			produce := pool.New()

			for i := 0; i < len(postTask.Result); i++ {
				userID := postTask.Result[i].UserID

				postDTO := model.PostDTO{
					Comments: make([]model.CommentDTO, 0),
					ID:       postTask.Result[i].ID,
					Title:    postTask.Result[i].Title,
					Body:     postTask.Result[i].Body,
				}

				posts[userID] = append(posts[userID], postDTO)

				produce.Go(func() {
					cChan <- ToTask[[]model.CommentResponse](func() ([]model.CommentResponse, error) {
						return r.userClient.GetComments(postDTO.ID)
					})
				})
			}

			produce.Wait()
		}
	})

	consume.Go(func() {
		for commentsTask := range cChan {
			if commentsTask.Err != nil {
				aggErr = multierr.Append(aggErr, commentsTask.Err)
				continue
			}
			for i := 0; i < len(commentsTask.Result); i++ {
				postID := commentsTask.Result[i].PostID

				commentDTO := model.CommentDTO{
					ID:    commentsTask.Result[i].ID,
					Name:  commentsTask.Result[i].Name,
					Email: commentsTask.Result[i].Email,
					Body:  commentsTask.Result[i].Body,
				}

				comments[postID] = append(comments[postID], commentDTO)
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
			for k := 0; k < len(posts[users[i].ID]); k++ {
				if comments[posts[users[i].ID][k].ID] != nil {
					users[i].Posts[k].Comments = append(users[i].Posts[k].Comments, comments[posts[users[i].ID][k].ID]...)
				}
			}
		}
		if todos[users[i].ID] != nil {
			users[i].Todos = append(users[i].Todos, todos[users[i].ID]...)
		}
	}

	return users, nil
}
