FROM golang:1.23.4 AS build-stage
LABEL authors="kimjammer"

WORKDIR /app/service
RUN mkdir /app/common

#Copy go.mod and go.sum
COPY ./common/go.mod /app/common/
COPY ./loadbalancer/go.mod ./loadbalancer/go.sum ./
#Install modules
RUN go mod download
#Copy code
COPY ./common/*.go /app/common/
COPY ./loadbalancer/*.go ./
#Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /loadbalancer

FROM gcr.io/distroless/base-debian12 AS build-release-stage

WORKDIR /

COPY --from=build-stage /loadbalancer /loadbalancer

EXPOSE 8080

USER nonroot:nonroot

CMD ["/loadbalancer"]