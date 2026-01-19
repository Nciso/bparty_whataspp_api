# WhatsApp API with Go

Simple WhatsApp API using whatsmeow that receives webhooks and sends messages.

## Features

- 📱 Connect WhatsApp via web-based QR code (no terminal needed!)
- 📤 Send WhatsApp messages via HTTP POST
- 💾 Persistent session storage
- 🔍 Health check endpoint

## Quick Start

1. Install dependencies:
```bash
go mod download
```

2. Run the server:
```bash
go run main.go
```

3. Open browser and go to `http://localhost:8080/qr` to scan QR code

## API Endpoints

### GET / 
Homepage with API documentation

### GET /qr
**Display QR code for WhatsApp connection**
- Opens a web page with QR code
- Scan with WhatsApp on your phone
- Auto-refreshes until connected
- No terminal access needed!

### POST /send
Send a WhatsApp message

**Request:**
```json
{
  "phone": "523332015171",
  "message": "Hello from the API!"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Message sent successfully"
}
```

### GET /health
Check connection status

**Response:**
```json
{
  "status": "connected",
  "connected": true,
  "qr_available": false
}
```

Status values:
- `connected` - WhatsApp is connected and ready
- `waiting_for_qr_scan` - QR code is available at /qr
- `disconnected` - Not connected

## Usage Example

```bash
# Send a message
curl -X POST https://your-app.railway.app/send \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "521234567890",
    "message": "Test message"
  }'

# Check status
curl https://your-app.railway.app/health
```

## Free Hosting with Persistent Storage ✅

### 1. **Railway.app** ⭐ RECOMENDADO
**Persistencia: SÍ (automática)**

✅ $5/mes gratis (suficiente para este proyecto)
✅ Persistencia automática del filesystem
✅ Deploy super fácil
✅ Logs en tiempo real
✅ Custom domains gratis

**Pasos:**
1. Crea cuenta en https://railway.app
2. "New Project" → "Deploy from GitHub repo"
3. Selecciona tu repo
4. Railway detecta el Dockerfile automáticamente
5. Deploy! 
6. Abre `https://tu-app.railway.app/qr` para escanear QR
7. **IMPORTANTE**: Railway mantiene el filesystem, tu sesión persiste automáticamente

**Costo**: Gratis ($5 crédito/mes, este proyecto usa ~$2-3/mes)

---

### 2. **Fly.io** ⭐ BUENA OPCIÓN
**Persistencia: SÍ (con volumen)**

✅ Tier gratis con 3 VMs
✅ Volúmenes persistentes incluidos
✅ Global deployment
✅ CLI poderoso

**Pasos:**
1. Instala Fly CLI: `curl -L https://fly.io/install.sh | sh`
2. Login: `fly auth login`
3. En tu proyecto: `fly launch` (usa el fly.toml incluido)
4. Crea volumen: `fly volumes create whatsapp_data --size 1`
5. Deploy: `fly deploy`
6. Abre: `https://tu-app.fly.dev/qr`

**Costo**: Gratis (con límites generosos)

---

### 3. **Render.com** ⚠️ CON LIMITACIONES
**Persistencia: NO en tier gratis**

⚠️ Tier gratis NO tiene discos persistentes
⚠️ Necesitas plan pagado ($7/mes) para persistencia
✅ Fácil de usar
✅ Auto-deploy desde GitHub

**Solo recomiendo si pagas**: Plan "Starter" ($7/mes) incluye persistent disk

---

### 4. **Koyeb** 
**Persistencia: NO en tier gratis**

❌ No soporta volúmenes persistentes en free tier
❌ No recomendado para este proyecto

---

## Comparación Rápida

| Servicio | Persistencia Gratis | Costo | Recomendación |
|----------|---------------------|-------|---------------|
| Railway  | ✅ Sí | $0-3/mes | ⭐⭐⭐⭐⭐ |
| Fly.io   | ✅ Sí | $0 | ⭐⭐⭐⭐ |
| Render   | ❌ No | $7/mes | ⭐⭐ |
| Koyeb    | ❌ No | $0 | ❌ |

## Mi Recomendación: Railway.app 🚂

Railway es la mejor opción porque:
1. **Persistencia automática** - no necesitas configurar nada
2. **$5 gratis/mes** - suficiente para este proyecto
3. **Super fácil** - conectas GitHub y listo
4. **Logs en vivo** - ves todo en tiempo real
5. **QR en el browser** - abres /qr y escaneas

## Cómo Conectar WhatsApp

### Opción 1: Browser (Recomendado)
1. Abre `https://tu-app.railway.app/qr` en tu navegador
2. Escanea el QR con WhatsApp
3. ¡Listo! La página se actualiza cuando conecta

### Opción 2: Terminal (si tienes acceso)
1. Los logs muestran el QR en texto
2. También puedes verlo en Railway logs

### Opción 3: API
```bash
# Ver si necesitas escanear QR
curl https://tu-app.railway.app/health

# Si status es "waiting_for_qr_scan", abre /qr en browser
```

## Important Notes

1. **Phone number format**: Use country code + number (e.g., "521234567890" for Mexico)
2. **Session persistence**: The `whatsapp.db` file stores your session
3. **QR Scanning**: Go to `/qr` endpoint to see QR code in browser
4. **Reconnection**: If disconnected, go back to `/qr` to get new QR code

## Environment Variables

- `PORT`: Server port (default: 8080, Railway sets this automatically)

## Security Considerations

For production use, consider adding:
- API key authentication
- Rate limiting  
- HTTPS only (Railway provides this automatically)
- Input validation
- Webhook signature verification

## Troubleshooting

**"WhatsApp client is not connected"**
- Go to `/qr` endpoint and scan QR code
- Check `/health` to see current status

**QR code not showing**
- Wait a few seconds, the page auto-refreshes
- Check Railway logs for errors

**Session lost after restart**
- Make sure you're using Railway or Fly.io with volumes
- Render free tier will lose session on restart
