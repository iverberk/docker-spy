FROM scratch

ADD ./docker-spy /bin/docker-spy

ENV DNS_RECURSOR 8.8.8.8

EXPOSE 53
EXPOSE 53/udp

ENTRYPOINT ["/bin/docker-spy"]
