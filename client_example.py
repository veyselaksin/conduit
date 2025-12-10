# Example VPN Client (Python)
# This is a reference implementation for testing the CipherWall server

import socket
import hashlib
import hmac
from Crypto.Cipher import AES
from Crypto.Random import get_random_bytes
import struct

# Configuration - MUST match server settings
PSK = b"this-is-a-strong-32byte-secret-"  # Exactly 32 bytes
SERVER_IP = "127.0.0.1"  # Change to server IP
UDP_PORT = 1194
PBKDF2_SALT = b"cipherwall-salt-2025"
PBKDF2_ITERATIONS = 100000

def derive_keys(psk):
    """Derive AES and HMAC keys using PBKDF2"""
    master_key = hashlib.pbkdf2_hmac('sha256', psk, PBKDF2_SALT, PBKDF2_ITERATIONS, dklen=64)
    aes_key = master_key[:32]
    hmac_key = master_key[32:]
    return aes_key, hmac_key

def encrypt_packet(plaintext, aes_key):
    """Encrypt using AES-256 CFB mode"""
    iv = get_random_bytes(16)
    cipher = AES.new(aes_key, AES.MODE_CFB, iv, segment_size=128)
    ciphertext = cipher.encrypt(plaintext)
    return iv + ciphertext

def add_hmac(data, hmac_key):
    """Add HMAC-SHA256 authentication tag"""
    h = hmac.new(hmac_key, data, hashlib.sha256)
    return h.digest() + data

def send_vpn_packet(sock, ip_packet, aes_key, hmac_key):
    """Encrypt, authenticate, and send IP packet"""
    # Encrypt
    encrypted = encrypt_packet(ip_packet, aes_key)
    # Add HMAC
    authenticated = add_hmac(encrypted, hmac_key)
    # Send
    sock.sendto(authenticated, (SERVER_IP, UDP_PORT))

# Example usage
if __name__ == "__main__":
    print("🔐 Deriving keys...")
    aes_key, hmac_key = derive_keys(PSK)
    print(f"✅ AES Key: {len(aes_key)} bytes")
    print(f"✅ HMAC Key: {len(hmac_key)} bytes")
    
    # Create UDP socket
    sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    
    # Example: Send an ICMP Echo Request (ping) packet
    # This is a simplified example - real IP packets have headers
    # IP Header (20 bytes) + ICMP Header (8 bytes) + Data
    
    print(f"\n📡 Sending test packet to {SERVER_IP}:{UDP_PORT}")
    
    # Simple test payload (not a real IP packet, just for testing encryption)
    test_payload = b"Hello CipherWall VPN Server!"
    
    send_vpn_packet(sock, test_payload, aes_key, hmac_key)
    print("✅ Packet sent successfully!")
    
    sock.close()
