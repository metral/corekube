FROM google/golang:stable

RUN apt-get update -y
RUN apt-get install net-tools -y

WORKDIR /gopath/src/app
ADD ./src /gopath/src/app/
RUN go get app

CMD []
ENTRYPOINT ["/gopath/bin/app"]
