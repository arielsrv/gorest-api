package app

import (
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/backend-api-sdk/v2/core/application"
)

type Application struct {
	application.APIApplication
}

func (r *Application) Init() {
	r.UseMetrics()
	r.UseSwagger()
	r.UseConfig()

	r.RegisterDependencyInjectionModule(new(ApplicationModule))
	r.RegisterRoutes(new(Routes))

	r.Build()
}
