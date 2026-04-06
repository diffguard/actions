FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /active-gate ./actions/active-gate/

FROM alpine:3.19
RUN apk add --no-cache ca-certificates git
COPY --from=builder /active-gate /usr/local/bin/active-gate
ENTRYPOINT ["active-gate"]
