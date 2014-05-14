FROM ubuntu
MAINTAINER Scott Ferguson <scott.ferguson@vokalinteractive.com>

RUN apt-get update -qq
RUN apt-get install -y ca-certificates

ADD ./vip ./vip

EXPOSE 8080
 
CMD ./vip
