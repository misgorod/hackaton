# build stage
FROM golang:alpine AS build
COPY go.mod /build/
WORKDIR /build
RUN apk add git && \
    go mod download
COPY . /build/
RUN go build -o main

# final stage
FROM alpine
WORKDIR /app
COPY --from=build /build/main /app/
EXPOSE 80
ENTRYPOINT ./main