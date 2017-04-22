FROM        jpetazzo/nsenter

COPY node_exporter /bin/node_exporter

EXPOSE      9100
ENTRYPOINT  [ "/bin/node_exporter" ]
