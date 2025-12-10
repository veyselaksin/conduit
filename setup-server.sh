#!/bin/bash
# CipherWall Server Setup Script
# This script configures the server to act as a VPN gateway with NAT

set -e

echo "🛡️  CipherWall VPN Server Setup"
echo "================================"

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
    echo "❌ Please run as root (use sudo)"
    exit 1
fi

echo "📦 Step 1: Enabling IP forwarding..."
# Enable IP forwarding temporarily
sysctl -w net.ipv4.ip_forward=1

# Make it persistent across reboots
if ! grep -q "net.ipv4.ip_forward=1" /etc/sysctl.conf; then
    echo "net.ipv4.ip_forward=1" >> /etc/sysctl.conf
    echo "✅ IP forwarding enabled and made persistent"
else
    echo "✅ IP forwarding already configured in /etc/sysctl.conf"
fi

echo ""
echo "🔥 Step 2: Configuring iptables NAT..."

# Get the default network interface (the one with internet access)
DEFAULT_INTERFACE=$(ip route | grep default | awk '{print $5}' | head -n 1)

if [ -z "$DEFAULT_INTERFACE" ]; then
    echo "❌ Could not detect default network interface"
    echo "Please manually specify your internet-facing interface:"
    echo "Example: export DEFAULT_INTERFACE=eth0"
    exit 1
fi

echo "📡 Detected internet interface: $DEFAULT_INTERFACE"

# Add NAT rule for VPN traffic
iptables -t nat -C POSTROUTING -s 10.8.0.0/24 -o $DEFAULT_INTERFACE -j MASQUERADE 2>/dev/null || \
    iptables -t nat -A POSTROUTING -s 10.8.0.0/24 -o $DEFAULT_INTERFACE -j MASQUERADE

# Allow forwarding for VPN subnet
iptables -C FORWARD -s 10.8.0.0/24 -j ACCEPT 2>/dev/null || \
    iptables -A FORWARD -s 10.8.0.0/24 -j ACCEPT

iptables -C FORWARD -d 10.8.0.0/24 -j ACCEPT 2>/dev/null || \
    iptables -A FORWARD -d 10.8.0.0/24 -j ACCEPT

echo "✅ iptables NAT configured"

echo ""
echo "💾 Step 3: Saving iptables rules..."

# Save iptables rules (method depends on distro)
if command -v netfilter-persistent &> /dev/null; then
    netfilter-persistent save
    echo "✅ Rules saved using netfilter-persistent"
elif command -v iptables-save &> /dev/null; then
    if [ -d /etc/iptables ]; then
        iptables-save > /etc/iptables/rules.v4
        echo "✅ Rules saved to /etc/iptables/rules.v4"
    else
        mkdir -p /etc/iptables
        iptables-save > /etc/iptables/rules.v4
        echo "✅ Rules saved to /etc/iptables/rules.v4"
    fi
else
    echo "⚠️  Warning: Could not find iptables-save. Rules may not persist after reboot."
    echo "Current rules:"
    iptables -t nat -L -n -v
fi

echo ""
echo "🔒 Step 4: Opening UDP port 1194..."

# Check if firewall is active and configure it
if command -v ufw &> /dev/null && ufw status | grep -q "Status: active"; then
    ufw allow 1194/udp
    echo "✅ UFW: Allowed UDP port 1194"
elif command -v firewall-cmd &> /dev/null && systemctl is-active --quiet firewalld; then
    firewall-cmd --permanent --add-port=1194/udp
    firewall-cmd --reload
    echo "✅ firewalld: Allowed UDP port 1194"
else
    # Add iptables rule if no high-level firewall is detected
    iptables -C INPUT -p udp --dport 1194 -j ACCEPT 2>/dev/null || \
        iptables -I INPUT -p udp --dport 1194 -j ACCEPT
    echo "✅ iptables: Allowed UDP port 1194"
fi

echo ""
echo "✅ Setup Complete!"
echo ""
echo "📋 Configuration Summary:"
echo "  - IP Forwarding: Enabled"
echo "  - NAT Interface: $DEFAULT_INTERFACE"
echo "  - VPN Subnet: 10.8.0.0/24"
echo "  - UDP Port: 1194 (open)"
echo ""
echo "🚀 You can now start the CipherWall server:"
echo "   sudo ./cipherwall-server"
echo ""
echo "📝 Current iptables NAT rules:"
iptables -t nat -L POSTROUTING -n -v | grep 10.8.0
echo ""
