package http

import (
	"time"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/rest"
)

func NewUserRequestBuilder() *rest.RequestBuilder {
	return &rest.RequestBuilder{
		BaseURL:        "https://gorest.co.in/public/v2",
		Timeout:        time.Duration(2500) * time.Millisecond,
		ConnectTimeout: time.Duration(5000) * time.Millisecond,
		CustomPool: &rest.CustomPool{
			MaxIdleConnsPerHost: 200,
		},
		Name: "gorest-client",
	}
}
