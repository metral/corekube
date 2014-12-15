FROM google/golang:stable

RUN apt-get update -y
RUN apt-get install net-tools -y

WORKDIR /gopath/src/overlord
ADD . /gopath/src/overlord/
RUN go get ./...

CMD []
ENTRYPOINT ["/gopath/bin/overlord"]
