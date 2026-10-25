FROM --platform=$BUILDPLATFORM golang:1.25.7-alpine@sha256:f6751d823c26342f9506c03797d2527668d095b0a15f1862cddb4d927a7a4ced AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -mod=readonly -trimpath -ldflags="-s -w" \
    -o /out/codex-oauth-proxy ./cmd/codex-oauth-proxy

FROM gcr.io/distroless/static-debian13:nonroot@sha256:18f4cc26c9ffa7f20a359298be7deadac5c607c8e03a02b8119f1a5fc2f82c99

WORKDIR /
COPY --from=builder /out/codex-oauth-proxy /usr/local/bin/codex-oauth-proxy
ENV ENV=production \
    PORT=9879 \
    XDG_CONFIG_HOME=/config
USER 65532:65532
EXPOSE 9879
ENTRYPOINT ["/usr/local/bin/codex-oauth-proxy"]
CMD ["--creds-store=xdg"]
