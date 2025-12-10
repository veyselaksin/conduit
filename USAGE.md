# 🎯 CipherWall VPN - Complete Usage Guide

## 📖 Table of Contents

- [What is CipherWall?](#what-is-cipherwall)
- [Quick Start](#quick-start)
- [Detailed Setup](#detailed-setup)
- [Using Your VPN](#using-your-vpn)
- [FAQ](#faq)

---

## 🛡️ What is CipherWall?

CipherWall is a **full-featured VPN** (like OpenVPN) that lets you:

- 🌍 **Access internet from any location** - Your traffic exits from the VPN server
- 🔒 **Encrypt all your traffic** - AES-256 military-grade encryption
- 🚀 **Easy deployment** - Works with Dokploy, Docker, or standalone
- 🇮🇹 **Perfect for your Italy server** - Route all traffic through Italy

### Real-World Example:

```
You in USA → Connect to Italy VPN → Browse internet as if you're in Italy
Your laptop → Encrypted tunnel → Italy server → Websites see Italy IP
```

---

## ⚡ Quick Start

### 1️⃣ Generate a Secure Key

```bash
openssl rand -base64 32 | cut -c1-32
```

**Save this key!** You'll need it for both server and client.

### 2️⃣ Deploy Server (Dokploy)

1. **Update `main.go` line 22:**

   ```go
   PSK = "your-generated-key-here-32bytes"
   ```

2. **Push to GitHub:**

   ```bash
   git add .
   git commit -m "Deploy CipherWall VPN"
   git push
   ```

3. **In Dokploy:**

   - New Application → Docker
   - Repository: Your GitHub URL
   - Dockerfile: `Dockerfile.dokploy`
   - Network Mode: **host**
   - Privileged: ✅ **Enable**
   - Device: `/dev/net/tun:/dev/net/tun`
   - Capabilities: `NET_ADMIN`, `NET_RAW`
   - Deploy!

4. **Open Firewall:**
   ```bash
   sudo ufw allow 1194/udp
   ```

### 3️⃣ Connect Client

1. **Update `client.go` line 22** with same key as server

2. **Build & Connect:**

   ```bash
   make client
   sudo ./cipherwall-client -server YOUR_ITALY_IP:1194
   ```

3. **Verify:**
   ```bash
   curl ifconfig.me  # Should show Italy IP!
   ```

**🎉 Done! You're now browsing through Italy!**

---

## 📚 Detailed Setup

### Server Deployment Options

#### Option A: Dokploy (Recommended)

See `DEPLOYMENT.md` for complete step-by-step guide.

#### Option B: Docker Compose

```bash
# 1. Update PSK in main.go
# 2. Build and run
docker-compose up -d

# Check logs
docker-compose logs -f
```

#### Option C: Standalone Binary

```bash
# Build
make server

# Run setup (NAT/routing)
sudo ./setup-server.sh

# Start server
sudo ./cipherwall-server
```

### Client Connection

```bash
# Build client
make client

# Option 1: Direct run
sudo ./cipherwall-client -server YOUR_SERVER_IP:1194

# Option 2: Using make
make run-client SERVER=YOUR_SERVER_IP:1194
```

---

## 🌐 Using Your VPN

### When Connected

**All internet traffic** goes through your VPN:

- ✅ Web browsing
- ✅ Apps and services
- ✅ Streaming
- ✅ Downloads
- ✅ Everything!

### Verify Connection

```bash
# Check your public IP (should be VPN server's IP)
curl ifconfig.me

# Detailed info
curl ipinfo.io

# Check VPN interface
ip addr show tun0
```

### Monitor Traffic

Client shows real-time stats:

```
📥 Received: 1234 bytes encrypted -> 1200 bytes decrypted
📤 Sent: 567 bytes plaintext -> 615 bytes encrypted
```

### Disconnect

Press `Ctrl+C` in the client terminal. It will:

- Clean up routes
- Restore normal connection
- Exit gracefully

---

## ❓ FAQ

### Q: How is this different from OpenVPN?

**CipherWall:**

- ✅ Simpler setup (no certificates)
- ✅ Easy to understand code
- ✅ Perfect for personal use
- ✅ Works with Dokploy

**OpenVPN:**

- ✅ Enterprise features
- ✅ More mature
- ✅ Certificate-based auth
- ✅ Larger community

### Q: Is it secure?

**Yes!** Uses:

- AES-256-CFB encryption (same as OpenVPN)
- HMAC-SHA256 authentication
- PBKDF2 key derivation (100,000 iterations)

### Q: Can multiple clients connect?

Current implementation supports one client at a time. For multiple clients, you'd need to:

- Implement client session management
- Use unique keys per client
- Or use the same PSK (less secure)

### Q: What about DNS leaks?

Configure your DNS to prevent leaks:

```bash
# Use public DNS through VPN
sudo tee /etc/resolv.conf << EOF
nameserver 8.8.8.8
nameserver 8.8.4.4
EOF
```

Or use `systemd-resolved`:

```bash
sudo systemd-resolve --interface tun0 --set-dns 8.8.8.8
```

### Q: Does it work on macOS?

The client works on macOS! You may need to adjust routing commands:

```bash
# macOS uses different syntax
sudo route add -net 0.0.0.0/1 -interface tun0
```

### Q: Can I use it on mobile?

Not directly. You'd need to:

- Compile for mobile (Go supports iOS/Android)
- Handle mobile-specific networking
- Create a UI/app wrapper

### Q: How fast is it?

**Very fast!** UDP protocol with minimal overhead. Speed depends on:

- Your connection speed
- Server connection speed
- Distance to server
- Server load

### Q: What about IPv6?

Current implementation is IPv4 only. IPv6 support would require:

- IPv6 routing configuration
- Dual-stack setup
- IPv6 firewall rules

### Q: How do I change the port?

1. Update `UDP_PORT` in `main.go` and `client.go`
2. Rebuild both
3. Update firewall rules for new port

### Q: Can I run multiple servers?

Yes! Each server can use:

- Different port
- Different TUN interface
- Different subnet (e.g., 10.9.0.0/24)

---

## 🔧 Advanced Configuration

### Custom Network Subnet

Edit `main.go` and `client.go`:

```go
SERVER_IP = "10.9.0.1/24"  // main.go
CLIENT_IP = "10.9.0.2/24"  // client.go
```

### Different Port

```go
UDP_PORT = 443  // Use HTTPS port (less likely to be blocked)
```

### Persistent Routes (survives reboot)

Add to `/etc/network/interfaces` or create systemd service.

### Systemd Service (auto-start server)

Create `/etc/systemd/system/cipherwall.service`:

```ini
[Unit]
Description=CipherWall VPN Server
After=network.target

[Service]
Type=simple
ExecStartPre=/path/to/setup-server.sh
ExecStart=/path/to/cipherwall-server
Restart=always
User=root

[Install]
WantedBy=multi-user.target
```

Enable:

```bash
sudo systemctl enable cipherwall
sudo systemctl start cipherwall
```

---

## 📊 Monitoring & Logs

### View Logs

**Docker/Dokploy:**

```bash
docker logs -f cipherwall-server
```

**Standalone:**

```bash
# Logs go to stdout, redirect to file:
sudo ./cipherwall-server 2>&1 | tee cipherwall.log
```

### Check Connections

```bash
# See active VPN connections
sudo ss -ulnp | grep 1194

# Check TUN interface
ip addr show tun0

# View NAT rules
sudo iptables -t nat -L -n -v
```

---

## 🛠️ Troubleshooting Quick Reference

| Problem                   | Solution                                      |
| ------------------------- | --------------------------------------------- |
| Can't connect             | Check firewall, verify server is running      |
| HMAC verification failed  | PSK mismatch - ensure identical on both       |
| Connected but no internet | Run `setup-server.sh` on server               |
| Permission denied         | Run with `sudo`                               |
| TUN device error          | Ensure `/dev/net/tun` exists, run as root     |
| Port already in use       | Change `UDP_PORT` or stop conflicting service |

Detailed troubleshooting: See `DEPLOYMENT.md`

---

## 📝 Files Reference

| File                 | Purpose                      |
| -------------------- | ---------------------------- |
| `main.go`            | VPN Server code              |
| `client.go`          | VPN Client code              |
| `setup-server.sh`    | Configure server NAT/routing |
| `Dockerfile`         | Docker image (standard)      |
| `Dockerfile.dokploy` | Docker image (Dokploy)       |
| `docker-compose.yml` | Docker Compose config        |
| `DEPLOYMENT.md`      | Full deployment guide        |
| `QUICKSTART.md`      | Quick reference card         |
| `SUMMARY.md`         | Project overview             |
| `Makefile`           | Build automation             |

---

## 🎓 Learning Resources

**Want to understand how it works?**

1. Read the code - it's well-commented
2. Check `SUMMARY.md` for architecture overview
3. Learn about: TUN/TAP, AES encryption, HMAC, NAT

**Want to extend it?**

- Add multi-client support
- Implement certificate auth
- Add compression
- Build a web UI
- Add traffic statistics

---

## 🤝 Support

Need help? Check in order:

1. ✅ This guide (FAQ section)
2. ✅ `DEPLOYMENT.md` (deployment issues)
3. ✅ `QUICKSTART.md` (quick commands)
4. ✅ `SUMMARY.md` (technical details)

---

**🎉 Enjoy your secure, private VPN! 🛡️**

_Built with Go • Deployed with Dokploy • Secured with AES-256_
