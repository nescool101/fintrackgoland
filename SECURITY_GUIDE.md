# 🔒 Guía de Seguridad - FinTrack GoLand

## ⚠️ IMPORTANTE: Protección de Credenciales

### 🚨 Problema Detectado
GitGuardian ha detectado credenciales SMTP expuestas en el repositorio. **NUNCA** subas credenciales reales a GitHub.

## 🛡️ Configuración Segura

### 1. Configurar Variables de Entorno

```bash
# Copia la plantilla
cp env.template .env

# Edita el archivo .env con tus credenciales reales
nano .env
```

### 2. Credenciales Requeridas

#### 📊 APIs Financieras
- **FMP API Key**: Obtén en [Financial Modeling Prep](https://financialmodelingprep.com/developer/docs)
- **Alpha Vantage Key**: Obtén en [Alpha Vantage](https://www.alphavantage.co/support/#api-key)

#### 📧 Configuración de Email (Gmail)
Para usar Gmail SMTP de forma segura:

1. **Habilita 2FA** en tu cuenta de Google
2. Ve a **Cuenta de Google** → **Seguridad** → **Verificación en 2 pasos**
3. Busca **Contraseñas de aplicaciones**
4. Genera una contraseña para "Correo"
5. Usa esta contraseña de 16 caracteres en `EMAIL_PASS`

```env
EMAIL_HOST=smtp.gmail.com
EMAIL_PORT=587
EMAIL_USER=tu_email@gmail.com
EMAIL_PASS=abcd efgh ijkl mnop  # App Password de 16 caracteres
EMAIL_RECIPIENT=nescool101@gmail.com
```

#### 🔐 Autenticación Básica
```env
AUTH_USERNAME=tu_usuario_seguro
AUTH_PASSWORD=tu_password_fuerte_123
```

## 🚫 Archivos Protegidos

Los siguientes archivos están en `.gitignore` y **NO** se suben a GitHub:
- `.env`
- `config_local.env`
- `*.env`

## 🏭 Configuración de Producción

Para producción, usa variables de entorno del sistema:

```bash
export FMP_API_KEY="tu_clave_real"
export ALPHA_VANTAGE_KEY="tu_clave_real"
export EMAIL_USER="tu_email@gmail.com"
export EMAIL_PASS="tu_app_password"
export AUTH_USERNAME="usuario_prod"
export AUTH_PASSWORD="password_super_seguro"
```

## ✅ Verificación de Seguridad

### Antes de hacer commit:
```bash
# Verifica que no hay credenciales en archivos tracked
git status
git diff --cached

# Asegúrate de que .env no está en staging
ls -la .env  # Debe existir
git ls-files | grep .env  # NO debe mostrar .env
```

### Comandos seguros:
```bash
# ✅ CORRECTO: Usar plantilla
cp env.template .env

# ❌ INCORRECTO: Nunca hagas esto
git add .env
git add config_local.env
```

## 🔄 Rotación de Credenciales

Si tus credenciales fueron expuestas:

1. **Inmediatamente**:
   - Cambia tu contraseña de Gmail
   - Revoca la App Password comprometida
   - Genera nuevas API keys

2. **Actualiza configuración**:
   ```bash
   # Actualiza .env con nuevas credenciales
   nano .env
   
   # Reinicia el servicio
   go run main.go
   ```

## 📋 Checklist de Seguridad

- [ ] `.env` está en `.gitignore`
- [ ] No hay credenciales reales en archivos tracked
- [ ] Usando App Passwords para Gmail
- [ ] Contraseñas fuertes y únicas
- [ ] Variables de entorno en producción
- [ ] Credenciales rotadas si fueron expuestas

## 🆘 En Caso de Exposición

Si accidentalmente expusiste credenciales:

1. **Cambia todas las credenciales inmediatamente**
2. **Revoca API keys comprometidas**
3. **Limpia el historial de Git** (si es necesario):
   ```bash
   # CUIDADO: Esto reescribe la historia
   git filter-branch --force --index-filter \
   'git rm --cached --ignore-unmatch config_local.env' \
   --prune-empty --tag-name-filter cat -- --all
   ```

## 📞 Contacto de Seguridad

Si detectas problemas de seguridad, reporta inmediatamente:
- Cambia credenciales comprometidas
- Actualiza la configuración
- Documenta el incidente 