FROM ubuntu
MAINTAINER Scott Ferguson <scott.ferguson@vokalinteractive.com>

RUN apt-get update -qq
RUN apt-get install -qqy ca-certificates

ADD ./vip ./vip
 
CMD ./vip
