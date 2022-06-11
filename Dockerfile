FROM golang:1.18.3-alpine3.16
WORKDIR /server

COPY go.mod ./
COPY go.sum ./

RUN go mod download

RUN mkdir ./bin

COPY . ./

RUN go build -o ./bin ./...

CMD ["bin/kaspa"]