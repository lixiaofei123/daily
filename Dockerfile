FROM golang:latest AS builder

WORKDIR /app

COPY app/ app/
COPY configs/ configs/
COPY main.go main.go
COPY config_hanlder.go config_hanlder.go
COPY go.mod go.mod
COPY go.sum go.sum

RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/main .

COPY public public
COPY static static

EXPOSE 8081
RUN date +"%Y%m%d%H%M%S" > buildtime

ENTRYPOINT ["./main"]