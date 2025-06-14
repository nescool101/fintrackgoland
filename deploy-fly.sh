#!/bin/bash

# Script para desplegar en Fly.io
echo "🚀 Desplegando fintrackgoland en Fly.io..."

# Verificar que flyctl esté instalado
if ! command -v flyctl &> /dev/null; then
    echo "❌ flyctl no está instalado. Instálalo desde https://fly.io/docs/hands-on/install-flyctl/"
    exit 1
fi

# Verificar que estemos logueados en Fly.io
if ! flyctl auth whoami &> /dev/null; then
    echo "❌ No estás logueado en Fly.io. Ejecuta: flyctl auth login"
    exit 1
fi

# Configurar variables de entorno (reemplaza con tus valores reales)
echo "🔧 Configurando variables de entorno..."

# Variables de autenticación básica
flyctl secrets set AUTH_USERNAME="nescao3" -a fintrackgoland
flyctl secrets set AUTH_PASSWORD="fintrack2024" -a fintrackgoland

# API Keys
flyctl secrets set FMP_API_KEY="tu_fmp_api_key_aqui" -a fintrackgoland
flyctl secrets set ALPHA_VANTAGE_API_KEY="yPZHRYXFUZL3STKAV" -a fintrackgoland

# Configuración de email
flyctl secrets set EMAIL_HOST="smtp.gmail.com" -a fintrackgoland
flyctl secrets set EMAIL_PORT="587" -a fintrackgoland
flyctl secrets set EMAIL_USER="nescool101@gmail.com" -a fintrackgoland
flyctl secrets set EMAIL_PASS="tu_email_password_aqui" -a fintrackgoland
flyctl secrets set RECIPIENT="nescool101@gmail.com,paulocesarcelis@gmail.com" -a fintrackgoland

# Configuración adicional
flyctl secrets set RUN_CRON="true" -a fintrackgoland
flyctl secrets set USE_BACKUP="false" -a fintrackgoland
flyctl secrets set GIN_MODE="release" -a fintrackgoland

echo "✅ Variables de entorno configuradas"

# Desplegar la aplicación
echo "🚀 Desplegando aplicación..."
flyctl deploy -a fintrackgoland

echo "✅ Despliegue completado!"
echo "🌐 Tu aplicación estará disponible en: https://fintrackgoland.fly.dev"
echo "🔍 Para ver logs: flyctl logs -a fintrackgoland"
echo "📊 Para ver estado: flyctl status -a fintrackgoland" 