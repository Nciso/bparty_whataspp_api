package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	_ "github.com/mattn/go-sqlite3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types/events"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	waLog "go.mau.fi/whatsmeow/util/log"
)

var (
	client       *whatsmeow.Client
	qrCode       string
	qrMutex      sync.RWMutex
	isConnected  bool
	statusMutex  sync.RWMutex
)

type WebhookRequest struct {
	Phone   string `json:"phone"`   // Phone number with country code (e.g., "521234567890")
	Message string `json:"message"` // Message to send
}

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

func initWhatsApp() error {
	// Setup database for storing session
	dbLog := waLog.Stdout("Database", "INFO", true)
	container, err := sqlstore.New(context.Background(), "sqlite3", "file:whatsapp.db?_foreign_keys=on", dbLog)
	if err != nil {
		return fmt.Errorf("failed to create store: %v", err)
	}

	// Get first device from store or create new
	deviceStore, err := container.GetFirstDevice(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get device: %v", err)
	}

	clientLog := waLog.Stdout("Client", "INFO", true)
	client = whatsmeow.NewClient(deviceStore, clientLog)

	// Register event handler
	client.AddEventHandler(func(evt interface{}) {
		switch evt.(type) {
		case *events.Connected:
			statusMutex.Lock()
			isConnected = true
			statusMutex.Unlock()
			log.Println("Connected to WhatsApp!")
		case *events.Disconnected:
			statusMutex.Lock()
			isConnected = false
			statusMutex.Unlock()
			log.Println("Disconnected from WhatsApp")
		}
	})

	// Check if already logged in
	if client.Store.ID == nil {
		// Not logged in, need to pair
		log.Println("Not logged in. QR code will be available at /qr endpoint")
		
		go func() {
			qrChan, _ := client.GetQRChannel(context.Background())
			err = client.Connect()
			if err != nil {
				log.Printf("Failed to connect: %v", err)
				return
			}

			for evt := range qrChan {
				if evt.Event == "code" {
					qrMutex.Lock()
					qrCode = evt.Code
					qrMutex.Unlock()
					log.Println("QR Code available at /qr endpoint")
					log.Println("QR Code (terminal):", evt.Code)
				} else {
					log.Println("Login event:", evt.Event)
					if evt.Event == "success" {
						qrMutex.Lock()
						qrCode = ""
						qrMutex.Unlock()
						statusMutex.Lock()
						isConnected = true
						statusMutex.Unlock()
					}
				}
			}
		}()
	} else {
		// Already logged in, just connect
		err = client.Connect()
		if err != nil {
			return fmt.Errorf("failed to connect: %v", err)
		}
		statusMutex.Lock()
		isConnected = true
		statusMutex.Unlock()
		log.Println("Connected to WhatsApp successfully!")
	}

	return nil
}

func sendMessageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Error:   "Only POST method is allowed",
		})
		return
	}

	var req WebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Error:   "Invalid JSON payload",
		})
		return
	}

	// Validate required fields
	if req.Phone == "" || req.Message == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Error:   "Both 'phone' and 'message' fields are required",
		})
		return
	}

	// Check if client is connected
	statusMutex.RLock()
	connected := isConnected
	statusMutex.RUnlock()
	
	if !connected {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Error:   "WhatsApp client is not connected. Please scan QR code at /qr",
		})
		return
	}

	// Parse phone number (expects format like "521234567890")
	jid := types.NewJID(req.Phone, types.DefaultUserServer)

	// Send message
	_, err := client.SendMessage(context.Background(), jid, &proto.Message{
		Conversation: &req.Message,
	})

	if err != nil {
		log.Printf("Error sending message: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Error:   fmt.Sprintf("Failed to send message: %v", err),
		})
		return
	}

	log.Printf("Message sent successfully to %s", req.Phone)
	json.NewEncoder(w).Encode(Response{
		Success: true,
		Message: "Message sent successfully",
	})
}

func qrHandler(w http.ResponseWriter, r *http.Request) {
	qrMutex.RLock()
	currentQR := qrCode
	qrMutex.RUnlock()

	statusMutex.RLock()
	connected := isConnected
	statusMutex.RUnlock()

	if connected {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>WhatsApp Connected</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style>
        body {
            font-family: Arial, sans-serif;
            display: flex;
            justify-content: center;
            align-items: center;
            min-height: 100vh;
            margin: 0;
            background: #075e54;
        }
        .container {
            text-align: center;
            background: white;
            padding: 40px;
            border-radius: 10px;
            box-shadow: 0 4px 6px rgba(0,0,0,0.1);
        }
        .checkmark {
            font-size: 80px;
            color: #25d366;
        }
        h1 {
            color: #075e54;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="checkmark">✓</div>
        <h1>WhatsApp Connected!</h1>
        <p>Your WhatsApp is connected and ready to send messages.</p>
    </div>
</body>
</html>
		`)
		return
	}

	if currentQR == "" {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>Waiting for QR Code</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <meta http-equiv="refresh" content="2">
    <style>
        body {
            font-family: Arial, sans-serif;
            display: flex;
            justify-content: center;
            align-items: center;
            min-height: 100vh;
            margin: 0;
            background: #075e54;
        }
        .container {
            text-align: center;
            background: white;
            padding: 40px;
            border-radius: 10px;
            box-shadow: 0 4px 6px rgba(0,0,0,0.1);
        }
        .spinner {
            border: 4px solid #f3f3f3;
            border-top: 4px solid #25d366;
            border-radius: 50%%;
            width: 60px;
            height: 60px;
            animation: spin 1s linear infinite;
            margin: 20px auto;
        }
        @keyframes spin {
            0%% { transform: rotate(0deg); }
            100%% { transform: rotate(360deg); }
        }
        h1 {
            color: #075e54;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="spinner"></div>
        <h1>Generating QR Code...</h1>
        <p>Please wait while we generate your QR code.</p>
        <p><small>This page will refresh automatically.</small></p>
    </div>
</body>
</html>
		`)
		return
	}

	// Display QR code using an HTML page with QR library
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>WhatsApp QR Code</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <script src="https://cdnjs.cloudflare.com/ajax/libs/qrcodejs/1.0.0/qrcode.min.js"></script>
    <style>
        body {
            font-family: Arial, sans-serif;
            display: flex;
            justify-content: center;
            align-items: center;
            min-height: 100vh;
            margin: 0;
            background: #075e54;
        }
        .container {
            text-align: center;
            background: white;
            padding: 40px;
            border-radius: 10px;
            box-shadow: 0 4px 6px rgba(0,0,0,0.1);
            max-width: 500px;
        }
        #qrcode {
            margin: 20px auto;
            display: inline-block;
        }
        h1 {
            color: #075e54;
            margin-bottom: 10px;
        }
        .instructions {
            color: #666;
            margin: 20px 0;
            line-height: 1.6;
        }
        .step {
            margin: 10px 0;
            text-align: left;
        }
        .refresh-note {
            color: #999;
            font-size: 12px;
            margin-top: 20px;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>Connect WhatsApp</h1>
        <div id="qrcode"></div>
        <div class="instructions">
            <div class="step">1. Open WhatsApp on your phone</div>
            <div class="step">2. Tap Menu or Settings</div>
            <div class="step">3. Tap Linked Devices</div>
            <div class="step">4. Tap Link a Device</div>
            <div class="step">5. Scan this QR code</div>
        </div>
        <div class="refresh-note">This page will refresh every 5 seconds</div>
    </div>
    <script>
        new QRCode(document.getElementById("qrcode"), {
            text: %q,
            width: 256,
            height: 256,
        });
        setTimeout(function() {
            location.reload();
        }, 5000);
    </script>
</body>
</html>
	`, currentQR)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	statusMutex.RLock()
	connected := isConnected
	statusMutex.RUnlock()
	
	qrMutex.RLock()
	hasQR := qrCode != ""
	qrMutex.RUnlock()
	
	status := "disconnected"
	if connected {
		status = "connected"
	} else if hasQR {
		status = "waiting_for_qr_scan"
	}
	
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     status,
		"connected":  connected,
		"qr_available": hasQR,
	})
}

func main() {
	// Initialize WhatsApp client
	log.Println("Initializing WhatsApp client...")
	if err := initWhatsApp(); err != nil {
		log.Fatalf("Failed to initialize WhatsApp: %v", err)
	}

	// Setup HTTP handlers
	http.HandleFunc("/send", sendMessageHandler)
	http.HandleFunc("/qr", qrHandler)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>WhatsApp API</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style>
        body {
            font-family: Arial, sans-serif;
            max-width: 800px;
            margin: 50px auto;
            padding: 20px;
            background: #f5f5f5;
        }
        .container {
            background: white;
            padding: 30px;
            border-radius: 10px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        h1 { color: #075e54; }
        .endpoint {
            background: #f9f9f9;
            padding: 15px;
            margin: 10px 0;
            border-left: 4px solid #25d366;
            border-radius: 4px;
        }
        .endpoint h3 { margin-top: 0; color: #075e54; }
        code {
            background: #e8e8e8;
            padding: 2px 6px;
            border-radius: 3px;
            font-family: monospace;
        }
        .button {
            display: inline-block;
            background: #25d366;
            color: white;
            padding: 10px 20px;
            text-decoration: none;
            border-radius: 5px;
            margin: 10px 0;
        }
        .button:hover { background: #20ba5a; }
    </style>
</head>
<body>
    <div class="container">
        <h1>📱 WhatsApp API</h1>
        <p>Simple WhatsApp messaging API built with Go and whatsmeow</p>
        
        <a href="/qr" class="button">🔗 Connect WhatsApp</a>
        <a href="/health" class="button" style="background: #128C7E;">📊 Check Status</a>
        
        <h2>Endpoints</h2>
        
        <div class="endpoint">
            <h3>POST /send</h3>
            <p>Send a WhatsApp message</p>
            <pre><code>{
  "phone": "521234567890",
  "message": "Hello!"
}</code></pre>
        </div>
        
        <div class="endpoint">
            <h3>GET /qr</h3>
            <p>Display QR code for WhatsApp authentication</p>
        </div>
        
        <div class="endpoint">
            <h3>GET /health</h3>
            <p>Check connection status</p>
        </div>
    </div>
</body>
</html>
		`)
	})

	// Get port from environment or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start server
	server := &http.Server{
		Addr: ":" + port,
	}

	go func() {
		log.Printf("Server starting on port %s...", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down gracefully...")
	client.Disconnect()
	server.Shutdown(context.Background())
}
