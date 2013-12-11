FROM ubuntu:13.04
MAINTAINER Scott Ferguson <scott.ferguson@vokalinteractive.com>

ADD ./vip ./vip
 
CMD ./vip
