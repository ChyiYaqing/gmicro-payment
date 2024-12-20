FROM golang:1.22 AS builder
WORKDIR /usr/src/app
ENV GOPROXY=https://goproxy.io,direct
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o payment ./cmd/main.go

FROM scratch
COPY --from=builder /usr/src/app/payment ./order
CMD ["./payment"]