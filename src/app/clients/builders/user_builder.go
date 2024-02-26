package http

import (
	"time"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/rest"
)

func NewUserRequestBuilder() *rest.RequestBuilder {
	return &rest.RequestBuilder{
		Name:           "gorest-client",
		BaseURL:        "https://gorest.co.in/public/v2",
		Timeout:        time.Duration(5000) * time.Millisecond,
		ConnectTimeout: time.Duration(10000) * time.Millisecond,
	}
}
