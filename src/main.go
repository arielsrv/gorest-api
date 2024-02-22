package main

import (
	"gitlab.com/iskaypetcom/digital/oms/api-core/gorest-api/src/app"
	_ "gitlab.com/iskaypetcom/digital/oms/api-core/gorest-api/src/resources/docs"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/backend-api-sdk/v2/core"
	log "gitlab.com/iskaypetcom/digital/sre/tools/dev/go-logger"
)

// @title IskayPet gorest-api
// @description Provide an interface to HTTP responses.
// @basePath /
// @version v1.
func main() {
	server := core.NewServer()

	server.On(new(app.Application))
	server.Start()

	if err := server.Join(); err != nil {
		log.Fatal(err)
	}
}
