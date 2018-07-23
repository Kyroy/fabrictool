FROM golang:1.10 AS builder
RUN go get -u golang.org/x/vgo
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 vgo build \
    -ldflags "-X github.com/kyroy/fabrictool/cmd.version=0.0.1 -X github.com/kyroy/fabrictool/cmd.gitCommit=$(git rev-parse HEAD)" \
    -o fabrictool

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /app/fabrictool ./fabrictool
ENTRYPOINT ["./fabrictool"]
