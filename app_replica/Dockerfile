## builder
FROM golang:1.14-alpine3.13 as builder

RUN apk update && apk upgrade

WORKDIR /go/src/github.com/leemov/redisclusterintro/app_replica

COPY main.go .

RUN go mod init github.com/leemov/redisclusterintro/app_replica
RUN go mod vendor
RUN go build -o app main.go

## Distribution 

FROM alpine:latest

RUN apk update && apk upgrade

WORKDIR /app

EXPOSE 8080

COPY --from=builder /go/src/github.com/leemov/redisclusterintro/app_replica/app /app

CMD /app/app
