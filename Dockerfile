FROM golang:1.21-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app ./cmd/backend

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=build /app /app
WORKDIR /
ENV PORT=8080
EXPOSE 8080
CMD ["/app"]
