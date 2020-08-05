FROM golang:1.14-alpine as builder
ARG PKG_NAME=github.com/sjansen/terraform-fargate-ambassador

RUN apk --update add ca-certificates git
ADD go.mod go.sum /go/src/${PKG_NAME}/
RUN cd /go/src/${PKG_NAME} && \
    go mod download
ADD . /go/src/${PKG_NAME}
RUN cd /go/src/${PKG_NAME} && \
    CGO_ENABLED=0 GOOS=linux \
    go build \
        -a -installsuffix cgo \
        -ldflags="-s -w" \
        -o /main \
        ./cmd/ambassador/

FROM scratch
COPY --from=builder /main /app/
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
EXPOSE 8000
WORKDIR /app
ENTRYPOINT ["./main"]
CMD ["server"]
HEALTHCHECK --interval=1m --timeout=3s \
  CMD ["./main", "check-health"]
