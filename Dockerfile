FROM golang:1.23-alpine AS build_base

RUN apk add --update make
WORKDIR /survey-bot

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN go build -o survey-bot ./cmd/bot/main.go

FROM alpine:3.9

RUN apk add ca-certificates
WORKDIR /survey-bot

COPY --from=build_base /survey-bot/survey-bot ./

ENTRYPOINT ["/survey-bot/survey-bot"]