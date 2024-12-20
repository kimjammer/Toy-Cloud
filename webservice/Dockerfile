FROM golang:1.23.4 AS build-stage
LABEL authors="kimjammer"

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /webservice

FROM gcr.io/distroless/base-debian12 AS build-release-stage

WORKDIR /

COPY --from=build-stage /webservice /webservice

EXPOSE 8080

USER nonroot:nonroot

CMD ["/webservice"]