FROM golang:alpine

WORKDIR /opt/app

COPY go.sum go.sum
COPY go.mod go.mod
RUN go mod download

COPY . .
RUN go get Homework
RUN go install ./cmd/grpc
ENTRYPOINT /go/bin/grpc

EXPOSE 8080
EXPOSE 8081