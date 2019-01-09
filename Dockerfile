FROM golang:latest
ADD . /go/src/github.com/madflojo/hord
WORKDIR /go/src/github.com/madflojo/hord/cmd/hord
RUN go install -v
CMD ["hord"]
