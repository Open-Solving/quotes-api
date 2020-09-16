# [build image]
FROM golang:1.15-alpine as builder

RUN apk update && apk upgrade && \
    apk add --no-cache bash git openssh

# therefore a c compiler is needed on build image
RUN apk add --no-cache gcc libc-dev

WORKDIR /app

# fetch dependencies
COPY go.mod .
COPY go.sum .
RUN go mod download

# build application
COPY . .
RUN CGO_ENABLED=1 go build -v cmd/api/api.go


# [runtime image]
FROM alpine:latest
COPY --from=builder /app/api /app/api

WORKDIR /app/
CMD ["./api"]
