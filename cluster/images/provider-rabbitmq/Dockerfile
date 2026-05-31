FROM golang:1.26.3 AS builder

ARG TARGETOS=linux
ARG TARGETARCH=amd64

WORKDIR /workspace
COPY go.mod go.sum ./
RUN go mod download

COPY apis/    apis/
COPY cmd/     cmd/
COPY internal/ internal/

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -o /provider ./cmd/provider/

FROM gcr.io/distroless/static:nonroot

LABEL org.opencontainers.image.title="provider-rabbitmq"
LABEL org.opencontainers.image.description="Crossplane provider for RabbitMQ management"
LABEL org.opencontainers.image.vendor="rossigee"
LABEL org.opencontainers.image.licenses="Apache-2.0"
LABEL org.opencontainers.image.source="https://github.com/rossigee/provider-rabbitmq"
LABEL org.opencontainers.image.url="https://github.com/rossigee/provider-rabbitmq"
LABEL org.opencontainers.image.documentation="https://github.com/rossigee/provider-rabbitmq/blob/master/README.md"

COPY --from=builder /provider /usr/local/bin/provider
COPY package/crossplane.yaml /package.yaml
COPY package/crds /crds

USER 65532:65532
ENTRYPOINT ["/usr/local/bin/provider"]
