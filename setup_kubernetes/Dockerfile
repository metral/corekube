FROM google/golang:stable

RUN apt-get update -y
RUN apt-get install net-tools -y

WORKDIR /gopath/src/setup_kubernetes
ADD . /gopath/src/setup_kubernetes/
RUN go get ./...

CMD []
ENTRYPOINT ["/gopath/bin/setup_kubernetes"]
