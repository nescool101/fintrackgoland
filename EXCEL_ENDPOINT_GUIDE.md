# 📊 Guía del Endpoint Excel - FinTrack GoLand

## 🎯 Resumen
Se ha agregado un nuevo endpoint `POST /api/excel/send` que permite:
1. Consumir datos financieros de índices (SPX, NDX, DJI, NYA, ES_F, NQ_F)
2. Generar un archivo Excel con la información
3. Enviar el archivo por correo electrónico a nescool101@gmail.com y paulocesarcelis@gmail.com

## 🔗 Endpoints

### 1. Reporte Básico (Solo Índices)
**URL:** `POST /api/excel/send`  
**Autenticación:** Basic Auth (nescao3:fintrack2024)  
**Contenido:** `application/json`  
**Descripción:** Genera reporte con índices objetivo (6 símbolos) o símbolos específicos

### 2. Reporte Completo (Todos los Símbolos)
**URL:** `GET /api/excel/full`  
**Autenticación:** Basic Auth (nescao3:fintrack2024)  
**Descripción:** Genera reporte completo con todos los símbolos (54 total) usando procesamiento por lotes

**⏰ Lógica de Fecha Automática:**
- **Antes de 3 PM:** Usa datos del día anterior
- **Después de 3 PM:** Usa datos del día actual
- **Fecha manual:** Especifica `?date=YYYY-MM-DD` para anular la lógica automática

## 📋 Parámetros (Opcionales)

| Parámetro | Tipo | Descripción | Valor por Defecto |
|-----------|------|-------------|-------------------|
| `symbols` | string | Símbolos separados por coma | `SPX,NDX,DJI,NYA,ES_F,NQ_F` |
| `date` | string | Fecha en formato YYYY-MM-DD | Automática (ver lógica de 3 PM) |
| `recipient` | string | Dirección de correo electrónico | `nescool101@gmail.com,paulocesarcelis@gmail.com` |

## 📤 Ejemplos de Uso

### Reporte Básico (Solo Índices)

#### 1. Envío básico (solo índices)
```bash
curl -X POST \
  -u "nescao3:fintrack2024" \
  "http://localhost:8080/api/excel/send"
```

#### 2. Envío con símbolos específicos
```bash
curl -X POST \
  -u "nescao3:fintrack2024" \
  "http://localhost:8080/api/excel/send?symbols=SPX,NDX,AAPL,NVDA&date=2024-01-15"
```

### Reporte Completo (Todos los Símbolos)

#### 3. Reporte completo con fecha automática
```bash
curl -X GET \
  -u "nescao3:fintrack2024" \
  "http://localhost:8080/api/excel/full"
# Antes de 3 PM: usa datos del día anterior
# Después de 3 PM: usa datos del día actual
```

#### 3b. Ver headers de fecha procesada
```bash
curl -X GET \
  -u "nescao3:fintrack2024" \
  -I "http://localhost:8080/api/excel/full"
# Muestra solo los headers HTTP incluyendo X-Processed-Date
```

#### 4. Reporte completo con fecha específica
```bash
curl -X GET \
  -u "nescao3:fintrack2024" \
  "http://localhost:8080/api/excel/full?date=2024-01-15"
# Anula la lógica automática y usa la fecha especificada
```

## ⏰ Lógica de Fecha Automática (Solo para `/api/excel/full`)

El endpoint `/api/excel/full` implementa una lógica inteligente para determinar qué fecha usar automáticamente:

### 🕐 Antes de las 3:00 PM
- **Fecha utilizada:** Día anterior
- **Razón:** Los datos del mercado del día actual aún no están completos
- **Ejemplo:** Si son las 10:00 AM del 15 de enero, usará datos del 14 de enero

### 🕒 Después de las 3:00 PM  
- **Fecha utilizada:** Día actual
- **Razón:** Los datos del mercado del día ya están disponibles
- **Ejemplo:** Si son las 4:00 PM del 15 de enero, usará datos del 15 de enero

### 📅 Anular la Lógica Automática
Para usar una fecha específica, simplemente agrega el parámetro `date`:
```bash
curl -X GET -u "nescao3:fintrack2024" \
  "http://localhost:8080/api/excel/full?date=2024-01-10"
```

**Nota:** Esta lógica solo aplica al endpoint `/api/excel/full`. El endpoint `/api/excel/send` siempre usa la fecha actual por defecto.

## 📊 Estructura del Archivo Excel

El archivo Excel generado contiene **2 hojas**:

### Hoja 1: "Reporte_Financiero"
Datos combinados de todos los símbolos con columnas:
| Columna | Descripción |
|---------|-------------|
| Tipo | Stock o Índice |
| Símbolo | Código del símbolo |
| Fecha | Fecha de los datos |
| Apertura | Precio de apertura |
| Máximo | Precio máximo |
| Mínimo | Precio mínimo |
| Cierre | Precio de cierre |
| Volumen | Volumen de transacciones |
| Fuente | API de origen |
| Estado | Estado de la consulta |

### Hoja 2: "Stocks"
Datos específicos de stocks únicamente

### Hoja 3: "Indices"
Datos específicos de índices únicamente

### Hoja 4: "Resumen"
- Total de símbolos procesados
- Distribución por tipo (Stocks vs Índices)
- Fuentes de datos utilizadas
- Estadísticas generales

## 📧 Email Automatizado

### 📮 Envío Individual
El sistema envía **correos individuales** a cada destinatario:
- **Destinatarios por defecto:** nescool101@gmail.com, paulocesarcelis@gmail.com
- **Método:** Un correo separado para cada dirección
- **Tolerancia a fallos:** Si un email falla, los otros se envían normalmente

### 📋 Contenido del Correo
- **Asunto:** 📊 Reporte Financiero - [FECHA]
- **Cuerpo:** Mensaje HTML con información del reporte
- **Adjunto:** Archivo Excel con nombre `Reporte_Financiero_YYYY-MM-DD.xlsx`

### 📊 Logs de Envío
```
✅ Correo enviado exitosamente a: nescool101@gmail.com
✅ Correo enviado exitosamente a: paulocesarcelis@gmail.com
📧 Correo enviado exitosamente a: [nescool101@gmail.com paulocesarcelis@gmail.com] (2 de 2 destinatarios)
```

## ✅ Respuesta de Éxito

### Respuesta JSON (Reporte Básico - `/api/excel/send`)
```json
{
  "message": "Reporte enviado exitosamente",
  "recipient": "nescool101@gmail.com",
  "date": "2024-01-15",
  "symbols_total": 6,
  "symbols_success": 5,
  "symbols_failed": 1,
  "excel_filename": "Reporte_Financiero_2024-01-15.xlsx",
  "excel_size_bytes": 15234,
  "data_summary": [
    {
      "symbol": "SPX",
      "date": "2024-01-15",
      "open": 600.05,
      "high": 603.72,
      "low": 599.52,
      "close": 603.68,
      "volume": 63904174,
      "from": "Financial Modeling Prep"
    }
  ]
}
```

### Respuesta JSON (Reporte Completo - `/api/excel/full`)
```json
{
  "message": "Reporte completo enviado exitosamente",
  "recipient": "nescool101@gmail.com,paulocesarcelis@gmail.com",
  "date": "2024-01-14",
  "symbols_total": 54,
  "symbols_success": 48,
  "symbols_failed": 6,
  "excel_filename": "Reporte_Completo_2024-01-14.xlsx",
  "excel_size_bytes": 45678,
  "batches_processed": 6,
  "date_logic": "Fecha automática: día anterior (antes de 3 PM)",
  "server_time": "2024-01-15 10:30:45",
  "data_summary": [...]
}
```

### Headers HTTP (Solo `/api/excel/full`)
```http
X-Processed-Date: 2024-01-14
X-Server-Time: 2024-01-15 10:30:45
X-Date-Logic: auto-previous-day
```

**Valores posibles para `X-Date-Logic`:**
- `manual`: Fecha especificada manualmente con parámetro `?date=`
- `auto-previous-day`: Fecha automática (día anterior, antes de 3 PM)
- `auto-current-day`: Fecha automática (día actual, después de 3 PM)

## ❌ Errores Comunes

### 400 - Formato de fecha inválido
```json
{
  "error": "Formato de fecha inválido. Use YYYY-MM-DD"
}
```

### 404 - Sin datos
```json
{
  "error": "No se encontraron datos para los símbolos especificados"
}
```

### 500 - Error interno
```json
{
  "error": "Error generando archivo Excel: [detalle]"
}
```

### 📧 Manejo de Errores de Email

#### ✅ Envío Parcial Exitoso
Si al menos un email se envía correctamente, el endpoint retorna éxito:
```json
{
  "message": "Reporte enviado exitosamente",
  "recipient": "nescool101@gmail.com,paulocesarcelis@gmail.com",
  ...
}
```

**Logs del servidor:**
```
✅ Correo enviado exitosamente a: nescool101@gmail.com
❌ Error enviando a paulocesarcelis@gmail.com: invalid address
⚠️ Algunos correos fallaron: Error enviando a paulocesarcelis@gmail.com: invalid address
📧 Correo enviado exitosamente a: [nescool101@gmail.com] (1 de 2 destinatarios)
```

#### ❌ Fallo Total de Email
Si ningún email se puede enviar:
```json
{
  "error": "Error enviando email: no se pudo enviar el correo a ningún destinatario: Error enviando a nescool101@gmail.com: [detalle]; Error enviando a paulocesarcelis@gmail.com: [detalle]"
}
```

## 🧪 Script de Prueba

Para probar el endpoint, ejecuta:
```bash
./test_excel_endpoint.sh
```

## 🔧 Configuración de Email

El endpoint utiliza la configuración SMTP del archivo `config_local.env`:
```env
EMAIL_HOST=smtp.gmail.com
EMAIL_PORT=587
EMAIL_USER=nescool10001@gmail.com
EMAIL_PASS=bndp fcme oyhh udyz
EMAIL_RECIPIENT=nescool101@gmail.com,paulocesarcelis@gmail.com
```

## 📱 APIs Utilizadas

- **FMP (Financial Modeling Prep):** Datos de stocks e índices
- **Límites:** 250 llamadas/día (FMP)
- **Nota:** Alpha Vantage se deshabilitó debido a límite muy bajo (25 llamadas/día)

## 🚀 Cómo Iniciar el Servicio

1. **Configurar variables de entorno:**
   ```bash
   cp config_local.env .env
   ```

2. **Iniciar servicio:**
   ```bash
   go run main.go
   ```

3. **Usar el endpoint:**
   ```bash
   curl -X POST -u "nescao3:fintrack2024" "http://localhost:8080/api/excel/send"
   # Enviará a ambos emails: nescool101@gmail.com y paulocesarcelis@gmail.com
   ```

4. **Verificar correo:**
   Revisa las bandejas de entrada de nescool101@gmail.com y paulocesarcelis@gmail.com

## 📊 Símbolos Soportados

| Símbolo | Descripción | Tipo |
|---------|-------------|------|
| SPX | S&P 500 Index | Índice |
| NDX | Nasdaq 100 Index | Índice |
| DJI | Dow Jones Industrial Average | Índice |
| NYA | NYSE Composite Index | Índice |
| ES_F | E-mini S&P 500 Futures | Futuro |
| NQ_F | E-mini Nasdaq-100 Futures | Futuro | 