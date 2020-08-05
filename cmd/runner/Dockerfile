FROM golang:1.14-alpine as builder
ARG PKG_NAME=github.com/sjansen/terraform-fargate-ambassador

RUN apk --no-cache add git
WORKDIR /go/src/${PKG_NAME}/
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux \
    go build \
        -a -installsuffix cgo \
        -ldflags="-s -w" \
        -o /main \
        ./cmd/runner/

FROM alpine:latest
RUN apk --no-cache add ca-certificates wget
COPY --from=builder /main /app/
EXPOSE 80
CMD ["/app/main"]
HEALTHCHECK --interval=1m --timeout=3s \
  CMD wget 127.0.0.1:80/availability -q -O -
