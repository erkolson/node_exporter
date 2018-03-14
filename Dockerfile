FROM quay.io/prometheus/busybox:glibc
# FROM jpetazzo/nsenter

COPY node_exporter /bin/node_exporter
COPY --from=jpetazzo/nsenter /nsenter /usr/local/bin/nsenter

# EXPOSE      9101
EXPOSE      9100

ENTRYPOINT  [ "/bin/node_exporter" ]
