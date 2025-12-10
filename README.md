# 🛡️ CipherWall VPN Server

A full-featured, **bidirectional VPN** server and client implementation in Go. CipherWall provides secure internet access through encrypted VPN tunnels, similar to OpenVPN but simpler and written in pure Go.

## 🌟 Features

- **🔄 Bidirectional Traffic** - Full VPN functionality (client ↔ server)
- **🌐 Internet Gateway** - Route all your traffic through the VPN server
- **🔐 Strong Encryption** - AES-256-CFB encryption + HMAC-SHA256 authentication
- **🔑 PBKDF2 Key Derivation** - Secure key generation from Pre-Shared Key (PSK)
- **🐳 Docker Ready** - Easy deployment with Docker and Dokploy
- **📡 UDP Transport** - Fast, lightweight protocol on port 1194

## 🚀 Quick Start

### Deploy Server (Dokploy)

See [DEPLOYMENT.md](DEPLOYMENT.md) for complete deployment guide to your Italy server.

### Run Client

```bash
# Build client
go build -o cipherwall-client client.go

# Connect (replace with your server's IP)
sudo ./cipherwall-client -server YOUR_SERVER_IP:1194
```

Now all your internet traffic goes through the VPN! 🌐

## 🔐 Security Features

- **AES-256 CFB Mode** encryption for data confidentiality
- **HMAC-SHA256** for message authentication and integrity
- **PBKDF2** key derivation from Pre-Shared Key (PSK)
- Protection against replay attacks through authenticated encryption

## 📋 Prerequisites

- **Go 1.18+** installed
- **Linux** operating system (uses `ip` commands for TUN interface management)
- **Root/sudo privileges** (required for TUN interface creation and configuration)

## 🚀 Installation

1. **Clone or download the project:**

   ```bash
   cd /path/to/cipherwall
   ```

2. **Install dependencies:**

   ```bash
   go mod download
   ```

3. **Build the server:**
   ```bash
   go build -o cipherwall-server main.go
   ```

## ⚙️ Configuration

### Server Configuration

Edit the constants in `main.go` before deploying:

```go
const (
    PSK         = "your-very-secure-32byte-key!!" // MUST be exactly 32 bytes
    SERVER_IP   = "10.8.0.1/24"                   // TUN interface IP
    UDP_PORT    = 1194                             // UDP listening port
)
```

### Client Configuration

Edit the constants in `client.go` to match server:

```go
const (
    PSK       = "your-very-secure-32byte-key!!" // MUST match server
    CLIENT_IP = "10.8.0.2/24"                   // Client TUN interface IP
)
```

### 🔑 Important: Change the PSK!

**Before deploying, you MUST change the Pre-Shared Key** to a secure, random 32-byte string:

```go
PSK = "your-secure-random-32byte-key!!"
```

You can generate a secure key using:

```bash
openssl rand -base64 32 | cut -c1-32
```

## 🎯 Usage

### Start the Server

Run with sudo (required for TUN interface management):

```bash
sudo ./cipherwall-server
```

Expected output:

```
🛡️  CipherWall VPN Server Starting...
📦 Deriving encryption and authentication keys from PSK...
✅ Keys derived successfully (AES: 32 bytes, HMAC: 32 bytes)
🌐 Setting up TUN interface...
✅ TUN interface 'tun0' created and configured with IP 10.8.0.1/24
🔌 Starting UDP listener on port 1194...
✅ UDP listener started successfully on 0.0.0.0:1194
🚀 Starting packet handler...
✅ CipherWall VPN Server is running!
📡 Waiting for incoming VPN connections...
```

### Stop the Server

Press `Ctrl+C` to stop the server.

## 📡 Protocol Specification

### Packet Format

Encrypted packets received by the server must follow this structure:

```
[HMAC_TAG (32 bytes)][IV (16 bytes)][ENCRYPTED_DATA (variable)]
```

1. **HMAC Tag** (32 bytes): HMAC-SHA256 authentication tag computed over `[IV + ENCRYPTED_DATA]`
2. **IV** (16 bytes): Initialization Vector for AES-CFB mode
3. **Encrypted Data**: The encrypted IP packet payload

### Encryption Process (Client-side)

1. Derive AES and HMAC keys from the shared PSK using PBKDF2
2. Generate a random 16-byte IV
3. Encrypt the IP packet using AES-256-CFB with the IV
4. Concatenate: `[IV][Ciphertext]`
5. Calculate HMAC-SHA256 over `[IV][Ciphertext]`
6. Send to server: `[HMAC][IV][Ciphertext]`

### Decryption Process (Server-side)

1. Receive UDP packet
2. Extract HMAC tag (first 32 bytes)
3. Verify HMAC over remaining data
4. If valid, extract IV (next 16 bytes)
5. Decrypt ciphertext using AES-256-CFB
6. Inject decrypted IP packet into TUN interface

## 🏗️ Architecture

```
Client → [Encrypted UDP Packet] → CipherWall Server → TUN Interface → Local Network Stack
```

### Key Components

- **UDP Listener**: Receives encrypted packets on port 1194
- **HMAC Verification**: Authenticates packets before processing
- **AES Decryption**: Decrypts validated packets
- **TUN Interface**: Injects decrypted IP packets into the OS network stack

## 🧪 Testing

### Check TUN Interface

After starting the server:

```bash
ip addr show tun0
```

Expected output:

```
tun0: <POINTOPOINT,MULTICAST,NOARP,UP,LOWER_UP> mtu 1500
    inet 10.8.0.1/24 scope global tun0
```

### Test with a Simple Client

Create a test client that:

1. Reads the same PSK
2. Derives keys using the same PBKDF2 parameters
3. Encrypts test IP packets
4. Sends to the server's UDP port

## 🔧 Troubleshooting

### Permission Denied Errors

**Issue**: `permission denied` when creating TUN interface

**Solution**: Run with sudo:

```bash
sudo ./cipherwall-server
```

### TUN Interface Already Exists

**Issue**: `RTNETLINK answers: File exists`

**Solution**: Delete the existing interface:

```bash
sudo ip link delete tun0
```

### Port Already in Use

**Issue**: `bind: address already in use`

**Solution**: Check if another process is using port 1194:

```bash
sudo lsof -i :1194
```

Kill the process or change `UDP_PORT` in the code.

## 🔒 Security Considerations

### Production Deployment Checklist

- [ ] **Change the PSK** to a cryptographically secure random value
- [ ] **Use a random salt** instead of the hardcoded salt for PBKDF2
- [ ] **Implement rate limiting** to prevent DoS attacks
- [ ] **Add client authentication** beyond the shared PSK
- [ ] **Enable firewall rules** to restrict access to port 1194
- [ ] **Use TLS/DTLS** for additional transport security (future enhancement)
- [ ] **Implement proper key rotation** mechanisms
- [ ] **Log to a secure location** with proper rotation

### Known Limitations

- **Unidirectional**: Only handles client-to-server traffic
- **No replay protection**: Same encrypted packet can be replayed
- **Single PSK**: All clients share the same key
- **No perfect forward secrecy**: Compromised PSK affects all sessions

## 📚 Dependencies

- [`github.com/songgao/water`](https://github.com/songgao/water) - TUN/TAP interface management
- [`golang.org/x/crypto/pbkdf2`](https://pkg.go.dev/golang.org/x/crypto/pbkdf2) - Key derivation function
- Standard Go libraries: `crypto/aes`, `crypto/cipher`, `crypto/hmac`, `crypto/sha256`, `net`

## 📄 License

This is a minimal educational implementation. Review and enhance security features before production use.

## 🤝 Contributing

This is a minimal implementation for learning purposes. For production VPN needs, consider established solutions like WireGuard or OpenVPN.

## 📞 Support

For issues or questions:

- Check the troubleshooting section above
- Review the Go documentation for networking and crypto packages
- Ensure your system meets all prerequisites

---

**⚠️ Warning**: This is a minimal implementation intended for educational purposes. Thoroughly review and test before using in production environments.
