<div align="center">

<img src="logo.png" alt="V2bX" width="100%">

<br>

### Multi-Core Multi-Protocol Node Backend Powered by Xray-core · Sing-box

**Built for XBoard / V2Board — Native Reality, XTLS-Vision & Modern Transport Support**

<br>

[![Release](https://img.shields.io/github/v/release/Sakawat-hossain/V2bX?style=for-the-badge&label=RELEASE&labelColor=0B0E14&color=7C3AED)](https://github.com/Sakawat-hossain/V2bX/releases)
[![Go](https://img.shields.io/badge/GO-1.26-0AB2F9?style=for-the-badge&labelColor=0B0E14&logo=go&logoColor=white)](go.mod)
[![License](https://img.shields.io/badge/LICENSE-MPL--2.0-A16FEE?style=for-the-badge&labelColor=0B0E14)](LICENSE)

<br>

📖 Primary documentation: **[简体中文（主文档）](README.zh-CN.md)** · **English** (this page)

</div>


# V2bX

**V2bX** is a high-performance, multi-protocol VPN node backend designed for **XBoard**, **V2Board**, and other **UniProxy-compatible panels**.

Powered by **Xray-core**, **sing-box**, and **Hysteria2** engines, V2bX provides a modern, scalable node architecture with support for advanced proxy technologies including **Reality**, **XTLS-Vision**, **Hysteria2**, and high-performance UDP transport.

Built for next-generation VPN infrastructure with:

- Multi-core processing
- Advanced transport protocols
- Anti-censorship technologies
- High-speed UDP acceleration
- Enterprise-level node management

---

## 🚀 Protocol Feature Support

V2bX provides unified management and advanced control features across all supported protocols.

| Feature | VLESS | VMess | Trojan | Shadowsocks | Hysteria | Hysteria2 | TUIC | SOCKS | HTTP | NaiveProxy | Mieru | AnyTLS |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| Automatic TLS Certificate Management | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Online User Statistics | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Routing Rules | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Custom DNS Support | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| IP Limit | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Connection Limit | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Cross-Node IP Limit | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| User-Level Speed Limit | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| UDP Relay Support | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ✅ | ✅ | ✅ |
| Multi-Core Engine Support | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |

---

# ✨ Features

| Feature | Description |
|---|---|
| ⚡ **Multi-Core Architecture** | Supports **Xray-core**, **sing-box**, and **Hysteria2** engines |
| 🔐 **Reality Support** | Native VLESS Reality with **XTLS-Vision support** |
| 🌐 **Multi Protocol Engine** | Run different protocols using the best backend engine |
| 🛰️ **Panel Integration** | Full XBoard / V2Board UniProxy API support |
| 👥 **User Management** | Automatic user synchronization from panel |
| 📊 **Traffic Reporting** | Real-time traffic and online device reporting |
| 🚀 **UDP Acceleration** | High-performance UDP relay support |
| 🧩 **Multi Architecture** | amd64 / arm64 / armv7 / armv6 / armv5 / s390x / riscv64 |
| 🔒 **Certificate Support** | none / self / http / dns / reality |
| ⚙️ **Production Ready** | Designed for large-scale VPN deployment |

---

# 🛰️ Panel Compatibility

Compatible with:

- **XBoard**
- **V2Board**
- **UniProxy compatible panels**

Supported functions:

✅ Node configuration sync  
✅ User synchronization  
✅ Traffic statistics  
✅ Online device reporting  
✅ Automatic configuration updates  

## 🚀 Install

One-line script Supports for Ubuntu, Debian, CentOS, Alpine, and Arch; supports amd64, arm64, armv7, armv6, armv5, s390x, and riscv64 

```bash
wget -N https://raw.githubusercontent.com/Sakawat-hossain/V2bX/main/V2bX-script-master/install.sh && bash install.sh
```

After install, run `V2bX` for the management menu, or use the commands below. You can also build from source (see the bottom of this page).

## ⚙️ Configuration

The default config file is `/etc/V2bX/config.json`; see [`example/config.json`](example/config.json) for a full example. The shape (keys are case-sensitive):

```jsonc
{
  "Log": { "Level": "info", "Output": "" },
  "Cores": [
    {
      "Type": "sing",
      "Log": { "Level": "info", "Timestamp": true },
      "OriginalPath": "/etc/V2bX/sing_origin.json"
    }
  ],
  "Nodes": [
    {
      "Core": "sing",
      "ApiHost": "https://panel.example.com",
      "ApiKey": "panel node communication key / token",
      "NodeID": 1,
      "NodeType": "vless",
      "ListenIP": "0.0.0.0",
      "EnableSniff": true,
      "CertConfig": { "CertMode": "none" }
    }
  ]
}
```

- **`Cores`** — the cores you enable (`sing` / `xray` / `hysteria2`); each node picks one via its `Core` field.
- **`NodeType`** — `shadowsocks` / `vless` / `vmess` / `trojan` / `hysteria2` / `socks`, matching the panel.
- **`CertConfig.CertMode`** — `none` / `self` / `http` / `dns` / `reality`.
- **VLESS-Reality / XTLS-Vision** — configured on the node in the panel (dest, keys, short ID, flow) and handled by the core; the agent reads it from the panel config, so you don't hand-write keys.

## 🛠️ Management Commands

Once installed, use `v2bx` or `V2bX` from your terminal to manage the node service.

| Command | Description |
|---|---|
| `v2bx` | Open interactive management menu |
| `v2bx start` | Start V2bX service |
| `v2bx stop` | Stop V2bX service |
| `v2bx restart` | Restart V2bX service |
| `v2bx status` | Check service status |
| `v2bx log` | View live logs |
| `v2bx update` | Update to latest version |
| `v2bx update <version>` | Update to a specific version |
| `v2bx generate` | Generate configuration interactively |
| `v2bx x25519` | Generate X25519 key pair for Reality |
| `v2bx enable` | Enable startup on boot |
| `v2bx disable` | Disable startup on boot |
| `v2bx uninstall` | Remove V2bX completely |
| `v2bx version` | Display installed version |

## 🧱 Build from source

```bash
# needs Go 1.26+; build tags and GOEXPERIMENT must match the Dockerfile
GOEXPERIMENT=jsonv2 go build -trimpath \
  -tags "sing xray hysteria2 with_quic with_grpc with_utls with_wireguard with_acme with_gvisor" \
  -o V2bX .
```

At runtime it needs the geo data files (`geoip.dat` / `geosite.dat`, and optionally `geoip.db` / `geosite.db`) — fetched by the install script or at deploy time, not shipped in the repo.

## 👤 Maintainer

**Sakawat Hossain**

## 📄 License

Licensed under the [Mozilla Public License 2.0](LICENSE).
