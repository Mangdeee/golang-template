FROM golang:1.19-alpine as prod
WORKDIR /root/

COPY . .

ENV GOOS="linux"
ENV CGO_ENABLED=0

RUN apk update \
    && apk add --no-cache \
    ca-certificates \
    curl \
    tzdata \
    git \
    && update-ca-certificates

RUN go mod download

RUN cd cmd/app/ && go build -mod=readonly -v -o /root/main

EXPOSE 8000

CMD ["./main"]

