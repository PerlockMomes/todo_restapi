FROM golang:1.24.0 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o todo-app .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/todo-app .

COPY web ./web

EXPOSE 7540

CMD ["./todo-app"]