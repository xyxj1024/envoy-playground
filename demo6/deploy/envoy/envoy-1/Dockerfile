FROM envoyproxy/envoy:dev

ENV ENVOY_UID=0
EXPOSE 8001
EXPOSE 9001
EXPOSE 10001

#ARG CONTROL_PLANE_HOST
COPY config.yaml /etc/envoy/envoy_1.yaml
#RUN sed -i "s/host.docker.internal/$CONTROL_PLANE_HOST/g" /etc/envoy/envoy_1.yaml

CMD ["/usr/local/bin/envoy", \
     "-c", "/etc/envoy/envoy_1.yaml", \
     "-l", "debug"]