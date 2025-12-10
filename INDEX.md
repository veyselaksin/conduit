# 📚 CipherWall VPN - Documentation Index

Welcome to CipherWall VPN! This index will help you find the right documentation.

---

## 🎯 Quick Navigation

### I want to...

**...get started quickly**
→ Read [QUICKSTART.md](QUICKSTART.md)

**...deploy to my Italy server with Dokploy**
→ Read [DEPLOYMENT.md](DEPLOYMENT.md)

**...understand how to use it**
→ Read [USAGE.md](USAGE.md)

**...learn how it works**
→ Read [ARCHITECTURE.md](ARCHITECTURE.md)

**...see an overview of the project**
→ Read [SUMMARY.md](SUMMARY.md)

**...understand the main features**
→ Read [README.md](README.md)

---

## 📖 Documentation Structure

### For End Users

| Document                       | Purpose                    | When to Read        |
| ------------------------------ | -------------------------- | ------------------- |
| [README.md](README.md)         | Project overview, features | First time          |
| [QUICKSTART.md](QUICKSTART.md) | Quick reference card       | Need quick commands |
| [DEPLOYMENT.md](DEPLOYMENT.md) | Complete deployment guide  | Ready to deploy     |
| [USAGE.md](USAGE.md)           | How to use the VPN         | After deployment    |

### For Developers

| Document                           | Purpose                | When to Read              |
| ---------------------------------- | ---------------------- | ------------------------- |
| [SUMMARY.md](SUMMARY.md)           | Technical overview     | Understanding the project |
| [ARCHITECTURE.md](ARCHITECTURE.md) | System architecture    | Deep dive into design     |
| Source code                        | Implementation details | Contributing/modifying    |

---

## 🗂️ File Reference

### Documentation Files

- **README.md** - Main project documentation with features and setup
- **DEPLOYMENT.md** - Step-by-step Dokploy deployment guide
- **USAGE.md** - Complete usage guide with FAQ
- **QUICKSTART.md** - Quick reference for common commands
- **SUMMARY.md** - Project summary and capabilities
- **ARCHITECTURE.md** - System architecture and diagrams
- **INDEX.md** - This file!

### Source Code

- **main.go** - VPN Server implementation (bidirectional)
- **client.go** - VPN Client implementation

### Configuration & Deployment

- **Dockerfile** - Standard Docker image
- **Dockerfile.dokploy** - Dokploy-optimized Docker image
- **docker-compose.yml** - Docker Compose configuration
- **setup-server.sh** - Server setup script (NAT, routing)
- **Makefile** - Build automation
- **.env.example** - Environment configuration template

### Auxiliary Files

- **client_example.py** - Python client example (legacy)
- **go.mod** / **go.sum** - Go dependencies
- **.gitignore** - Git ignore rules

---

## 🚀 Getting Started Paths

### Path 1: Quick Test (Local Network)

1. Read: [README.md](README.md)
2. Build: `make all`
3. Run server: `sudo ./cipherwall-server`
4. Run client: `sudo ./cipherwall-client -server localhost:1194`

### Path 2: Deploy to Production (Dokploy)

1. Read: [QUICKSTART.md](QUICKSTART.md) - Overview
2. Read: [DEPLOYMENT.md](DEPLOYMENT.md) - Full guide
3. Deploy using Dokploy
4. Read: [USAGE.md](USAGE.md) - Connect and use

### Path 3: Learn the Internals

1. Read: [SUMMARY.md](SUMMARY.md) - Overview
2. Read: [ARCHITECTURE.md](ARCHITECTURE.md) - Design
3. Review: Source code (`main.go`, `client.go`)
4. Experiment: Modify and test

---

## 🎓 Learning Sequence

### Beginner (Just want to use it)

1. ✅ [README.md](README.md) - What is CipherWall?
2. ✅ [QUICKSTART.md](QUICKSTART.md) - Key commands
3. ✅ [DEPLOYMENT.md](DEPLOYMENT.md) - How to deploy
4. ✅ [USAGE.md](USAGE.md) - How to use

### Intermediate (Want to understand it)

1. ✅ [SUMMARY.md](SUMMARY.md) - Project capabilities
2. ✅ [ARCHITECTURE.md](ARCHITECTURE.md) - How it works
3. ✅ Source code review
4. ✅ Experiment with modifications

### Advanced (Want to extend it)

1. ✅ Deep dive into source code
2. ✅ Study crypto implementation
3. ✅ Review networking code
4. ✅ Implement new features

---

## 🔍 Find Information By Topic

### Deployment

- [DEPLOYMENT.md](DEPLOYMENT.md) - Full Dokploy deployment guide
- [QUICKSTART.md](QUICKSTART.md) - Quick deployment reference
- [Dockerfile.dokploy](Dockerfile.dokploy) - Dokploy Docker image

### Configuration

- [README.md](README.md) - Configuration options
- [.env.example](.env.example) - Environment variables
- Source code constants

### Security

- [SUMMARY.md](SUMMARY.md) - Security features
- [USAGE.md](USAGE.md) - Security best practices
- [ARCHITECTURE.md](ARCHITECTURE.md) - Encryption flow

### Troubleshooting

- [DEPLOYMENT.md](DEPLOYMENT.md) - Deployment issues
- [USAGE.md](USAGE.md) - Common problems FAQ
- [QUICKSTART.md](QUICKSTART.md) - Quick fixes

### Architecture

- [ARCHITECTURE.md](ARCHITECTURE.md) - Complete diagrams
- [SUMMARY.md](SUMMARY.md) - Technical overview
- Source code comments

---

## 📞 Support Resources

### Quick Help

1. Check [QUICKSTART.md](QUICKSTART.md) for commands
2. Review [USAGE.md](USAGE.md) FAQ section
3. See [DEPLOYMENT.md](DEPLOYMENT.md) troubleshooting

### Understanding Issues

1. Review [ARCHITECTURE.md](ARCHITECTURE.md) for flow diagrams
2. Check source code for implementation details
3. Test with verbose logging

### Common Questions

- **"How do I deploy?"** → [DEPLOYMENT.md](DEPLOYMENT.md)
- **"How do I connect?"** → [USAGE.md](USAGE.md)
- **"How does it work?"** → [ARCHITECTURE.md](ARCHITECTURE.md)
- **"What are the commands?"** → [QUICKSTART.md](QUICKSTART.md)

---

## 🎯 Cheat Sheet

### Quick Commands

```bash
# Build
make all              # Build server and client
make server           # Build server only
make client           # Build client only

# Run
sudo ./cipherwall-server                        # Start server
sudo ./cipherwall-client -server IP:1194       # Connect client

# Deploy
docker-compose up -d                            # Docker deployment
# OR use Dokploy (see DEPLOYMENT.md)

# Verify
curl ifconfig.me      # Check your IP
ip addr show tun0     # Check VPN interface
```

### Key Files to Edit

```bash
main.go              # Change PSK (line 22)
client.go            # Change PSK (line 22)
setup-server.sh      # Modify NAT rules
Dockerfile.dokploy   # Customize Docker image
```

---

## 📚 Documentation Changelog

- **v1.0** - Initial complete documentation
  - Added comprehensive guides
  - Created architecture diagrams
  - Included troubleshooting sections
  - Added this index

---

## 🎉 Ready to Start?

1. **New user?** Start with [README.md](README.md)
2. **Want to deploy?** Go to [DEPLOYMENT.md](DEPLOYMENT.md)
3. **Need quick help?** Check [QUICKSTART.md](QUICKSTART.md)
4. **Want to learn?** Read [ARCHITECTURE.md](ARCHITECTURE.md)

---

**Happy VPN'ing! 🛡️**

_All documentation is in Markdown format for easy reading in GitHub or any text editor._
