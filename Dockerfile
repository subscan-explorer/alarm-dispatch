FROM golang:1.19 as build

ENV CGO_ENABLED 0
ENV GOOS linux
ENV GOARCH=amd64

WORKDIR /build/cache
ADD go.mod .
ADD go.sum .
RUN go mod download

WORKDIR /workspace/release

ADD . .
RUN go build -o alarm-dispatch cmd/main.go

FROM alpine as prod

RUN mkdir -p /workspace/bin/

COPY --from=build /workspace/release/alarm-dispatch /workspace/bin/alarm-dispatch

WORKDIR /workspace/

CMD ["/workspace/bin/alarm-dispatch"]



