FROM golang:1.12.6 as builder
RUN go version

RUN mkdir /build
WORKDIR /build
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN make build

FROM alpine:latest
COPY --from=builder /build/dist /go/bin/build
EXPOSE 8081
WORKDIR /go/bin/build/
CMD ["./server"]


