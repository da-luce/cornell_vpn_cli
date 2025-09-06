#!/usr/bin/env python3
import argparse
import platform
import subprocess
import getpass
import os
import shlex
import getpass
import sys


def find_vpn_exec() -> str:
    """
    Attempt to locate the Cisco Secure Client (AnyConnect/Secure Connect) VPN executable
    depending on the OS. Falls back to PATH lookup if unknown.
    """
    os_type = platform.system()

    # macOS
    if os_type == "Darwin":
        candidates = [
            "/opt/cisco/secureclient/bin/vpn",
            "/Applications/Cisco/Cisco Secure Client.app/Contents/MacOS/vpn",
            "/Applications/Cisco AnyConnect Secure Mobility Client.app/Contents/MacOS/vpn",
        ]
    # Linux
    elif os_type == "Linux":
        candidates = [
            "/opt/cisco/secureclient/bin/vpn",
            "/opt/cisco/anyconnect/bin/vpn",
            "/usr/local/bin/vpn",
            "/usr/bin/vpn",
        ]
    # Windows
    elif os_type == "Windows":
        candidates = [
            r"C:\Program Files (x86)\Cisco\Cisco Secure Client\vpncli.exe",
            r"C:\Program Files (x86)\Cisco\Cisco AnyConnect Secure Mobility Client\vpncli.exe",
            r"C:\Program Files\Cisco\Cisco Secure Client\vpncli.exe",
            r"C:\Program Files\Cisco\Cisco AnyConnect Secure Mobility Client\vpncli.exe",
        ]
    else:
        candidates = []

    # Check each candidate
    for path in candidates:
        if os.path.exists(path) and os.access(path, os.X_OK):
            return path

    # Fallback: try PATH lookup
    path_in_path = shutil.which("vpn") or shutil.which("vpncli")
    if path_in_path:
        return path_in_path

    raise FileNotFoundError(
        "Could not locate Cisco Secure Client/AnyConnect executable."
    )


def run(cmd: str) -> str:
    try:
        return subprocess.check_output(shlex.split(cmd), text=True).strip()
    except subprocess.CalledProcessError:
        return ""


def get_os():
    return platform.system()


def get_ssid():
    os_type = get_os()
    if os_type == "Darwin":
        # macOS airport / wdutil
        ssid = run(
            "/System/Library/PrivateFrameworks/Apple80211.framework/Resources/airport -I"
        )
        if "SSID" in ssid:
            return ssid.split(" SSID: ")[1].splitlines()[0].strip()
        try:
            return run("sudo wdutil info").split("SSID:")[1].strip()
        except:
            return ""
    elif os_type == "Linux":
        return run("nmcli -t -f active,ssid dev wifi | awk -F: '$1==\"yes\"{print $2}'")
    return ""


def vpn_connected(vpn_exec: str):
    return "Connected" in run(f"{vpn_exec} status")


def connect_vpn(vpn_exec: str, host: str, username: str, method: str = "push"):
    if vpn_connected(vpn_exec):
        print("Error: VPN is already connected.", file=sys.stderr)
        sys.exit(1)

    password = getpass.getpass("Enter VPN password: ")
    script = f"connect {host}\n{username}\n{password}\n{method}\ny\nexit\n"
    proc = subprocess.Popen([vpn_exec, "-s"], stdin=subprocess.PIPE, text=True)
    proc.communicate(script)

    if vpn_connected(vpn_exec):
        return True
    else:
        print("Error: VPN connection failed.", file=sys.stderr)
        sys.exit(1)


def disconnect_vpn(vpn_exec: str) -> bool:
    if not vpn_connected(vpn_exec):
        print("Error: VPN is not connected.", file=sys.stderr)
        sys.exit(1)

    # Simple disconnect command
    proc = subprocess.Popen([vpn_exec, "-s"], stdin=subprocess.PIPE, text=True)
    proc.communicate("disconnect\nexit\n")

    if not vpn_connected(vpn_exec):
        return True
    else:
        print("Error: VPN disconnection failed.", file=sys.stderr)
        sys.exit(1)


def main():
    parser = argparse.ArgumentParser(
        description="CLI wrapper around Cisco Secure Client"
    )
    subparsers = parser.add_subparsers(dest="command")

    # Connect subcommand (default)
    parser_connect = subparsers.add_parser("connect", help="Connect to VPN")
    parser_connect.add_argument("--username", required=True, help="Your VPN username")
    parser_connect.add_argument("--vpn-host", required=True, help="VPN URL")
    parser_connect.add_argument("--vpn-exec", help="Path to VPN executable")
    parser_connect.add_argument("--method", default=os.getenv("VPN_METHOD", "push"))

    # Disconnect subcommand
    parser_disconnect = subparsers.add_parser("disconnect", help="Disconnect from VPN")
    parser_disconnect.add_argument("--vpn-exec", help="Path to VPN executable")

    # Status subcommand
    parser_status = subparsers.add_parser("status", help="Show VPN connection status")
    parser_status.add_argument("--vpn-exec", help="Path to VPN executable")

    args = parser.parse_args()

    # Determine VPN executable
    vpn_exec = getattr(args, "vpn_exec", None)
    if not vpn_exec:
        try:
            vpn_exec = find_vpn_exec()  # your OS-based auto-detection
        except FileNotFoundError as e:
            print(f"Error: {e}", file=sys.stderr)
            sys.exit(1)

    if args.command == "status":
        connected = vpn_connected(vpn_exec)
        print(f"VPN Connected: {'Yes' if connected else 'No'}")
    elif args.command == "connect":
        success = connect_vpn(vpn_exec, args.vpn_host, args.username, args.method)
        print(f"VPN connection {'successful' if success else 'failed'}")
    elif args.command == "disconnect":
        success = disconnect_vpn(vpn_exec)
        print(f"VPN disconnection {'successful' if success else 'failed'}")
    else:
        raise ValueError(f"Unrecognized command: {args.command}")


if __name__ == "__main__":
    main()
