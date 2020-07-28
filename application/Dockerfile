FROM golang:1.14-alpine as builder
ARG PKG_NAME=github.com/sjansen/terraform-fargate-ambassador/application

RUN apk --update add ca-certificates git
ADD go.mod go.sum /go/src/${PKG_NAME}/
RUN cd /go/src/${PKG_NAME} && \
    go mod download && \
    mkdir /build
ADD . /go/src/${PKG_NAME}
RUN cd /go/src/${PKG_NAME} && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -a -installsuffix cgo -o /main

FROM scratch
COPY --from=builder /main /app/
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
WORKDIR /app
ENTRYPOINT ["./main"]
EXPOSE 8080
