FROM ubuntu:latest
MAINTAINER kookxiang@gmail.com

# Install required packages
RUN \
    apt-get -qq update && \
    apt-get -qq -y install curl cron git

# Install Go
RUN \
    mkdir -p /golang && \
    curl https://storage.googleapis.com/golang/go1.6.2.linux-amd64.tar.gz | tar xvzf - -C /golang --strip-components=1 > /dev/null && \
    mkdir /usr/go

# Set environment variables.
ENV GOROOT /golang
ENV GOPATH /usr/go
ENV PATH $GOROOT/bin:$GOPATH/bin:$PATH
ENV HOME /root

# Define working directory.
WORKDIR /root

# Add files
ADD *.go /root
ADD TiebaSign /root/TiebaSign

# Build binary
RUN \
    cd /root && \
    mkdir /root/cookies && \
    export GOPATH=/usr/go && \
    go get github.com/bitly/go-simplejson && \
    go get golang.org/x/text/encoding && \
    go get golang.org/x/text/encoding/simplifiedchinese && \
    go get golang.org/x/text/transform && \
    go build -o signer

# Remove Golang files to shrink image size
RUN rm -rf /golang /usr/go

# Add crontab
RUN \
    echo "0 1,22 * * * cd /root; ./signer -retry=10 -batch >> log.txt" > /etc/cron.d/TiebaSign && \
    chmod 0644 /etc/cron.d/TiebaSign && \
    touch /root/log.txt

CMD cron && tail -f /root/log.txt
