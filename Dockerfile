FROM ubuntu:trusty
MAINTAINER Scott Ferguson <scott.ferguson@vokalinteractive.com>

RUN apt-get update -qq
RUN apt-get install -y bzr wget ca-certificates libvips37 libxml2-dev \
    && apt-get -y install automake build-essential git gobject-introspection libglib2.0-dev libjpeg-turbo8-dev libpng12-dev gtk-doc-tools \
    && git clone https://github.com/jcupitt/libvips.git \
    && cd libvips \
    && ./bootstrap.sh \
    && ./configure --enable-debug=no --without-python --without-fftw --without-libexif --without-libgf --without-little-cms --without-orc --without-pango --prefix=/usr \
    && make \
    && make install \
    && ldconfig
RUN mkdir /etc/vip
RUN ln -s /usr/lib/libvips.so.42 /usr/lib/libvips.so.38
RUN wget https://storage.googleapis.com/golang/go1.4.2.linux-amd64.tar.gz && tar -C /usr/local -xzf go1.4.2.linux-amd64.tar.gz
ENV PATH=$PATH:/usr/local/go/bin
ENV GOPATH=/go
COPY . /go/src/vip
WORKDIR /go/src/vip
RUN go get && go build

EXPOSE 8080

CMD ./vip
