FROM golang:1.23.4 AS build-stage
LABEL authors="kimjammer"

WORKDIR /app/service
RUN mkdir /app/common

#Copy go.mod and go.sum
COPY ./common/go.mod /app/common/
COPY ./orchestrator/go.mod ./orchestrator/go.sum ./
#Install modules
RUN go mod download
#Copy code
COPY ./common/*.go /app/common/
COPY ./orchestrator/*.go ./
#Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /orchestrator

FROM docker:latest AS build-release-stage

WORKDIR /

COPY --from=build-stage /orchestrator /orchestrator

EXPOSE 8080

CMD ["/orchestrator"]