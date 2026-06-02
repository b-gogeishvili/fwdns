FROM golang:1.26 AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN make build

FROM busybox
COPY --from=build /fwdns /fwdns

EXPOSE 53/udp 8080

ENTRYPOINT ["/fwdns -dns :53"]
