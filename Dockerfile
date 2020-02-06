FROM golang:1.13.7-alpine AS binary

ENV WORKDIR /go/src/github.com/41North/kompoze

WORKDIR $WORKDIR

RUN apk -U add openssl git

COPY . $WORKDIR

RUN go install $WORKDIR/cmd/kompoze/main.go

FROM alpine:3.10

COPY --from=binary /go/bin/main /usr/local/bin/kompoze

RUN chmod +x /usr/local/bin/kompoze

ENTRYPOINT ["kompoze"]
CMD ["--help"]