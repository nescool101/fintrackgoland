# Build stage
FROM golang:1.21-alpine AS builder

# Instalar dependencias necesarias
RUN apk add --no-cache git ca-certificates tzdata

# Establecer directorio de trabajo
WORKDIR /app

# Copiar archivos de dependencias
COPY go.mod go.sum ./

# Descargar dependencias
RUN go mod download

# Copiar código fuente
COPY . .

# Compilar la aplicación
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Production stage
FROM alpine:latest

# Instalar ca-certificates para HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Crear usuario no-root
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Establecer directorio de trabajo
WORKDIR /root/

# Copiar el binario desde el stage de build
COPY --from=builder /app/main .

# Cambiar propietario del archivo
RUN chown appuser:appgroup main

# Cambiar a usuario no-root
USER appuser

# Exponer puerto
EXPOSE 8080

# Comando para ejecutar la aplicación
CMD ["./main"] 