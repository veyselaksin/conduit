# 🚀 CipherWall VPN - Deployment Guide for Dokploy

Complete guide to deploy CipherWall VPN server on your Italy server using Dokploy and connect from anywhere.

## 📋 Table of Contents

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Server Deployment (Dokploy)](#server-deployment-dokploy)
- [Client Setup](#client-setup)
- [Usage](#usage)
- [Troubleshooting](#troubleshooting)

---

## 🌍 Overview

This setup allows you to:

- ✅ Deploy CipherWall VPN server on your Italy server using Dokploy
- ✅ Connect from anywhere in the world
- ✅ Route ALL your internet traffic through the Italy server
- ✅ Browse the internet as if you're in Italy
- ✅ Secure your connection with AES-256 encryption

```
Your Device → [VPN Tunnel] → Italy Server → Internet
(Anywhere)      (Encrypted)    (Exit Point)
```

---

## 📋 Prerequisites

### On Your Italy Server:

- ✅ **Dokploy installed** and running
- ✅ **Docker support** in Dokploy
- ✅ **Public IP address** (note this IP, you'll need it)
- ✅ **UDP port 1194** open in firewall
- ✅ Root/sudo access

### On Your Client Device (Laptop/Desktop):

- ✅ **Linux/macOS** (Windows WSL2 also works)
- ✅ **Go 1.18+** installed
- ✅ Root/sudo access

---

## 🚀 Server Deployment (Dokploy)

### Step 1: Prepare Your Repository

1. **Push your code to GitHub** (or any git provider):

```bash
cd /path/to/conduit
git add .
git commit -m "Add CipherWall VPN server"
git push origin main
```

2. **Important: Update the PSK** before deploying!

Edit `main.go` and change the PSK to a secure 32-character string:

```go
const (
    PSK = "your-very-secure-32byte-key!!" // CHANGE THIS!
    // ...
)
```

Generate a secure key:

```bash
openssl rand -base64 32 | cut -c1-32
```

**⚠️ Save this key securely! You'll need it for the client.**

### Step 2: Deploy with Dokploy

1. **Log into your Dokploy dashboard** on your Italy server

2. **Create a new application:**

   - Click "New Application"
   - Name: `cipherwall-vpn`
   - Type: **Docker**

3. **Configure the application:**

   **Source:**

   - Repository: `https://github.com/veyselaksin/conduit` (your repo URL)
   - Branch: `main`
   - Dockerfile: `Dockerfile.dokploy`

   **Network:**

   - Network Mode: **host** (REQUIRED for VPN)
   - Port Mapping: Not needed (using host network)

   **Advanced Settings:**

   - Privileged Mode: **✅ Enable** (REQUIRED)
   - Add Capability: `NET_ADMIN`
   - Add Capability: `NET_RAW`

   **Devices:**

   - Add device: `/dev/net/tun:/dev/net/tun`

   **Environment Variables:**

   - `TZ=Europe/Rome` (or your timezone)

4. **Deploy:**
   - Click "Deploy"
   - Wait for build to complete (~2-3 minutes)
   - Check logs to ensure server started successfully

### Step 3: Verify Server is Running

In Dokploy logs, you should see:

```
🛡️  CipherWall VPN Server Starting...
✅ Keys derived successfully
✅ TUN interface 'tun0' created
✅ UDP listener started successfully on 0.0.0.0:1194
✅ CipherWall VPN Server is running!
📡 Waiting for incoming VPN connections...
```

### Step 4: Check Firewall

**Make sure UDP port 1194 is open!**

On your Italy server (SSH into it):

```bash
# Check if port is accessible
sudo ss -ulnp | grep 1194

# If using UFW:
sudo ufw allow 1194/udp
sudo ufw status

# If using firewalld:
sudo firewall-cmd --permanent --add-port=1194/udp
sudo firewall-cmd --reload

# If using iptables directly:
sudo iptables -I INPUT -p udp --dport 1194 -j ACCEPT
```

**Also check your cloud provider's firewall** (Security Groups, etc.):

- AWS: EC2 Security Groups
- DigitalOcean: Firewall settings
- Hetzner: Firewall rules
- etc.

---

## 💻 Client Setup

### Step 1: Build the Client

On your local machine (laptop/desktop):

```bash
# Clone the repository
git clone https://github.com/veyselaksin/conduit.git
cd conduit

# Build the client
go build -o cipherwall-client client.go
```

### Step 2: Update Client PSK

**IMPORTANT:** Edit `client.go` and change the PSK to match your server:

```go
const (
    PSK = "your-very-secure-32byte-key!!" // MUST match server!
    // ...
)
```

Then rebuild:

```bash
go build -o cipherwall-client client.go
```

### Step 3: Test Connection

```bash
# Replace with your Italy server's IP
sudo ./cipherwall-client -server YOUR_ITALY_SERVER_IP:1194
```

You should see:

```
🛡️  CipherWall VPN Client Starting...
📡 Connecting to server: YOUR_ITALY_SERVER_IP:1194
✅ Keys derived successfully
✅ TUN interface 'tun0' created and configured
✅ Connected to server successfully
✅ Routing configured successfully
✅ CipherWall VPN Client is running!
🌐 All internet traffic is now routed through the VPN
```

---

## 🎯 Usage

### Connect to VPN

```bash
sudo ./cipherwall-client -server YOUR_ITALY_SERVER_IP:1194
```

Once connected:

- ✅ All your internet traffic goes through Italy
- ✅ Your public IP appears as the Italy server's IP
- ✅ All traffic is encrypted with AES-256

### Verify You're Connected

In a new terminal:

```bash
# Check your public IP (should show Italy server's IP)
curl ifconfig.me

# Or use a more detailed check
curl ipinfo.io
```

You should see your Italy server's IP and location!

### Disconnect from VPN

Press `Ctrl+C` in the client terminal. The client will:

- Clean up routes
- Restore normal internet connection
- Exit gracefully

---

## 🔧 Troubleshooting

### Issue: "Permission denied" on client

**Solution:** Run with sudo:

```bash
sudo ./cipherwall-client -server YOUR_SERVER_IP:1194
```

### Issue: Cannot connect to server

**Checks:**

1. Verify server is running (check Dokploy logs)
2. Verify firewall allows UDP 1194:
   ```bash
   # On server
   sudo ss -ulnp | grep 1194
   ```
3. Check cloud provider firewall/security groups
4. Test with netcat:

   ```bash
   # On server
   nc -u -l 1194

   # On client
   nc -u YOUR_SERVER_IP 1194
   ```

### Issue: "HMAC verification failed"

**Solution:** PSK mismatch! Ensure client and server have identical PSK values.

### Issue: Connected but no internet

**Checks:**

1. Verify server NAT is configured:
   ```bash
   # On server (SSH in)
   sudo iptables -t nat -L -n -v | grep 10.8.0
   ```
2. Check IP forwarding:
   ```bash
   # On server
   sysctl net.ipv4.ip_forward
   # Should show: net.ipv4.ip_forward = 1
   ```
3. Manually run setup script:
   ```bash
   # On server
   sudo ./setup-server.sh
   ```

### Issue: Docker container won't start

**Checks:**

1. Ensure privileged mode is enabled in Dokploy
2. Verify `/dev/net/tun` device is added
3. Check Dokploy logs for errors
4. Try restarting the container

---

## 📊 Monitoring

### View Server Logs (Dokploy)

In Dokploy dashboard:

- Go to your `cipherwall-vpn` application
- Click "Logs"
- Watch for connection messages

### View Client Status

The client shows real-time packet statistics:

```
📥 Received: 1234 bytes encrypted -> 1200 bytes decrypted
📤 Sent: 567 bytes plaintext -> 615 bytes encrypted
```

---

## 🔒 Security Best Practices

### Production Checklist:

- [ ] **Changed PSK** from default value
- [ ] **PSK is 32 bytes** exactly
- [ ] **PSK stored securely** (password manager, encrypted notes)
- [ ] **Firewall configured** to only allow UDP 1194
- [ ] **Server logs monitored** for suspicious activity
- [ ] **Client-side DNS leak protection** enabled
- [ ] **Regular updates** applied to server

### Advanced Security:

```bash
# On server: Limit connections to specific IPs
sudo iptables -I INPUT -p udp --dport 1194 ! -s YOUR_CLIENT_IP -j DROP

# Enable fail2ban for additional protection
sudo apt install fail2ban
```

---

## 🌐 Advanced: Multiple Clients

To support multiple clients simultaneously:

1. Each client gets the same PSK
2. Server automatically tracks client addresses
3. No additional configuration needed

**Note:** Current implementation uses a simple "default" client mapping. For production with many clients, you'd want to implement proper client session management.

---

## 🚀 Quick Start Summary

**Server (Italy - Dokploy):**

```bash
1. Push code to GitHub
2. Create Dokploy app with Dockerfile.dokploy
3. Enable: host network, privileged mode, /dev/net/tun
4. Deploy and verify logs
5. Open UDP port 1194 in firewall
```

**Client (Your Laptop):**

```bash
git clone <your-repo>
cd conduit
# Edit client.go - update PSK
go build -o cipherwall-client client.go
sudo ./cipherwall-client -server ITALY_SERVER_IP:1194
# Test: curl ifconfig.me
```

---

## 📞 Support

If you encounter issues:

1. Check server logs in Dokploy
2. Verify firewall rules on both ends
3. Ensure PSK matches exactly
4. Test basic connectivity with netcat
5. Review this troubleshooting guide

---

## 🎉 Success!

Once connected, you're browsing the internet through Italy! 🇮🇹

All your traffic is:

- ✅ Encrypted with AES-256
- ✅ Routed through your Italy server
- ✅ Appearing with Italy IP address

Enjoy your secure, private internet access! 🛡️
