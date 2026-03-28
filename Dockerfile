# --- Frontend build stage ---
FROM node:24-alpine AS frontend-builder

WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# --- Backend build stage ---
FROM golang:1.26-alpine AS backend-builder

WORKDIR /app/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

# --- Final stage ---
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app
COPY --from=backend-builder /server /app/server
COPY --from=frontend-builder /app/frontend/dist /app/frontend/dist

RUN mkdir -p /app/uploads

EXPOSE 8080

ENV PORT=8080
ENV BLOB_STORAGE_PATH=/app/uploads

CMD ["/app/server"]
