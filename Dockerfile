ARG GO_VERSION
FROM golang:${GO_VERSION} AS build
ARG GITLAB_TOKEN

ADD . /app
WORKDIR /app

RUN echo "machine gitlab.com login master_token password $GITLAB_TOKEN" > /root/.netrc
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o gorest-api src/*.go

FROM gcr.io/distroless/static-debian12 as release
COPY --from=build /app /release

WORKDIR /release

USER nonroot:nonroot
ENTRYPOINT ["./gorest-api"]
