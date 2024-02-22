package app

import (
	"gitlab.com/iskaypetcom/digital/oms/api-core/gorest-api/src/app/clients"
	http "gitlab.com/iskaypetcom/digital/oms/api-core/gorest-api/src/app/clients/builders"
	"gitlab.com/iskaypetcom/digital/oms/api-core/gorest-api/src/app/controllers"
	"gitlab.com/iskaypetcom/digital/oms/api-core/gorest-api/src/app/services"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/backend-api-sdk/v2/core/container"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/rest"
	"go.uber.org/dig"
)

type ApplicationModule struct {
	container.DependencyInjectionModule
}

func (r *ApplicationModule) Configure() {
	r.Bind(http.NewUserRequestBuilder, dig.As(new(rest.IRequestBuilder)))
	r.Bind(clients.NewUserClient, dig.As(new(clients.IUserClient)))
	r.Bind(services.NewUserService, dig.As(new(services.IUsersService)))
	r.Bind(controllers.NewUsersController, dig.As(new(controllers.IUsersController)))
}
