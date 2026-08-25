FROM golang:1.25 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/dual-teacher ./cmd/server
FROM gcr.io/distroless/static-debian12
WORKDIR /app
COPY --from=build /out/dual-teacher /app/dual-teacher
COPY --from=build /src/internal/storage/sqlite/migration.sql /app/internal/storage/sqlite/migration.sql
EXPOSE 8080
ENTRYPOINT ["/app/dual-teacher"]
