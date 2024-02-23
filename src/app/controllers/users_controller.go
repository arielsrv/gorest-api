package controllers

import (
	"gitlab.com/iskaypetcom/digital/oms/api-core/gorest-api/src/app/services"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/backend-api-sdk/v2/core/routing"
)

type IUsersController interface {
	GetUsers(ctx *routing.HTTPContext) error
	GetUsers2(ctx *routing.HTTPContext) error
}

type UsersController struct {
	usersService services.IUsersService
}

func NewUsersController(usersService services.IUsersService) *UsersController {
	return &UsersController{
		usersService: usersService,
	}
}

func (r UsersController) GetUsers(ctx *routing.HTTPContext) error {
	result, err := r.usersService.GetUsers()
	if err != nil {
		return err
	}

	return ctx.JSON(result)
}

func (r UsersController) GetUsers2(ctx *routing.HTTPContext) error {
	result, err := r.usersService.GetUsers2()
	if err != nil {
		return err
	}

	return ctx.JSON(result)
}
