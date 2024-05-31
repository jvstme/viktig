FROM golang:1.22 as build
WORKDIR /app
COPY ./go.mod ./go.sum ./
RUN go mod download
COPY . .
# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o app ./cmd/app

FROM scratch as app
COPY --from=build /app/app ./app
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
EXPOSE $APP_PORT