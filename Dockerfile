FROM golang:1.22

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

CMD ["cd app"]

COPY . .

RUN go build -o main .

EXPOSE 8081

CMD ["./main"]
