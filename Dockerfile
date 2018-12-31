FROM golang:latest

ADD . /go/src/github.com/madflojo/hord

#RUN go install github.com/madflojo/hord

CMD ["hord"]
