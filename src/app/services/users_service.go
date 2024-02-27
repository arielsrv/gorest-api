package services

import (
	"cmp"
	"runtime"
	"slices"

	"gitlab.com/iskaypetcom/digital/oms/api-core/gorest-api/src/app/model/paging"

	"gitlab.com/iskaypetcom/digital/oms/api-core/gorest-api/src/app/tpl"

	"gitlab.com/iskaypetcom/digital/oms/api-core/gorest-api/src/app/clients"
	"gitlab.com/iskaypetcom/digital/oms/api-core/gorest-api/src/app/model"
	"go.uber.org/multierr"
)

type IUsersService interface {
	GetUsers(page int, perPage int) (*paging.PagedResultDTO[model.UserDTO], error)
}

type UsersService struct {
	userClient clients.IUserClient
}

func NewUserService(userClient clients.IUserClient) *UsersService {
	return &UsersService{
		userClient: userClient,
	}
}

func (r *UsersService) GetUsers(page int, perPage int) (*paging.PagedResultDTO[model.UserDTO], error) {
	pagedResult, aggErr := r.userClient.GetUsers(page, perPage)
	if aggErr != nil {
		return nil, aggErr
	}

	pool := tpl.NewWorkerPool13[model.UserResponse, model.UserDTO, model.PostDTO, model.TodoDTO]()

	var users []model.UserDTO
	err := pool.Zip(pagedResult.Results, r.getUsers, r.getPosts, r.getTodos,
		func(usersDTOs []model.UserDTO, postDTOs []model.PostDTO, todoDTOs []model.TodoDTO, err error) {
			if err != nil {
				aggErr = multierr.Append(aggErr, err)
			}

			for i := 0; i < len(usersDTOs); i++ {
				userDTO := &usersDTOs[i]

				userDTO.Posts = make([]model.PostDTO, 0)
				for k := 0; k < len(postDTOs); k++ {
					postDTO := postDTOs[k]
					if postDTO.UserID == userDTO.ID {
						userDTO.Posts = append(userDTO.Posts, postDTO)
					}
				}

				userDTO.Todos = make([]model.TodoDTO, 0)
				for k := 0; k < len(todoDTOs); k++ {
					todoDTO := todoDTOs[k]
					if todoDTO.UserID == userDTO.ID {
						userDTO.Todos = append(userDTO.Todos, todoDTO)
					}
				}

				users = append(users, *userDTO)
			}

			slices.SortFunc(users, func(a, b model.UserDTO) int {
				return cmp.Compare(a.ID, b.ID)
			})
		})

	if err != nil {
		return nil, err
	}

	return &paging.PagedResultDTO[model.UserDTO]{
		Limit:   pagedResult.Limit,
		Page:    pagedResult.Page,
		Pages:   pagedResult.Pages,
		Total:   pagedResult.Total,
		Results: users,
	}, nil
}

func (r *UsersService) getPosts(userResponses []model.UserResponse) ([]model.PostDTO, error) {
	var (
		posts  []model.PostDTO
		aggErr error
	)

	rChan := make(chan tpl.Task[[]model.PostResponse], len(userResponses))

	pool := tpl.New().WithMaxGoroutines(2)

	pool.Submit(func() {
		for i := 0; i < len(userResponses); i++ {
			task := <-rChan
			if task.Err != nil {
				aggErr = multierr.Append(aggErr, task.Err)
				continue
			}

			for k := 0; k < len(task.Result); k++ {
				postDTO := model.PostDTO{
					Comments: make([]model.CommentDTO, 0),
					ID:       task.Result[k].ID,
					UserID:   task.Result[k].UserID,
					Title:    task.Result[k].Title,
					Body:     task.Result[k].Body,
				}
				posts = append(posts, postDTO)
			}
		}

		var comments []model.CommentDTO

		child := tpl.New().WithMaxGoroutines(1)
		child.Submit(func() {
			commentsDTO, err := r.getComments(posts)
			if err != nil {
				aggErr = multierr.Append(aggErr, err)
				return
			}

			comments = append(comments, commentsDTO...)
		})

		child.Wait()

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
				return cmp.Compare(a.ID, b.ID)
			})
		}
	})

	pool.Submit(func() {
		tpl.ForEach(userResponses, func(userResponse *model.UserResponse) {
			rChan <- tpl.ToTask[[]model.PostResponse](func() ([]model.PostResponse, error) {
				return r.userClient.GetPosts(userResponse.ID)
			})
		}, runtime.NumCPU()-1)
	})

	pool.Wait()
	close(rChan)

	return posts, aggErr
}

func (r *UsersService) getTodos(userResponses []model.UserResponse) ([]model.TodoDTO, error) {
	var (
		todos  []model.TodoDTO
		aggErr error
	)

	rChan := make(chan tpl.Task[[]model.TodoResponse], len(userResponses))

	pool := tpl.New().WithMaxGoroutines(2)
	pool.Submit(func() {
		for i := 0; i < len(userResponses); i++ {
			task := <-rChan
			if task.Err != nil {
				aggErr = multierr.Append(aggErr, task.Err)
				return
			}
			for k := 0; k < len(task.Result); k++ {
				todoDTO := model.TodoDTO{
					ID:     task.Result[k].ID,
					UserID: task.Result[k].UserID,
					Title:  task.Result[k].Title,
					DueOn:  task.Result[k].DueOn,
					Status: task.Result[k].Status,
				}

				todos = append(todos, todoDTO)
			}
		}
	})

	pool.Submit(func() {
		tpl.ForEach(userResponses, func(userResponse *model.UserResponse) {
			rChan <- tpl.ToTask[[]model.TodoResponse](func() ([]model.TodoResponse, error) {
				return r.userClient.GetTodos(userResponse.ID)
			})
		}, runtime.NumCPU()-1)
	})

	pool.Wait()
	close(rChan)

	return todos, aggErr
}

func (r *UsersService) getUsers(userResponses []model.UserResponse) ([]model.UserDTO, error) {
	var (
		users  []model.UserDTO
		aggErr error
	)

	rChan := make(chan tpl.Task[*model.UserResponse], len(userResponses))
	pool := tpl.New().WithMaxGoroutines(2)

	pool.Submit(func() {
		for i := 0; i < len(userResponses); i++ {
			task := <-rChan
			if task.Err != nil {
				aggErr = multierr.Append(aggErr, task.Err)
				continue
			}

			userDTO := &model.UserDTO{
				ID:     task.Result.ID,
				Name:   task.Result.Name,
				Email:  task.Result.Email,
				Gender: task.Result.Gender,
				Status: task.Result.Status,
			}

			users = append(users, *userDTO)
		}
	})

	pool.Submit(func() {
		tpl.ForEach(userResponses, func(userResponse *model.UserResponse) {
			rChan <- tpl.ToTask[*model.UserResponse](func() (*model.UserResponse, error) {
				return r.userClient.GetUser(userResponse.ID)
			})
		}, runtime.NumCPU()-1)
	})

	pool.Wait()
	close(rChan)

	return users, aggErr
}

func (r *UsersService) getComments(posts []model.PostDTO) ([]model.CommentDTO, error) {
	var (
		comments []model.CommentDTO
		aggErr   error
	)

	rChan := make(chan tpl.Task[[]model.CommentResponse], len(posts))

	pool := tpl.New().WithMaxGoroutines(2)
	pool.Submit(func() {
		for i := 0; i < len(posts); i++ {
			commentTask := <-rChan
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

	pool.Submit(func() {
		tpl.ForEach(posts, func(postDTO *model.PostDTO) {
			rChan <- tpl.ToTask[[]model.CommentResponse](func() ([]model.CommentResponse, error) {
				return r.userClient.GetComments(postDTO.ID)
			})
		}, runtime.NumCPU()-1)
	})

	pool.Wait()
	close(rChan)

	return comments, aggErr
}
