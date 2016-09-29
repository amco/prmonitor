FROM golang:1.7
ENV HOME $GOPATH/src/github.com/brentdrich/prmonitor

ADD . $HOME
WORKDIR $HOME

RUN go get github.com/tools/godep
RUN godep restore
RUN godep go install -v ./cmd/prmonitor

CMD prmonitor
