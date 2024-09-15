FROM golang:1.22 as local_app

RUN #apt-get -y update && apt-get -y install htop git gettext make && apt-get -y clean

WORKDIR $GOPATH/src/app/

CMD /bin/bash