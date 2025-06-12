# 🚀 API Guide - Sistema de Datos Financieros

## 📊 API Utilizada

### Financial Modeling Prep (FMP) - API Única
- **Llamadas gratuitas**: 250 por día
- **Soporta índices**: ✅ Sí (SPX, NDX, DJI, NYA, ES_F, NQ_F)
- **Registrarse**: https://financialmodelingprep.com/developer
- **Estado**: **ÚNICA** (proveedor exclusivo)

## 🔧 Configuración

### Variables de Entorno Requeridas

```bash
# API Key
FMP_API_KEY=tu_clave_fmp_aqui          # ÚNICA - 250 llamadas gratis

# Base de datos
DATABASE_URL=postgresql://user:password@localhost/fintrack

# Email
EMAIL_HOST=smtp.gmail.com
EMAIL_PORT=587
EMAIL_USER=tu_email@gmail.com
EMAIL_PASS=tu_password_de_aplicacion
EMAIL_RECIPIENT=destinatario@ejemplo.com

# Autenticación
AUTH_USERNAME=admin
AUTH_PASSWORD=tu_password_seguro

# Configuración
RUN_CRON=true
GIN_MODE=release
```

## 🌐 Endpoints REST API

### Rutas Públicas (Sin autenticación)

#### 1. Health Check
```http
GET /health
```
**Respuesta:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:00:00Z",
  "version": "1.0.0"
}
```

#### 2. Estado de API
```http
GET /api/status
```
**Respuesta:**
```json
{
  "timestamp": "2024-01-15T10:00:00Z",
  "api": {
    "name": "Financial Modeling Prep",
    "free_calls_per_day": 250,
    "supports_indices": true,
    "status": "active",
    "website": "https://financialmodelingprep.com"
  },
  "supported_indices": ["SPX", "NDX", "DJI", "NYA", "ES_F", "NQ_F"],
  "total_symbols": 6
}
```

#### 3. Índices Soportados
```http
GET /api/indices
```
**Respuesta:**
```json
{
  "target_indices": ["SPX", "NDX", "DJI", "NYA", "ES_F", "NQ_F"],
  "all_symbols": ["SPX", "NDX", "DJI", "NYA", "ES_F", "NQ_F"],
  "total_indices": 6,
  "total_symbols": 6,
  "api_provider": "FMP (Financial Modeling Prep)",
  "daily_free_calls": 250
}
```

### Rutas Protegidas (Requieren autenticación básica)

#### 4. Datos de un Símbolo
```http
GET /api/stock/:symbol?date=YYYY-MM-DD
```

**Ejemplos:**
```bash
# Datos de SPX para hoy
curl -u admin:password http://localhost:8080/api/stock/SPX

# Datos de NDX para fecha específica
curl -u admin:password http://localhost:8080/api/stock/NDX?date=2024-01-15
```

**Respuesta:**
```json
{
  "symbol": "SPX",
  "date": "2024-01-15",
  "success": true,
  "data": {
    "status": "OK",
    "symbol": "SPX",
    "date": "2024-01-15",
    "open": 4700.5,
    "high": 4750.2,
    "low": 4690.1,
    "close": 4740.8,
    "volume": 1500000,
    "afterHours": 0,
    "preMarket": 0
  }
}
```

#### 5. Datos de Múltiples Símbolos
```http
GET /api/stocks?symbols=SPX,NDX,DJI&date=YYYY-MM-DD
```

**Ejemplo:**
```bash
curl -u admin:password "http://localhost:8080/api/stocks?symbols=SPX,NDX,DJI&date=2024-01-15"
```

#### 6. Datos Semanales
```http
GET /api/weekly?symbols=SPX,NDX (opcional)
```

**Ejemplo:**
```bash
# Todos los índices para la semana actual
curl -u admin:password http://localhost:8080/api/weekly

# Índices específicos para la semana actual
curl -u admin:password "http://localhost:8080/api/weekly?symbols=SPX,NDX"
```

#### 7. Enviar Reporte Semanal
```http
POST /api/report/send
```

**Ejemplo:**
```bash
curl -X POST -u admin:password http://localhost:8080/api/report/send
```

## 🎯 Índices Soportados

| Símbolo | Nombre | Mapeo FMP |
|---------|--------|-----------|
| SPX | S&P 500 Index | ^GSPC |
| NDX | NASDAQ Composite | ^IXIC |
| DJI | Dow Jones Industrial Average | ^DJI |
| NYA | NYSE Composite Index | ^NYA |
| ES_F | E-mini S&P 500 Futures | ES=F |
| NQ_F | E-mini NASDAQ 100 Futures | NQ=F |

## 🚀 Ejecutar el Servidor

```bash
# Compilar
go build -o fintrack

# Ejecutar
./fintrack
```

**Salida esperada:**
```
Usando FMP API como proveedor único (250 llamadas gratis/día)
🎯 Índices objetivo configurados: [SPX NDX DJI NYA ES_F NQ_F]
📊 Total de símbolos a procesar: 6
🔑 API única: FMP (250 llamadas gratis/día)
🚀 Iniciando servidor REST API en puerto :8080
📋 Endpoints disponibles:
   GET  /health                    - Estado del servicio
   GET  /api/status                - Información de API
   GET  /api/indices               - Índices soportados
   GET  /api/stock/:symbol         - Datos de un símbolo
   GET  /api/stocks                - Datos de múltiples símbolos
   GET  /api/weekly                - Datos semanales
   POST /api/report/send           - Enviar reporte semanal
⏰ Cron configurado: cada viernes a las 9:00 AM
```

## 🔒 Autenticación

Todas las rutas `/api/*` (excepto `/api/status` y `/api/indices`) requieren autenticación básica HTTP:

```bash
curl -u username:password http://localhost:8080/api/endpoint
```

## 📈 Límites y Consideraciones

- **FMP API**: 250 llamadas gratis/día
- **Concurrencia**: Máximo 5 solicitudes simultáneas
- **Cron**: Ejecuta cada viernes a las 9:00 AM
- **Datos**: Lunes a viernes (días de mercado)
- **Failover**: No necesario (un solo proveedor)

## 🛠️ Desarrollo

### Estructura de Archivos
```
.
├── api/
│   └── handlers.go          # Manejadores REST con Gin
├── config/
│   └── config.go           # Configuración de variables
├── service/
│   └── fmp_service.go      # Servicio FMP único
├── models/
│   └── data.go            # Estructuras de datos
├── utils/
│   └── (utilidades)
└── main.go                # Servidor principal
```

### Agregar Nuevos Endpoints

1. Agregar método en `api/handlers.go`
2. Registrar ruta en `main.go`
3. Probar con curl o Postman

### Optimización

Con un solo proveedor de API, el sistema es más simple y directo:
- Sin lógica de failover
- Sin switching entre APIs
- Configuración simplificada
- Mejor rendimiento (menos overhead) 