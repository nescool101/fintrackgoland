# Guía de Despliegue en Fly.io

Esta guía te ayudará a desplegar la aplicación **fintrackgoland** en Fly.io.

## Prerrequisitos

1. **Instalar flyctl**:
   ```bash
   # macOS
   brew install flyctl
   
   # Linux/WSL
   curl -L https://fly.io/install.sh | sh
   
   # Windows
   # Descargar desde https://fly.io/docs/hands-on/install-flyctl/
   ```

2. **Crear cuenta y autenticarse**:
   ```bash
   flyctl auth signup  # o flyctl auth login si ya tienes cuenta
   ```

## Configuración Inicial

1. **Crear la aplicación en Fly.io** (solo la primera vez):
   ```bash
   flyctl apps create fintrackgoland
   ```

2. **Configurar variables de entorno**:
   
   Edita el archivo `deploy-fly.sh` y reemplaza los valores:
   - `FMP_API_KEY`: Tu clave de API de Financial Modeling Prep
   - `EMAIL_PASS`: Tu contraseña de aplicación de Gmail
   
   Luego ejecuta:
   ```bash
   ./deploy-fly.sh
   ```

## Despliegue Manual

Si prefieres configurar manualmente:

1. **Configurar secretos**:
   ```bash
   # Autenticación
   flyctl secrets set AUTH_USERNAME="nescao3" -a fintrackgoland
   flyctl secrets set AUTH_PASSWORD="fintrack2024" -a fintrackgoland
   
   # APIs
   flyctl secrets set FMP_API_KEY="tu_fmp_api_key" -a fintrackgoland
   flyctl secrets set ALPHA_VANTAGE_API_KEY="yPZHRYXFUZL3STKAV" -a fintrackgoland
   
   # Email
   flyctl secrets set EMAIL_HOST="smtp.gmail.com" -a fintrackgoland
   flyctl secrets set EMAIL_PORT="587" -a fintrackgoland
   flyctl secrets set EMAIL_USER="nescool101@gmail.com" -a fintrackgoland
   flyctl secrets set EMAIL_PASS="tu_password_gmail" -a fintrackgoland
   flyctl secrets set RECIPIENT="nescool101@gmail.com,paulocesarcelis@gmail.com" -a fintrackgoland
   
   # Configuración
   flyctl secrets set RUN_CRON="true" -a fintrackgoland
   flyctl secrets set USE_BACKUP="false" -a fintrackgoland
   ```

2. **Desplegar**:
   ```bash
   flyctl deploy -a fintrackgoland
   ```

## Verificación

1. **Verificar estado**:
   ```bash
   flyctl status -a fintrackgoland
   ```

2. **Ver logs**:
   ```bash
   flyctl logs -a fintrackgoland
   ```

3. **Probar endpoints**:
   ```bash
   # Health check
   curl https://fintrackgoland.fly.dev/health
   
   # API status
   curl https://fintrackgoland.fly.dev/api/status
   
   # Endpoint protegido
   curl -u nescao3:fintrack2024 https://fintrackgoland.fly.dev/api/indices
   ```

## Comandos Útiles

- **Escalar aplicación**: `flyctl scale count 1 -a fintrackgoland`
- **Ver métricas**: `flyctl dashboard -a fintrackgoland`
- **SSH a la máquina**: `flyctl ssh console -a fintrackgoland`
- **Reiniciar**: `flyctl restart -a fintrackgoland`
- **Destruir app**: `flyctl apps destroy fintrackgoland`

## Solución de Problemas

### Error: "app appears to be crashing"

1. Verificar logs: `flyctl logs -a fintrackgoland`
2. Verificar que todas las variables de entorno estén configuradas
3. Verificar que la aplicación escuche en `0.0.0.0:8080`

### Error: "smoke checks failed"

1. Verificar que el endpoint `/health` responda correctamente
2. Verificar que el puerto 8080 esté expuesto
3. Verificar configuración en `fly.toml`

### Variables de entorno faltantes

```bash
# Ver variables configuradas
flyctl secrets list -a fintrackgoland

# Configurar variable faltante
flyctl secrets set VARIABLE_NAME="valor" -a fintrackgoland
```

## URLs de la Aplicación

- **Aplicación**: https://fintrackgoland.fly.dev
- **Health Check**: https://fintrackgoland.fly.dev/health
- **API Status**: https://fintrackgoland.fly.dev/api/status
- **Dashboard Fly.io**: https://fly.io/apps/fintrackgoland 