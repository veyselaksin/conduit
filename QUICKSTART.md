# Quick Reference Card

## 🚀 Server Setup (Dokploy)

1. **Update PSK in main.go**

   ```go
   PSK = "your-32byte-secure-key-here!"
   ```

2. **Deploy in Dokploy**

   - Dockerfile: `Dockerfile.dokploy`
   - Network: `host`
   - Privileged: ✅ YES
   - Devices: `/dev/net/tun:/dev/net/tun`
   - Capabilities: `NET_ADMIN`, `NET_RAW`

3. **Open Firewall**
   ```bash
   sudo ufw allow 1194/udp
   ```

---

## 💻 Client Connection

1. **Update PSK in client.go** (must match server)

2. **Build**

   ```bash
   go build -o cipherwall-client client.go
   ```

3. **Connect**

   ```bash
   sudo ./cipherwall-client -server SERVER_IP:1194
   ```

4. **Verify**
   ```bash
   curl ifconfig.me  # Should show server's IP
   ```

---

## 🔧 Troubleshooting

### Can't connect?

```bash
# Check server is listening
sudo ss -ulnp | grep 1194

# Check firewall
sudo iptables -L -n | grep 1194
```

### HMAC verification failed?

- PSK mismatch! Ensure client and server have identical PSK.

### Connected but no internet?

```bash
# On server, run setup:
sudo ./setup-server.sh

# Verify NAT:
sudo iptables -t nat -L -n -v | grep 10.8.0
```

---

## 📝 Files

- `main.go` - VPN Server
- `client.go` - VPN Client
- `setup-server.sh` - Server NAT/routing setup
- `Dockerfile.dokploy` - Dokploy deployment
- `DEPLOYMENT.md` - Full deployment guide

---

## 🔑 Security Checklist

- [ ] Changed PSK from default
- [ ] PSK is exactly 32 bytes
- [ ] Firewall allows only UDP 1194
- [ ] Client PSK matches server PSK
- [ ] Server has IP forwarding enabled

---

## 📊 Port Forwarding

**Cloud Providers:**

- AWS: Edit Security Group → Add UDP 1194
- DigitalOcean: Networking → Firewall → Add UDP 1194
- Hetzner: Firewalls → Add UDP 1194

**Server Firewall:**

```bash
# UFW
sudo ufw allow 1194/udp

# firewalld
sudo firewall-cmd --permanent --add-port=1194/udp
sudo firewall-cmd --reload

# iptables
sudo iptables -I INPUT -p udp --dport 1194 -j ACCEPT
```
