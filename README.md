# seccli

`seccli` is a lightweight CLI wrapper around [Cisco Secure Client](https://www.cisco.com/site/us/en/products/security/secure-client/index.html) (formerly AnyConnect) for managing VPN connections.

It provides a simpler interface to connect, disconnect, and check the status of your VPN.

## Why?

Cisco Secure Client (AnyConnect) is primarily accessed through a GUI. While it has a CLI, it is often cumbersome, non-intuitive, and hard to automate (e.g., requiring awkward manual input piping).

`seccli` wraps the existing CLI, providing a simple, cross-platform, terminal-friendly interface.

## Installation

### From Source

1. Make sure you have [Go](https://golang.org/dl/) installed (version 1.16 or later)
2. Clone this repository
3. Build the binary:
   ```bash
   go build -o seccli
   ```
4. Optionally, move the binary to your PATH:
   ```bash
   # macOS/Linux
   sudo mv seccli /usr/local/bin/
   
   # Or add to your PATH in ~/.bashrc or ~/.zshrc
   export PATH=$PATH:/path/to/seccli
   ```

## Usage

This tool was primarily created to simplify connecting to [Cornell's VPN](https://it.cornell.edu/cuvpn), but it can be used with any Cisco Secure Client VPN.

```bash
# Connect to VPN
./seccli connect --username myNetID --vpn-host cuvpn.cuvpn.cornell.edu

# Connect with specific authentication method
./seccli connect --username myNetID --vpn-host cuvpn.cuvpn.cornell.edu --method push

# Connect with verbose output (shows VPN tool output)
./seccli connect --username myNetID --vpn-host cuvpn.cuvpn.cornell.edu --verbose

# Disconnect from VPN
./seccli disconnect

# Disconnect with verbose output
./seccli disconnect --verbose

# Check VPN status
./seccli status

# Show help
./seccli --help
```

### Environment Variables

You can set the default authentication method using the `VPN_METHOD` environment variable:

```bash
export VPN_METHOD=push
./seccli connect --username myNetID --vpn-host cuvpn.cuvpn.cornell.edu
```

### Custom VPN Executable Path

If the tool cannot auto-detect your Cisco Secure Client installation, you can specify the path manually:

```bash
./seccli status --vpn-exec /path/to/vpn
```

## Requirements

- [Cisco Secure Client](https://www.cisco.com/site/us/en/products/security/secure-client/index.html) (formerly AnyConnect) must be installed
- The tool automatically detects the VPN executable on:
  - **macOS**: `/opt/cisco/secureclient/bin/vpn`, `/Applications/Cisco/Cisco Secure Client.app/Contents/MacOS/vpn`, or `/Applications/Cisco AnyConnect Secure Mobility Client.app/Contents/MacOS/vpn`
  - **Linux**: `/opt/cisco/secureclient/bin/vpn`, `/opt/cisco/anyconnect/bin/vpn`, `/usr/local/bin/vpn`, or `/usr/bin/vpn`
  - **Windows**: `C:\Program Files (x86)\Cisco\Cisco Secure Client\vpncli.exe` or `C:\Program Files (x86)\Cisco\Cisco AnyConnect Secure Mobility Client\vpncli.exe`

## Development

This project is written in Go and uses only standard library packages plus `golang.org/x/term` for secure password input.

### Building

```bash
go mod tidy
go build -o seccli
```

### Testing

```bash
go test ./...
```
