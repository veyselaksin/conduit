//go:build client
// +build client

package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/songgao/water"
	"golang.org/x/crypto/pbkdf2"
)

const (
	// MUST be 32 bytes for AES-256
	PSK         = "this-is-strong-32byte-secret-key" // Exactly 32 bytes - MUST match server
	CLIENT_IP   = "10.8.0.2/24"
	BUFFER_SIZE = 1500
	KEY_LEN     = 32            // For AES-256
	HMAC_LEN    = 32            // SHA256 output size
	IV_LEN      = aes.BlockSize // 16 bytes for AES

	// PBKDF2 parameters - MUST match server
	PBKDF2_ITERATIONS = 100000
	PBKDF2_SALT       = "cipherwall-salt-2025"
)

// Global variables for derived keys and the TUN interface pointer
var (
	aesKey  []byte
	hmacKey []byte
	iface   *water.Interface
)

func main() {
	// Command line flags
	serverAddr := flag.String("server", "", "VPN server address (IP:PORT)")
	flag.Parse()

	if *serverAddr == "" {
		log.Fatal("❌ Server address is required. Usage: ./cipherwall-client -server <SERVER_IP>:1194")
	}

	log.Println("🛡️  CipherWall VPN Client Starting...")
	log.Printf("📡 Connecting to server: %s", *serverAddr)

	// 1. Derive Keys
	log.Println("📦 Deriving encryption and authentication keys from PSK...")
	deriveKeys([]byte(PSK))
	log.Printf("✅ Keys derived successfully (AES: %d bytes, HMAC: %d bytes)", len(aesKey), len(hmacKey))

	// 2. Setup TUN Interface
	log.Println("🌐 Setting up TUN interface...")
	var err error
	iface, err = setupTUN()
	if err != nil {
		log.Fatalf("❌ Failed to setup TUN interface: %v", err)
	}
	log.Printf("✅ TUN interface '%s' created and configured with IP %s", iface.Name(), CLIENT_IP)

	// 3. Setup UDP Connection
	log.Printf("🔌 Connecting to server %s...", *serverAddr)
	serverUDPAddr, err := net.ResolveUDPAddr("udp", *serverAddr)
	if err != nil {
		log.Fatalf("❌ Failed to resolve server address: %v", err)
	}

	conn, err := net.DialUDP("udp", nil, serverUDPAddr)
	if err != nil {
		log.Fatalf("❌ Failed to connect to server: %v", err)
	}
	defer conn.Close()
	log.Printf("✅ Connected to server successfully")

	// 4. Setup routing for all traffic through VPN
	log.Println("🔀 Configuring routing...")
	if err := setupRouting(*serverAddr); err != nil {
		log.Printf("⚠️  Failed to setup routing: %v", err)
		log.Println("⚠️  You may need to manually configure routes")
	} else {
		log.Println("✅ Routing configured successfully")
	}

	// 5. Start Packet Handlers (bidirectional)
	log.Println("🚀 Starting packet handlers...")
	go handleIncomingPackets(conn) // UDP -> TUN
	go handleOutgoingPackets(conn) // TUN -> UDP
	log.Println("✅ CipherWall VPN Client is running!")
	log.Println("🌐 All internet traffic is now routed through the VPN")

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("\n👋 Shutting down gracefully...")
	cleanupRouting(*serverAddr)
	log.Println("✅ Cleanup complete. Goodbye!")
}

// executeCommand runs a system command
func executeCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Printf("⚙️  Executing: %s %v", name, args)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command '%s %v' failed: %w", name, args, err)
	}

	return nil
}

// deriveKeys uses PBKDF2 to generate keys from PSK
func deriveKeys(psk []byte) {
	if len(psk) != 32 {
		log.Fatalf("❌ PSK must be exactly 32 bytes, got %d bytes", len(psk))
	}

	masterKey := pbkdf2.Key(psk, []byte(PBKDF2_SALT), PBKDF2_ITERATIONS, KEY_LEN*2, sha256.New)
	aesKey = masterKey[:KEY_LEN]
	hmacKey = masterKey[KEY_LEN:]
}

// setupTUN configures the virtual network interface
func setupTUN() (*water.Interface, error) {
	config := water.Config{
		DeviceType: water.TUN,
	}

	iface, err := water.New(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create TUN interface: %w", err)
	}

	ifaceName := iface.Name()
	log.Printf("📝 TUN interface created: %s", ifaceName)

	// Configure IP address based on OS
	if runtime.GOOS == "darwin" {
		// macOS uses ifconfig
		if err := executeCommand("ifconfig", ifaceName, "10.8.0.2", "10.8.0.1", "up"); err != nil {
			return nil, fmt.Errorf("failed to configure TUN interface: %w", err)
		}
	} else {
		// Linux uses ip command
		if err := executeCommand("ip", "addr", "add", CLIENT_IP, "dev", ifaceName); err != nil {
			return nil, fmt.Errorf("failed to assign IP to TUN interface: %w", err)
		}

		// Bring interface up
		if err := executeCommand("ip", "link", "set", "dev", ifaceName, "up"); err != nil {
			return nil, fmt.Errorf("failed to bring up TUN interface: %w", err)
		}
	}

	return iface, nil
}

// setupRouting configures system routes to send all traffic through VPN
func setupRouting(serverAddr string) error {
	// Extract server IP (remove port)
	host, _, err := net.SplitHostPort(serverAddr)
	if err != nil {
		return fmt.Errorf("failed to parse server address: %w", err)
	}

	if runtime.GOOS == "darwin" {
		// macOS routing setup
		// Get default gateway
		cmd := exec.Command("route", "-n", "get", "default")
		output, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("failed to get default gateway: %w", err)
		}

		log.Printf("📋 Current default route: %s", string(output))

		// Get the active network interface (en0 or en1)
		activeInterface := "en1" // Your system uses en1 based on the route output

		// Add specific route to VPN server through existing gateway
		// This must be done BEFORE changing default routes
		if err := executeCommand("route", "add", "-host", host, "-gateway", "192.168.1.1"); err != nil {
			log.Printf("⚠️  Warning: Failed to add server route: %v", err)
			// Try alternative method
			if err := executeCommand("route", "add", "-host", host, "-interface", activeInterface); err != nil {
				log.Printf("⚠️  Warning: Alternative server route also failed: %v", err)
			}
		}

		// Delete existing default route temporarily and add VPN as default
		// First, save the original gateway
		originalGateway := "192.168.1.1"

		// Delete old default route
		if err := executeCommand("route", "delete", "default"); err != nil {
			log.Printf("⚠️  Warning: Failed to delete default route: %v", err)
		}

		// Add new default route through VPN
		if err := executeCommand("route", "add", "default", "-interface", iface.Name()); err != nil {
			return fmt.Errorf("failed to add VPN default route: %w", err)
		}

		// Also add the /1 routes as backup
		if err := executeCommand("route", "add", "-net", "0.0.0.0/1", "-interface", iface.Name()); err != nil {
			log.Printf("⚠️  Warning: Failed to add 0/1 route: %v", err)
		}

		if err := executeCommand("route", "add", "-net", "128.0.0.0/1", "-interface", iface.Name()); err != nil {
			log.Printf("⚠️  Warning: Failed to add 128/1 route: %v", err)
		}

		_ = originalGateway // Will use for cleanup
	} else {
		// Linux routing setup
		// Get default gateway
		cmd := exec.Command("ip", "route", "show", "default")
		output, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("failed to get default gateway: %w", err)
		}

		log.Printf("📋 Current default route: %s", string(output))

		// Add route to VPN server through existing gateway to avoid routing loop
		if err := executeCommand("ip", "route", "add", host+"/32", "via", "default"); err != nil {
			log.Printf("⚠️  Failed to add server route (may already exist): %v", err)
		}

		// Add default route through VPN
		if err := executeCommand("ip", "route", "add", "0.0.0.0/1", "dev", iface.Name()); err != nil {
			return fmt.Errorf("failed to add VPN route: %w", err)
		}

		if err := executeCommand("ip", "route", "add", "128.0.0.0/1", "dev", iface.Name()); err != nil {
			return fmt.Errorf("failed to add VPN route: %w", err)
		}
	}

	return nil
}

// cleanupRouting removes VPN routes
func cleanupRouting(serverAddr string) {
	log.Println("🧹 Cleaning up routes...")

	if runtime.GOOS == "darwin" {
		// macOS cleanup
		// Delete VPN default route
		executeCommand("route", "delete", "default")

		// Delete backup routes
		executeCommand("route", "delete", "-net", "0.0.0.0/1")
		executeCommand("route", "delete", "-net", "128.0.0.0/1")

		// Delete server-specific route
		host, _, _ := net.SplitHostPort(serverAddr)
		executeCommand("route", "delete", "-host", host)

		// Restore original default route
		executeCommand("route", "add", "default", "192.168.1.1")
	} else {
		// Linux cleanup
		executeCommand("ip", "route", "del", "0.0.0.0/1")
		executeCommand("ip", "route", "del", "128.0.0.0/1")

		host, _, _ := net.SplitHostPort(serverAddr)
		executeCommand("ip", "route", "del", host+"/32")
	}
}

// handleIncomingPackets reads from UDP and writes to TUN after decrypting/authenticating
func handleIncomingPackets(conn *net.UDPConn) {
	buffer := make([]byte, BUFFER_SIZE)

	log.Println("🎯 Incoming packet handler ready (UDP -> TUN)")

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			log.Printf("⚠️  Error reading from UDP: %v", err)
			continue
		}

		if n < HMAC_LEN {
			log.Printf("⚠️  Packet too short (%d bytes)", n)
			continue
		}

		packet := buffer[:n]

		// Packet structure: [HMAC_TAG (32 bytes)][IV (16 bytes)][ENCRYPTED_DATA]
		receivedHMAC := packet[:HMAC_LEN]
		dataWithIV := packet[HMAC_LEN:]

		// Verify HMAC
		if !verifyHMAC(dataWithIV, receivedHMAC) {
			log.Printf("❌ HMAC verification failed")
			continue
		}

		// Decrypt the data
		decryptedData, err := decrypt(dataWithIV)
		if err != nil {
			log.Printf("⚠️  Decryption failed: %v", err)
			continue
		}

		// Write decrypted packet to TUN interface
		_, err = iface.Write(decryptedData)
		if err != nil {
			log.Printf("⚠️  Failed to write to TUN interface: %v", err)
			continue
		}

		log.Printf("📥 Received: %d bytes encrypted -> %d bytes decrypted", n, len(decryptedData))
	}
}

// handleOutgoingPackets reads from TUN and sends to UDP after encrypting/authenticating
func handleOutgoingPackets(conn *net.UDPConn) {
	buffer := make([]byte, BUFFER_SIZE)

	log.Println("🎯 Outgoing packet handler ready (TUN -> UDP)")

	for {
		n, err := iface.Read(buffer)
		if err != nil {
			log.Printf("⚠️  Error reading from TUN: %v", err)
			continue
		}

		packet := buffer[:n]

		// Encrypt and authenticate the packet
		encryptedPacket, err := encryptAndAuthenticate(packet)
		if err != nil {
			log.Printf("⚠️  Failed to encrypt packet: %v", err)
			continue
		}

		// Send to server
		_, err = conn.Write(encryptedPacket)
		if err != nil {
			log.Printf("⚠️  Failed to send packet: %v", err)
			continue
		}

		log.Printf("📤 Sent: %d bytes plaintext -> %d bytes encrypted", n, len(encryptedPacket))
	}
}

// verifyHMAC checks if the received HMAC matches the computed HMAC
func verifyHMAC(data, receivedHMAC []byte) bool {
	mac := hmac.New(sha256.New, hmacKey)
	mac.Write(data)
	expectedHMAC := mac.Sum(nil)
	return hmac.Equal(expectedHMAC, receivedHMAC)
}

// decrypt decrypts data using AES-256 CFB mode
func decrypt(data []byte) ([]byte, error) {
	if len(data) < IV_LEN {
		return nil, fmt.Errorf("data too short: need at least %d bytes for IV", IV_LEN)
	}

	iv := data[:IV_LEN]
	ciphertext := data[IV_LEN:]

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	stream := cipher.NewCFBDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	stream.XORKeyStream(plaintext, ciphertext)

	return plaintext, nil
}

// encrypt encrypts data using AES-256 CFB mode
func encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	iv := make([]byte, IV_LEN)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, fmt.Errorf("failed to generate IV: %w", err)
	}

	ciphertext := make([]byte, len(plaintext))
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext, plaintext)

	result := make([]byte, IV_LEN+len(ciphertext))
	copy(result[:IV_LEN], iv)
	copy(result[IV_LEN:], ciphertext)

	return result, nil
}

// addHMAC adds HMAC tag to data
func addHMAC(data []byte) []byte {
	mac := hmac.New(sha256.New, hmacKey)
	mac.Write(data)
	hmacTag := mac.Sum(nil)

	result := make([]byte, HMAC_LEN+len(data))
	copy(result[:HMAC_LEN], hmacTag)
	copy(result[HMAC_LEN:], data)

	return result
}

// encryptAndAuthenticate performs both encryption and authentication
func encryptAndAuthenticate(plaintext []byte) ([]byte, error) {
	encrypted, err := encrypt(plaintext)
	if err != nil {
		return nil, err
	}
	return addHMAC(encrypted), nil
}
