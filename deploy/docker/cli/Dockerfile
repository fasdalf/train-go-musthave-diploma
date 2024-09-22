FROM golang:1.22 as local_app

RUN #apt-get -y update && apt-get -y install htop git gettext make && apt-get -y clean
RUN go install github.com/go-delve/delve/cmd/dlv@1.23.0

WORKDIR $GOPATH/src/app/

CMD /bin/bash