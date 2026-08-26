# syntax=docker/dockerfile:1

FROM node:22-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

FROM golang:1.22-alpine AS builder
WORKDIR /app/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
COPY --from=frontend-builder /app/frontend/dist /app/frontend/dist
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/api ./cmd/api

FROM alpine:3.20
WORKDIR /app
RUN addgroup -S app && adduser -S app -G app
COPY --from=builder /app/api /app/api
COPY --from=builder /app/backend/internal/db/migrations /app/backend/internal/db/migrations
COPY --from=builder /app/frontend/dist /app/frontend/dist
RUN mkdir -p /app/data/uploads && chown -R app:app /app
EXPOSE 8080
USER app
CMD ["/app/api"]
