package app

import (
	"net/http"

	"gitlab.com/iskaypetcom/digital/oms/api-core/gorest-api/src/app/controllers"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/backend-api-sdk/v2/core/container"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/backend-api-sdk/v2/core/routing"
)

type Routes struct {
	routing.APIRoutes
}

func (r *Routes) Register() {
	r.AddRoute(http.MethodGet, "/users", container.Provide[controllers.IUsersController]().GetUsers)
	r.AddRoute(http.MethodGet, "/users2", container.Provide[controllers.IUsersController]().GetUsers2)
}
