FROM ubuntu:latest
MAINTAINER kookxiang@gmail.com

ENV HOME /root
RUN apt-get update -y
RUN apt-get install -y golang git cron
RUN mkdir /root/cookies
ADD *.go /root
ADD TiebaSign /root/TiebaSign
RUN mkdir /usr/go
RUN export GOPATH=/usr/go && go get github.com/bitly/go-simplejson
RUN export GOPATH=/usr/go && go get golang.org/x/text/encoding
RUN export GOPATH=/usr/go && go get golang.org/x/text/encoding/simplifiedchinese
RUN export GOPATH=/usr/go && go get golang.org/x/text/transform
RUN cd /root && export GOPATH=/usr/go && go build -o signer
RUN echo "0 1,22 * * * cd /root; ./signer -retry=10 -batch >> log.txt" > /etc/cron.d/TiebaSign
RUN chmod 0644 /etc/cron.d/TiebaSign
RUN touch /root/log.txt
CMD cron && tail -f /root/log.txt