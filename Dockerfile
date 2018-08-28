FROM golang

WORKDIR $GOPATH/src/couch-sso
ADD . $GOPATH/src/couch-sso

RUN go build

EXPOSE 8008

CMD ./couch-sso
