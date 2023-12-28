FROM scratch

LABEL org.opencontainers.image.source=https://github.com/gomicro/forge
LABEL org.opencontainers.image.authors="dev@gomicro.io"

ADD forge forge

CMD ["/forge"]
