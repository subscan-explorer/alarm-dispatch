FROM golang:1.19 as build

ENV CGO_ENABLED 0
ENV GOOS linux

WORKDIR /build/cache
ADD go.mod .
ADD go.sum .
RUN go mod download

WORKDIR /app/release

ADD . .
RUN go build -o alarm-dispatch cmd/main.go

FROM alpine as prod

RUN mkdir -p /app/bin/

COPY --from=build /app/release/alarm-dispatch /app/bin/alarm-dispatch

WORKDIR /app/

CMD ["bin/alarm-dispatch"]



