FROM stackbrew/ubuntu:raring

#golang
RUN apt-get install -y --force-yes curl && \
    curl -O https://go.googlecode.com/files/go1.1.2.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.1.2.linux-amd64.tar.gz
ENV GOPATH /gopath
ENV PATH $PATH:$GOPATH/bin:/usr/local/go/bin

#install git
RUN apt-get install -y --force-yes git-core

#postgres
RUN locale-gen en_US.UTF-8
RUN update-locale LANG=en_US.UTF-8
RUN apt-get -y --force-yes install wget && \
    echo 'deb http://apt.postgresql.org/pub/repos/apt/ squeeze-pgdg main' >> /etc/apt/sources.list.d/pgdg.list && \
    wget --quiet -O - http://apt.postgresql.org/pub/repos/apt/ACCC4CF8.asc | apt-key add - && \
    apt-get -y update && \
    apt-get -y --force-yes install postgresql postgresql-contrib libpq-dev

ADD . /gopath/src/github.com/AndrewVos/builder
WORKDIR /gopath/src/github.com/AndrewVos/builder
