FROM golang:1.24.3-alpine

WORKDIR /home/app

COPY . .

RUN go install github.com/air-verse/air@latest

RUN go install github.com/go-delve/delve/cmd/dlv@latest

CMD ["air", "./cmd/todo-app"]