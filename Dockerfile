FROM ubuntu:trusty
MAINTAINER Scott Ferguson <scott.ferguson@vokalinteractive.com>

RUN apt-get update -qq
RUN apt-get install -y ca-certificates libvips37
RUN mkdir /etc/vip

ADD ./vip ./vip

EXPOSE 8080
 
CMD ./vip
