FROM golang:1.17 AS build
WORKDIR /go/src/app
COPY go.mod go.sum /go/src/app/
RUN go mod download
COPY *.go /go/src/app/
RUN go build -o /go/bin/app

FROM gcr.io/distroless/base-debian11
COPY --from=build /go/bin/app /
CMD ["/app"]
