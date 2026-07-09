FROM gcr.io/distroless/static:nonroot

ARG TARGETOS=linux
ARG TARGETARCH=amd64

LABEL org.opencontainers.image.title="provider-rabbitmq"
LABEL org.opencontainers.image.description="Crossplane provider for RabbitMQ management"
LABEL org.opencontainers.image.vendor="rossigee"
LABEL org.opencontainers.image.licenses="Apache-2.0"
LABEL org.opencontainers.image.source="https://github.com/rossigee/provider-rabbitmq"

COPY --chmod=0755 bin/${TARGETOS}_${TARGETARCH}/provider /usr/local/bin/provider

USER 65532:65532
ENTRYPOINT ["/usr/local/bin/provider"]