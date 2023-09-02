FROM golang:1.21 AS build
WORKDIR /go/src
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 go build

FROM gcr.io/distroless/static-debian12
WORKDIR /app/
COPY --from=build /go/src/household-accounts-subscriptions ./
CMD ["/app/household-accounts-subscriptions"]
