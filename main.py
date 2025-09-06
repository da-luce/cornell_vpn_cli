#!/usr/bin/env python3
import argparse
import platform
import subprocess
import getpass
import os
import shlex
import getpass
import sys

VPN_HOST = "cuvpn.cuvpn.cornell.edu"

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
    return not vpn_connected(vpn_exec)

    if not vpn_connected(vpn_exec):
        return True
    else:
        print("Error: VPN disconnection failed.", file=sys.stderr)
        sys.exit(1)


def main():
    parser = argparse.ArgumentParser(description="Cornell VPN CLI Tool")
    subparsers = parser.add_subparsers(dest="command")

    # Connect subcommand (default)
    parser_connect = subparsers.add_parser("connect", help="Connect to VPN")
    parser_connect.add_argument(
        "--username", required=True, help="Your Cornell VPN username"
    )
    parser_connect.add_argument("--vpn-host", default=os.getenv("VPN_HOST", VPN_HOST))
    parser_connect.add_argument(
        "--vpn-exec", default=os.getenv("VPN_EXEC", "/opt/cisco/secureclient/bin/vpn")
    )
    parser_connect.add_argument("--method", default=os.getenv("VPN_METHOD", "push"))
    # Disconnect subcommand
    parser_disconnect = subparsers.add_parser("disconnect", help="Disconnect from VPN")
    parser_disconnect.add_argument(
        "--vpn-exec", default=os.getenv("VPN_EXEC", "/opt/cisco/secureclient/bin/vpn")
    )

    # Status subcommand
    parser_status = subparsers.add_parser("status", help="Show VPN connection status")
    parser_status.add_argument(
        "--vpn-exec", default=os.getenv("VPN_EXEC", "/opt/cisco/secureclient/bin/vpn")
    )

    args = parser.parse_args()
    command = args.command or "connect"

    if command == "status":
        connected = vpn_connected(args.vpn_exec)
        print(f"VPN Connected: {'Yes' if connected else 'No'}")
    elif command == "connect":
        success = connect_vpn(args.vpn_exec, args.vpn_host, args.username, args.method)
        print(f"VPN connection {'successful' if success else 'failed'}")
    elif command == "disconnect":
        success = disconnect_vpn(args.vpn_exec)
        print(f"VPN disconnection {'successful' if success else 'failed'}")


if __name__ == "__main__":
    main()
