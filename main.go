//go:build !client
// +build !client

package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"

	"github.com/songgao/water"
	"golang.org/x/crypto/pbkdf2"
)

const (
	// MUST be 32 bytes for AES-256
	PSK         = "this-is-a-strong-32byte-secret" // Exactly 32 bytes
	SERVER_IP   = "10.8.0.1/24"
	UDP_PORT    = 1194
	BUFFER_SIZE = 1500
	KEY_LEN     = 32            // For AES-256
	HMAC_LEN    = 32            // SHA256 output size
	IV_LEN      = aes.BlockSize // 16 bytes for AES

	// PBKDF2 parameters
	PBKDF2_ITERATIONS = 100000
	PBKDF2_SALT       = "cipherwall-salt-2025" // In production, use a proper random salt
)

// Global variables for derived keys and the TUN interface pointer
var (
	aesKey      []byte
	hmacKey     []byte
	iface       *water.Interface
	clientAddrs map[string]*net.UDPAddr // Track client addresses
)

func main() {
	log.Println("🛡️  CipherWall VPN Server Starting...")

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
	log.Printf("✅ TUN interface '%s' created and configured with IP %s", iface.Name(), SERVER_IP)

	// 3. Setup UDP Listener
	log.Printf("🔌 Starting UDP listener on port %d...", UDP_PORT)
	udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", UDP_PORT))
	if err != nil {
		log.Fatalf("❌ Failed to resolve UDP address: %v", err)
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Fatalf("❌ Failed to start UDP listener: %v", err)
	}
	defer conn.Close()
	log.Printf("✅ UDP listener started successfully on 0.0.0.0:%d", UDP_PORT)

	// 4. Initialize client tracking
	clientAddrs = make(map[string]*net.UDPAddr)

	// 5. Start Packet Handlers (bidirectional)
	log.Println("🚀 Starting packet handlers...")
	go handleIncomingPackets(conn) // UDP -> TUN
	go handleOutgoingPackets(conn) // TUN -> UDP
	log.Println("✅ CipherWall VPN Server is running!")
	log.Println("📡 Waiting for incoming VPN connections...")

	// Keep main function alive
	select {}
}

// executeCommand runs a system command and logs its output/errors
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
	// Validate PSK length
	if len(psk) != 32 {
		log.Fatalf("❌ PSK must be exactly 32 bytes, got %d bytes", len(psk))
	}

	// Derive a master key of 64 bytes (32 for AES + 32 for HMAC)
	masterKey := pbkdf2.Key(psk, []byte(PBKDF2_SALT), PBKDF2_ITERATIONS, KEY_LEN*2, sha256.New)

	// Split the derived key
	aesKey = masterKey[:KEY_LEN]
	hmacKey = masterKey[KEY_LEN:]

	log.Printf("🔑 Derived AES key: %d bytes", len(aesKey))
	log.Printf("🔑 Derived HMAC key: %d bytes", len(hmacKey))
}

// setupTUN configures the virtual network interface
func setupTUN() (*water.Interface, error) {
	// Create TUN interface
	config := water.Config{
		DeviceType: water.TUN,
	}

	iface, err := water.New(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create TUN interface: %w", err)
	}

	ifaceName := iface.Name()
	log.Printf("📝 TUN interface created: %s", ifaceName)

	// Configure IP address (Linux-specific commands)
	if err := executeCommand("ip", "addr", "add", SERVER_IP, "dev", ifaceName); err != nil {
		return nil, fmt.Errorf("failed to assign IP to TUN interface: %w", err)
	}

	// Bring interface up
	if err := executeCommand("ip", "link", "set", "dev", ifaceName, "up"); err != nil {
		return nil, fmt.Errorf("failed to bring up TUN interface: %w", err)
	}

	return iface, nil
}

// handleIncomingPackets reads from UDP and writes to TUN after decrypting/authenticating
func handleIncomingPackets(conn *net.UDPConn) {
	buffer := make([]byte, BUFFER_SIZE)

	log.Println("🎯 Incoming packet handler ready (UDP -> TUN)")

	for {
		// Read from UDP
		n, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("⚠️  Error reading from UDP: %v", err)
			continue
		}

		// Track client address (use destination IP from decrypted packet as key)
		// For now, use the first client as default
		if _, exists := clientAddrs["default"]; !exists {
			log.Printf("👤 New client connected from: %s", addr.String())
			clientAddrs["default"] = addr
		} else if clientAddrs["default"].String() != addr.String() {
			log.Printf("👤 Client address updated: %s", addr.String())
			clientAddrs["default"] = addr
		}

		// Process the packet
		if n < HMAC_LEN {
			log.Printf("⚠️  Packet too short (%d bytes), expected at least %d bytes for HMAC", n, HMAC_LEN)
			continue
		}

		packet := buffer[:n]

		// Packet structure: [HMAC_TAG (32 bytes)][IV (16 bytes)][ENCRYPTED_DATA]
		receivedHMAC := packet[:HMAC_LEN]
		dataWithIV := packet[HMAC_LEN:]

		// Verify HMAC
		if !verifyHMAC(dataWithIV, receivedHMAC) {
			log.Printf("❌ HMAC verification failed for packet from %s", addr.String())
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

		log.Printf("✅ Processed packet: %d bytes encrypted -> %d bytes decrypted from %s",
			n, len(decryptedData), addr.String())
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
// Data format: [IV (16 bytes)][ENCRYPTED_DATA]
func decrypt(data []byte) ([]byte, error) {
	if len(data) < IV_LEN {
		return nil, fmt.Errorf("data too short: need at least %d bytes for IV, got %d", IV_LEN, len(data))
	}

	// Extract IV and ciphertext
	iv := data[:IV_LEN]
	ciphertext := data[IV_LEN:]

	// Create AES cipher block
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Decrypt using CFB mode
	stream := cipher.NewCFBDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	stream.XORKeyStream(plaintext, ciphertext)

	return plaintext, nil
}

// encrypt encrypts data using AES-256 CFB mode (for potential client-to-server responses)
// Returns: [IV][ENCRYPTED_DATA]
func encrypt(plaintext []byte) ([]byte, error) {
	// Create AES cipher block
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Generate random IV
	iv := make([]byte, IV_LEN)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, fmt.Errorf("failed to generate IV: %w", err)
	}

	// Encrypt using CFB mode
	ciphertext := make([]byte, len(plaintext))
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext, plaintext)

	// Prepend IV to ciphertext
	result := make([]byte, IV_LEN+len(ciphertext))
	copy(result[:IV_LEN], iv)
	copy(result[IV_LEN:], ciphertext)

	return result, nil
}

// addHMAC adds HMAC tag to data
// Returns: [HMAC_TAG][DATA]
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
// This is a helper function that combines encrypt and addHMAC
func encryptAndAuthenticate(plaintext []byte) ([]byte, error) {
	// First encrypt
	encrypted, err := encrypt(plaintext)
	if err != nil {
		return nil, err
	}

	// Then add HMAC
	return addHMAC(encrypted), nil
}

// handleOutgoingPackets reads from TUN and sends to UDP after encrypting/authenticating
func handleOutgoingPackets(conn *net.UDPConn) {
	buffer := make([]byte, BUFFER_SIZE)

	log.Println("🎯 Outgoing packet handler ready (TUN -> UDP)")

	for {
		// Read from TUN interface
		n, err := iface.Read(buffer)
		if err != nil {
			log.Printf("⚠️  Error reading from TUN: %v", err)
			continue
		}

		packet := buffer[:n]

		// Get client address (for now, send to the default client)
		clientAddr, exists := clientAddrs["default"]
		if !exists {
			// No client connected yet, drop packet
			continue
		}

		// Encrypt and authenticate the packet
		encryptedPacket, err := encryptAndAuthenticate(packet)
		if err != nil {
			log.Printf("⚠️  Failed to encrypt packet: %v", err)
			continue
		}

		// Send to client
		_, err = conn.WriteToUDP(encryptedPacket, clientAddr)
		if err != nil {
			log.Printf("⚠️  Failed to send packet to client: %v", err)
			continue
		}

		log.Printf("📤 Sent packet: %d bytes plaintext -> %d bytes encrypted to %s",
			n, len(encryptedPacket), clientAddr.String())
	}
}
