FROM gcr.io/distroless/static:nonroot

# OpenContainers Image Spec Labels
LABEL org.opencontainers.image.title="provider-rabbitmq"
LABEL org.opencontainers.image.description="Crossplane provider for RabbitMQ management"
LABEL org.opencontainers.image.vendor="rossigee"
LABEL org.opencontainers.image.licenses="Apache-2.0"
LABEL org.opencontainers.image.source="https://github.com/rossigee/provider-rabbitmq"
LABEL org.opencontainers.image.url="https://github.com/rossigee/provider-rabbitmq"
LABEL org.opencontainers.image.documentation="https://github.com/rossigee/provider-rabbitmq/blob/master/README.md"

ARG TARGETOS
ARG TARGETARCH

ADD bin/${TARGETOS}_${TARGETARCH}/provider /usr/local/bin/provider
ADD package.yaml /package.yaml
ADD crds /crds

USER 65532:65532
ENTRYPOINT ["/usr/local/bin/provider"]
