FROM google/golang:stable

RUN apt-get update -y
RUN apt-get install net-tools -y

WORKDIR /gopath/src/setup
ADD . /gopath/src/setup/
RUN go build setup/lib
RUN go get setup

CMD []
ENTRYPOINT ["/gopath/bin/setup"]
