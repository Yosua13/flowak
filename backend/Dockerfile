# syntax=docker/dockerfile:1

# Build the Vite frontend served by the Go application at runtime.
FROM node:20-alpine AS frontend-builder
WORKDIR /frontend

COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci

COPY frontend/ ./
RUN npm run build

# Compile a static Linux binary for the production Go runtime.
FROM golang:1.25-alpine AS backend-builder
WORKDIR /src/backend

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/flowak .

FROM alpine:3.20 AS runtime
RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app/backend
COPY --from=backend-builder /out/flowak ./flowak
COPY --from=backend-builder /src/backend/db/migrations ./db/migrations
# main.go resolves frontend assets as ../frontend/dist from /app/backend.
COPY --from=frontend-builder /frontend/dist /app/frontend/dist

RUN chown -R appuser:appgroup /app
USER appuser

EXPOSE 3001
ENV PORT=3001
CMD ["./flowak"]
