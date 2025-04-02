# Build stage
FROM golang:1.21.0-alpine as builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o server .

# Runtime stage
FROM alpine:3.18

WORKDIR /root/
RUN apk add --no-cache libc6-compat
COPY --from=builder /app/server .
COPY conf.yaml . 
EXPOSE 8080
CMD ["./server", "-config", "conf.yaml"]