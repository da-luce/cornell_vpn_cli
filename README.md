# seccli

`seccli` is a lightweight CLI wrapper around [Cisco Secure Client](https://www.cisco.com/site/us/en/products/security/secure-client/index.html) (formerly AnyConnect) for managing VPN connections.

It provides a simpler interface to connect, disconnect, and check the status of your VPN.

---

## Why?

Cisco Secure Client (AnyConnect) is primarily accessed through a GUI. While it has a CLI, it is often cumbersome, non-intuitive, and hard to automat (e.g., requiring awkward manual input piping).

`seccli` wraps the existing CLI, providing a simple, cross-platform, terminal-friendly interface.

---

## Installation

Currently, just run `main.py`

---

## Usage

This tool was primarily created to simplify connecting to [Cornell's VPN](https://it.cornell.edu/cuvpn), but it can be used with any Cisco Secure Client VPN.

```shell
# Connect to VPN
python3 main.py connect --username myNetID --vpn-host cuvpn.cuvpn.cornell.edu

# Disconnect from VPN
python3 main.py disconnect

# Check VPN status
python3 main.py status
```
