package services

import (
	"runtime"
	"slices"

	"gitlab.com/iskaypetcom/digital/oms/api-core/gorest-api/src/app/gpars"

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

func (r *UsersService) GetUsers() ([]model.UserDTO, error) {
	userResponses, aggErr := r.userClient.GetUsers()
	if aggErr != nil {
		return nil, aggErr
	}

	pool := gpars.New().WithMaxGoroutines(3)

	var users []model.UserDTO
	pool.Go(func() {
		usersDTO, err := r.getUsers(userResponses)
		if err != nil {
			aggErr = multierr.Append(aggErr, err)
			return
		}
		users = append(users, usersDTO...)
	})

	var posts []model.PostDTO
	pool.Go(func() {
		postsDTO, err := r.getPosts(userResponses)
		if err != nil {
			aggErr = multierr.Append(aggErr, err)
			return
		}
		posts = append(posts, postsDTO...)
	})

	var todos []model.TodoDTO
	pool.Go(func() {
		todosDTO, err := r.getTodos(userResponses)
		if err != nil {
			aggErr = multierr.Append(aggErr, err)
			return
		}
		todos = append(todos, todosDTO...)
	})

	pool.Wait()

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

func (r *UsersService) getPosts(userResponses []model.UserResponse) ([]model.PostDTO, error) {
	var (
		posts  []model.PostDTO
		aggErr error
	)

	rChan := make(chan gpars.Task[[]model.PostResponse], len(userResponses))

	pool := gpars.New().WithMaxGoroutines(2)

	pool.Go(func() {
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

		child := gpars.New().WithMaxGoroutines(1)
		child.Go(func() {
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
				return a.ID - b.ID
			})
		}
	})

	pool.Go(func() {
		gpars.ForEach(userResponses, func(userResponse *model.UserResponse) {
			rChan <- gpars.ToTask[[]model.PostResponse](func() ([]model.PostResponse, error) {
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

	rChan := make(chan gpars.Task[[]model.TodoResponse], len(userResponses))

	pool := gpars.New().WithMaxGoroutines(2)
	pool.Go(func() {
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

	pool.Go(func() {
		gpars.ForEach(userResponses, func(userResponse *model.UserResponse) {
			rChan <- gpars.ToTask[[]model.TodoResponse](func() ([]model.TodoResponse, error) {
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

	rChan := make(chan gpars.Task[*model.UserResponse], len(userResponses))
	pool := gpars.New().WithMaxGoroutines(2)

	pool.Go(func() {
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

	pool.Go(func() {
		gpars.ForEach(userResponses, func(userResponse *model.UserResponse) {
			rChan <- gpars.ToTask[*model.UserResponse](func() (*model.UserResponse, error) {
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

	rChan := make(chan gpars.Task[[]model.CommentResponse], len(posts))

	pool := gpars.New().WithMaxGoroutines(2)
	pool.Go(func() {
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

	pool.Go(func() {
		gpars.ForEach(posts, func(postDTO *model.PostDTO) {
			rChan <- gpars.ToTask[[]model.CommentResponse](func() ([]model.CommentResponse, error) {
				return r.userClient.GetComments(postDTO.ID)
			})
		}, runtime.NumCPU()-1)
	})

	pool.Wait()
	close(rChan)

	return comments, aggErr
}
