# Deploy a Railway.app - Guía Completa 🚂

## Por qué Railway?

✅ **Persistencia automática** - Tu archivo `whatsapp.db` se guarda solo
✅ **$5 gratis/mes** - Suficiente para este proyecto (~$2-3/mes de uso real)
✅ **Super fácil** - Literalmente 5 clicks
✅ **Logs en vivo** - Ves todo lo que pasa
✅ **HTTPS gratis** - Dominio y certificado incluido

## Paso a Paso

### 1. Prepara tu código
```bash
# Asegúrate de tener estos archivos en tu repo:
# - main.go
# - go.mod
# - Dockerfile
# - railway.json
# - .gitignore

git init
git add .
git commit -m "Initial commit"

# Sube a GitHub
git remote add origin https://github.com/tu-usuario/whatsapp-api.git
git push -u origin main
```

### 2. Crea cuenta en Railway
1. Ve a https://railway.app
2. Click en "Start a New Project"
3. Login con GitHub

### 3. Deploy tu proyecto
1. Click "New Project"
2. Selecciona "Deploy from GitHub repo"
3. Busca tu repo `whatsapp-api`
4. Railway detecta automáticamente el Dockerfile
5. Click "Deploy Now"
6. **Espera 2-3 minutos** mientras se construye

### 4. Configura el dominio
1. En tu proyecto, ve a "Settings"
2. Click "Generate Domain"
3. Railway te da un dominio tipo: `whatsapp-api-production.up.railway.app`
4. Copia ese dominio

### 5. Conecta WhatsApp
1. Abre tu navegador
2. Ve a: `https://TU-DOMINIO.railway.app/qr`
3. Escanea el QR con WhatsApp:
   - Abre WhatsApp en tu teléfono
   - Ve a Configuración → Dispositivos vinculados
   - Toca "Vincular un dispositivo"
   - Escanea el QR de la pantalla
4. La página se refresca automáticamente cuando conecta
5. Verás "WhatsApp Connected!" ✅

### 6. Prueba que funciona
```bash
# Envía un mensaje de prueba
curl -X POST https://TU-DOMINIO.railway.app/send \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "5213121234567",
    "message": "Hola desde mi API!"
  }'

# Deberías recibir:
# {"success":true,"message":"Message sent successfully"}
```

### 7. Ver logs (opcional)
1. En Railway, click en tu proyecto
2. Ve a "Deployments"
3. Click en el deployment activo
4. Ves los logs en tiempo real

## Costos

Railway cobra por uso:
- **CPU**: ~$0.000463/min
- **RAM**: ~$0.000231/min por GB
- **Network**: Primer GB gratis

**Tu API en idle**: ~$0.50/mes
**Con tráfico moderado**: ~$2-3/mes

**Crédito gratis**: $5/mes → Suficiente para empezar

## Persistencia en Railway

✅ **Railway guarda automáticamente el filesystem**
- Tu `whatsapp.db` persiste entre deploys
- No necesitas configurar volúmenes
- Incluso si Railway reinicia tu app, tu sesión sigue ahí

## Variables de Entorno (opcional)

Si quieres agregar API key u otras configs:

1. En Railway → Settings → Variables
2. Agrega variables:
   ```
   API_KEY=tu-api-key-secreta
   PORT=8080  (Railway lo pone automáticamente)
   ```

## Troubleshooting

### "Application failed to respond"
- Railway puede tardar 2-3 min en el primer deploy
- Revisa los logs para ver errores

### "QR code not showing"
- Espera 30 segundos después del deploy
- Refresca la página /qr
- Revisa logs: `railway logs`

### "WhatsApp disconnected"
- Ve a `/qr` y escanea de nuevo
- La sesión se guarda, pero a veces WhatsApp desconecta
- Normal después de updates del código

### "Exceeded free tier"
- Revisa uso en Railway dashboard
- Si usas más de $5/mes, agrega tarjeta o reduce uso
- Considera poner tu app en "sleep" cuando no la uses

## Comandos útiles de Railway CLI (opcional)

```bash
# Instalar CLI
npm install -g @railway/cli

# Login
railway login

# Ver logs
railway logs

# Variables de entorno
railway variables

# SSH a tu contenedor
railway shell
```

## Próximos pasos

Una vez deployado:
1. ✅ Guarda tu URL: `https://TU-DOMINIO.railway.app`
2. ✅ Conecta WhatsApp en `/qr`
3. ✅ Prueba enviando mensajes con `/send`
4. ✅ Integra con JaroVerify u otros servicios

## Seguridad para Producción

Considera agregar:
```go
// API Key middleware
func apiKeyMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        apiKey := r.Header.Get("X-API-Key")
        if apiKey != os.Getenv("API_KEY") {
            w.WriteHeader(http.StatusUnauthorized)
            return
        }
        next(w, r)
    }
}

// Úsalo así:
http.HandleFunc("/send", apiKeyMiddleware(sendMessageHandler))
```

¡Listo! 🎉 Ahora tienes tu API de WhatsApp corriendo 24/7 con persistencia.
