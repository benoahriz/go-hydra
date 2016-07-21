FROM ubuntu:trusty
ARG APT_PROXY_PORT=
ARG HOST_IP=
COPY detect-apt-proxy.sh /root
RUN /root/detect-apt-proxy.sh ${APT_PROXY_PORT}

ENV UBUNTU_VERSION trusty
RUN echo "deb http://ppa.launchpad.net/libreoffice/ppa/ubuntu ${UBUNTU_VERSION} main" >> /etc/apt/sources.list
RUN echo "deb-src http://ppa.launchpad.net/libreoffice/ppa/ubuntu ${UBUNTU_VERSION} main" >> /etc/apt/sources.list
RUN apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv 36E81C9267FD1383FCC4490983FBA1751378B444
RUN \
    apt-get update && \
    DEBIAN_FRONTEND=noninteractive \
        apt-get install -y \
            unoconv \
            curl \
            make \
            git \
            asciidoc \
    && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/

# RUN curl -OL https://github.com/dagwieers/unoconv/archive/0.7.tar.gz
# RUN tar xzvf 0.7.tar.gz
RUN git clone https://github.com/dagwieers/unoconv.git
RUN cd unoconv && make && make install
ENV UNO_PATH /usr/lib/libreoffice/program/

RUN sed -i "/\#\!\/usr\/bin\/env python/c \#\!\/usr\/bin\/env python3" /usr/bin/unoconv
ENTRYPOINT [ "/usr/bin/unoconv" ]
CMD [ "--help" ]
