FROM golang:1.8

WORKDIR /go/src/github.com/oscp/openshift-selfservice/server

COPY . /go/src/github.com/oscp/openshift-selfservice

RUN go get gopkg.in/gin-gonic/gin.v1 \
    && go get gopkg.in/appleboy/gin-jwt.v2 \
    && go get gopkg.in/dgrijalva/jwt-go.v3 \
    && go get github.com/jtblin/go-ldap-client

RUN go install -v

CMD ["server"]