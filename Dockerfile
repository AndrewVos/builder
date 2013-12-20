FROM ubuntu

#golang
RUN apt-get install -y --force-yes curl && \
    curl -O https://go.googlecode.com/files/go1.1.2.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.1.2.linux-amd64.tar.gz
ENV GOPATH /gopath
ENV PATH $PATH:$GOPATH/bin:/usr/local/go/bin

#install git
RUN apt-get install -y --force-yes git-core

ADD . /gopath/src/github.com/AndrewVos/builder
WORKDIR /gopath/src/github.com/AndrewVos/builder
RUN go get
RUN go build
RUN mv /gopath/src/github.com/AndrewVos/builder /builder
