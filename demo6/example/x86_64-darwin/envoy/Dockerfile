FROM envoyproxy/envoy:dev

ENV ENVOY_UID=0
EXPOSE 80
EXPOSE 443

#ARG CONTROL_PLANE_HOST
COPY config.yaml /etc/envoy/envoy.yaml
#RUN sed -i "s/host.docker.internal/$CONTROL_PLANE_HOST/g" /etc/envoy/envoy.yaml

CMD ["/usr/local/bin/envoy", \
     "-c", "/etc/envoy/envoy.yaml", \
     "-l", "debug"]