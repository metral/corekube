FROM google/golang

WORKDIR /gopath/src/app
ADD . /gopath/src/app/
ADD /run/systemd/system/etcd.service.d/20-cloudinit.conf /gopath/src/app/
RUN go get app

CMD []
ENTRYPOINT ["/gopath/bin/app"]
