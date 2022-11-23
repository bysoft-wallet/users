# syntax=docker/dockerfile:1

## Build
FROM golang:1.19-buster AS build

ARG user

RUN useradd -u 1001 -m $user

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./

RUN  CGO_ENABLED=0 go build -o /bysoft-users cmd/main.go

## Deploy
FROM golang:1.19-alpine3.15

ARG app_port=8809
ARG user

WORKDIR /app

COPY --from=build /bysoft-users /app/bysoft-users

RUN addgroup -S $user && adduser -S $user -G $user

RUN chown -R $user:$user /app
RUN chmod +x /app/bysoft-users

USER $user
RUN mkdir -p /app/logs
RUN chmod -R 777 /app/logs

EXPOSE $app_port

CMD ["/app/bysoft-users"]