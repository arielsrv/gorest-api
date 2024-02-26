package controllers

import (
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/backend-api-sdk/v2/core"
	"net/http"
	"strconv"

	"gitlab.com/iskaypetcom/digital/oms/api-core/gorest-api/src/app/services"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/backend-api-sdk/v2/core/routing"
)

type IUsersController interface {
	GetUsers(ctx *routing.HTTPContext) error
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
	pageValue := ctx.Query("page", "1")
	page, err := strconv.Atoi(pageValue)
	if err != nil {
		return core.NewAPIErr(http.StatusBadRequest, err)
	}

	perPageValue := ctx.Query("per_page", "10")
	perPage, err := strconv.Atoi(perPageValue)
	if err != nil {
		return core.NewAPIErr(http.StatusBadRequest, err)
	}

	pagedResultDTO, err := r.usersService.GetUsers(page, perPage)
	if err != nil {
		return err
	}

	return ctx.JSON(pagedResultDTO)
}
