# FinTrack GoLand - Servicio Mejorado de Datos Financieros

## Descripción

FinTrack GoLand es un servicio mejorado para obtener datos financieros que ahora admite múltiples proveedores de API, incluyendo soporte específico para los índices solicitados: SPX, NDX, DJI, NYA, ES_F, y NQ_F.

## Nuevas Características

### Múltiples Proveedores de API
- **Polygon.io** (proveedor primario)
- **Financial Modeling Prep (FMP)** (proveedor de respaldo)
- Sistema de failover automático entre proveedores

### Índices Soportados
- **SPX** - S&P 500 Index
- **NDX** - NASDAQ-100 Index  
- **DJI** - Dow Jones Industrial Average
- **NYA** - NYSE Composite Index
- **ES_F** - E-mini S&P 500 Futures
- **NQ_F** - E-mini NASDAQ 100 Futures

### APIs Gratuitas Disponibles

#### 1. Financial Modeling Prep (FMP)
- **Plan gratuito**: 250 llamadas/día
- **Datos disponibles**: Acciones, índices, forex, cripto
- **Características**: 
  - Datos históricos de fin de día
  - Datos fundamentales
  - Cobertura global limitada en plan gratuito
- **Registro**: https://financialmodelingprep.com/developer/docs

#### 2. EOD Historical Data (Alternativa)
- **Plan gratuito**: 20 llamadas/día
- **Datos disponibles**: Más de 150,000 instrumentos
- **Registro**: https://eodhd.com/

## Configuración

### Variables de Entorno

```bash
# API Keys
API_KEY=tu_polygon_api_key              # Clave de Polygon.io (primaria)
FMP_API_KEY=tu_fmp_api_key             # Clave de Financial Modeling Prep

# Configuración del sistema
USE_BACKUP=true                         # Usar proveedor de respaldo (FMP)
RUN_CRON=true                          # Ejecutar trabajos programados
DATABASE_URL=tu_database_url

# Configuración de email
EMAIL_HOST=smtp.gmail.com
EMAIL_PORT=587
EMAIL_USER=tu_email@gmail.com
EMAIL_PASS=tu_password_de_app
EMAIL_RECIPIENT=destinatario@email.com

# Autenticación HTTP
AUTH_USERNAME=admin
AUTH_PASSWORD=tu_password_seguro
```

### Ejemplo de archivo .env

```env
API_KEY=YOUR_POLYGON_API_KEY
FMP_API_KEY=YOUR_FMP_API_KEY
USE_BACKUP=true
RUN_CRON=true
DATABASE_URL=postgresql://user:pass@localhost/dbname
EMAIL_HOST=smtp.gmail.com
EMAIL_PORT=587
EMAIL_USER=your_email@gmail.com
EMAIL_PASS=your_app_password
EMAIL_RECIPIENT=recipient@email.com
AUTH_USERNAME=admin
AUTH_PASSWORD=secure_password123
```

## Instalación y Uso

### 1. Clonar el repositorio
```bash
git clone https://github.com/nescool101/fintrackgoland.git
cd fintrackgoland
```

### 2. Instalar dependencias
```bash
go mod download
```

### 3. Configurar variables de entorno
```bash
cp .env.example .env
# Editar .env con tus claves de API
```

### 4. Ejecutar la aplicación
```bash
go run main.go
```

### 5. Usar librerías externas (opcional)

Para usar la librería `fmpcloud-go`:
```bash
go get -u github.com/spacecodewor/fmpcloud-go
```

Luego descomentar las líneas relevantes en `service/fmp_cloud_service.go`.

## Endpoints HTTP

### Health Check
```
GET /health
```

### Test Manual
```
GET /test
Authorization: Basic (admin:password)
```

## Arquitectura del Sistema

### Servicios Disponibles

1. **DataService** (`service/data_service.go`)
   - Servicio original usando Polygon API

2. **FMPService** (`service/fmp_service.go`)
   - Nuevo servicio usando Financial Modeling Prep API
   - Implementación directa de la API REST

3. **FMPCloudService** (`service/fmp_cloud_service.go`)
   - Servicio preparado para usar la librería `fmpcloud-go`
   - Requiere instalación de la librería externa

4. **UnifiedService** (`service/unified_service.go`)
   - Servicio unificado que combina múltiples proveedores
   - Failover automático entre APIs
   - Interfaz común para todos los proveedores

### Interfaz DataProvider

```go
type DataProvider interface {
    FetchData(symbol, date string)
    FetchWeeklyData(symbols []string, dates []string)
    GetResults() []models.StockData
    GetFailed() []string
    ClearResults()
}
```

## Mapeo de Símbolos

El sistema convierte automáticamente los símbolos a los formatos apropiados para cada API:

| Símbolo Interno | Polygon | FMP     | Descripción |
|----------------|---------|---------|-------------|
| SPX            | SPX     | ^GSPC   | S&P 500 Index |
| NDX            | NDX     | ^IXIC   | NASDAQ Composite |
| DJI            | DJI     | ^DJI    | Dow Jones Industrial |
| NYA            | NYA     | ^NYA    | NYSE Composite |
| ES_F           | ES=F    | ES=F    | E-mini S&P 500 Futures |
| NQ_F           | NQ=F    | NQ=F    | E-mini NASDAQ 100 Futures |

## Programación (Cron)

El sistema ejecuta automáticamente la recopilación de datos:
- **Frecuencia**: Viernes a las 9:00 AM
- **Formato**: `0 0 9 ? * FRI`

## Manejo de Errores y Failover

1. **Proveedor Primario**: Polygon API
2. **Proveedor de Respaldo**: FMP API (si `USE_BACKUP=true`)
3. **Límites de Rate**: 5 solicitudes concurrentes por proveedor
4. **Timeout**: 10 segundos por solicitud

## Consideraciones de Planes Gratuitos

### Financial Modeling Prep
- ✅ 250 llamadas/día gratis
- ✅ Datos históricos EOD
- ✅ Soporte para índices
- ❌ Sin datos intraday en plan gratuito
- ❌ Sin datos pre/post market

### Polygon.io
- ❌ Plan gratuito muy limitado
- ✅ Excelente para planes pagos
- ✅ Datos en tiempo real

## Desarrollo y Extensión

### Agregar Nuevo Proveedor de API

1. Crear nuevo archivo en `service/`
2. Implementar interfaz `DataProvider`
3. Agregar al `UnifiedService` si es necesario

### Ejemplo de implementación:
```go
type NewAPIService struct {
    APIKey  string
    Results []models.StockData
    Failed  []string
    Mutex   sync.Mutex
}

func (nas *NewAPIService) FetchData(symbol, date string) {
    // Implementar lógica de API
}

func (nas *NewAPIService) GetResults() []models.StockData {
    nas.Mutex.Lock()
    defer nas.Mutex.Unlock()
    return nas.Results
}

// ... otros métodos requeridos por DataProvider
```

## Logs y Monitoreo

El sistema genera logs detallados sobre:
- Símbolos procesados
- APIs utilizadas
- Errores y failovers
- Rendimiento de requests

## Contribución

1. Fork el proyecto
2. Crear una rama para tu feature (`git checkout -b feature/AmazingFeature`)
3. Commit tus cambios (`git commit -m 'Add some AmazingFeature'`)
4. Push a la rama (`git push origin feature/AmazingFeature`)
5. Abrir un Pull Request

## Licencia

Este proyecto está bajo la licencia MIT. Ver el archivo `LICENSE` para más detalles.

## Links Útiles

- [Financial Modeling Prep API Docs](https://financialmodelingprep.com/developer/docs)
- [fmpcloud-go Library](https://github.com/spacecodewor/fmpcloud-go)
- [EOD Historical Data](https://eodhd.com/)
- [Polygon.io Documentation](https://polygon.io/docs)