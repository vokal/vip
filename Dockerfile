FROM ubuntu:trusty
MAINTAINER Scott Ferguson <scott.ferguson@vokalinteractive.com>

RUN apt-get update -qq
RUN apt-get install -y ca-certificates libvips37 \
    && apt-get install automake build-essential git gobject-introspection libglib2.0-dev libjpeg-turbo8-dev libpng12-dev gtk-doc-tools \
    && git clone https://github.com/jcupitt/libvips.git \
    && cd libvips \
    && ./bootstrap.sh \
    && ./configure --enable-debug=no --without-python --without-fftw --without-libexif --without-libgf --without-little-cms --without-orc --without-pango --prefix=/usr \
    && make \
    && make install \
    && ldconfig
RUN mkdir /etc/vip

ADD ./vip ./vip

EXPOSE 8080
 
CMD ./vip
