# [build image]
FROM golang:1.15-alpine as builder

RUN apk update && apk upgrade && \
    apk add --no-cache bash git openssh

# gco is needed to compile echo or one of echo dependencies
ENV CGO_ENABLED=1

# therefore a c compiler is needed on build image
RUN apk add --no-cache gcc libc-dev

WORKDIR /app

# build application
COPY . .
RUN go build -v cmd/api/api.go


# [runtime image]
FROM alpine:latest
COPY --from=builder /app/api /app/api

WORKDIR /app/
CMD ["./api"]
