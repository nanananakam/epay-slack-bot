FROM golang:1.12 as builder

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

COPY . /
RUN go build -o /main /main.go
CMD /main $PORT

