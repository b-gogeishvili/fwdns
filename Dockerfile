FROM golang:1.26 AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 make build

FROM busybox
COPY --from=build /src/fwdns /fwdns

EXPOSE 53/udp 8080

ENTRYPOINT ["/fwdns", "-dns", ":53"]
