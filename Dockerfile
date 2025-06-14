# Build stage
FROM golang:1.22-alpine AS builder

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
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s" -o main .

# Production stage
FROM alpine:latest

# Instalar ca-certificates para HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Crear usuario no-root
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Crear directorio de trabajo para el usuario
RUN mkdir -p /app && chown appuser:appgroup /app

# Establecer directorio de trabajo
WORKDIR /app

# Copiar el binario desde el stage de build
COPY --from=builder /app/main ./main

# Hacer el binario ejecutable y cambiar propietario
RUN chmod +x ./main && chown appuser:appgroup ./main

# Cambiar a usuario no-root
USER appuser

# Exponer puerto
EXPOSE 8080

# Comando para ejecutar la aplicación
CMD ["./main"] 