ARG GO_VERSION=1.26.5
FROM golang:${GO_VERSION}-bookworm AS builder

WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/notify ./cmd/notify

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /out/notify /notify
USER nonroot:nonroot
ENTRYPOINT ["/notify"]
