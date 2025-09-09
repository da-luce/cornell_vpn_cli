package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/briandowns/spinner"
	"github.com/urfave/cli/v3"
	"golang.org/x/term"
)

// findVPNExec attempts to locate the Cisco Secure Client VPN executable
// depending on the OS. Falls back to PATH lookup if unknown.
func findVPNExec() (string, error) {
	osType := runtime.GOOS
	var candidates []string

	switch osType {
	case "darwin": // macOS
		candidates = []string{
			"/opt/cisco/secureclient/bin/vpn",
			"/Applications/Cisco/Cisco Secure Client.app/Contents/MacOS/vpn",
			"/Applications/Cisco AnyConnect Secure Mobility Client.app/Contents/MacOS/vpn",
		}
	case "linux":
		candidates = []string{
			"/opt/cisco/secureclient/bin/vpn",
			"/opt/cisco/anyconnect/bin/vpn",
			"/usr/local/bin/vpn",
			"/usr/bin/vpn",
		}
	case "windows":
		candidates = []string{
			`C:\Program Files (x86)\Cisco\Cisco Secure Client\vpncli.exe`,
			`C:\Program Files (x86)\Cisco\Cisco AnyConnect Secure Mobility Client\vpncli.exe`,
			`C:\Program Files\Cisco\Cisco Secure Client\vpncli.exe`,
			`C:\Program Files\Cisco\Cisco AnyConnect Secure Mobility Client\vpncli.exe`,
		}
	}

	// Check each candidate
	for _, path := range candidates {
		if fileExists(path) && isExecutable(path) {
			return path, nil
		}
	}

	// Fallback: try PATH lookup
	vpnExecs := []string{"vpn", "vpncli"}
	for _, execName := range vpnExecs {
		if path, err := exec.LookPath(execName); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("could not locate Cisco Secure Client/AnyConnect executable")
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// isExecutable checks if a file is executable
func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.Mode()&0111 != 0
}

// runCommand executes a command and returns its output
func runCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// vpnConnected checks if VPN is currently connected
func vpnConnected(vpnExec string) bool {
	output, err := runCommand(vpnExec, "status")
	if err != nil {
		return false
	}
	return strings.Contains(output, "Connected")
}

// getPassword prompts for password input without echoing
func getPassword(prompt string) (string, error) {
	fmt.Print(prompt)
	password, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println() // Add newline after password input
	if err != nil {
		return "", err
	}
	return string(password), nil
}

// connectVPN connects to the VPN
func connectVPN(vpnExec, host, username, method string, verbose bool) error {
	// Start spinner for connection process
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Checking VPN Status..."
	s.Start()
	defer s.Stop()

	if vpnConnected(vpnExec) {
		return fmt.Errorf("VPN is already connected")
	}

	s.Stop()

	password, err := getPassword("Enter VPN password: ")
	if err != nil {
		return fmt.Errorf("failed to read password: %v", err)
	}

	// FIXME: this text is interrupted by the Duo (push/sms/phone): thing
	// s = spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	// s.Suffix = " Connecting to VPN..."
	// s.Start()
	// defer s.Stop()

	// Create the script for VPN connection like Python version
	script := fmt.Sprintf("connect %s\n%s\n%s\n%s\ny\nexit\n", host, username, password, method)

	cmd := exec.Command(vpnExec, "-s")
	cmd.Stdin = strings.NewReader(script)

	if verbose {
		// s.Stop() // Stop spinner if verbose mode to show VPN output
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("VPN command failed: %v", err)
	}

	// Check if connection was successful
	if !vpnConnected(vpnExec) {
		return fmt.Errorf("VPN connection failed")
	}

	return nil
}

// disconnectVPN disconnects from the VPN
func disconnectVPN(vpnExec string, verbose bool) error {

	// FIXME: this code is duplicated
	// Start spinner for connection process
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Checking VPN Status..."
	s.Start()
	defer s.Stop()

	if !vpnConnected(vpnExec) {
		return fmt.Errorf("VPN is not connected.")
	}

	s.Stop()

	// Start spinner for connection process
	s = spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Disconnecting from VPN..."
	s.Start()
	defer s.Stop()

	script := "disconnect\nexit\n"
	cmd := exec.Command(vpnExec, "-s")
	cmd.Stdin = strings.NewReader(script)

	if verbose {
		s.Stop() // Stop spinner if verbose mode to show VPN output
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("VPN disconnect command failed: %v", err)
	}

	// Check if disconnection was successful
	if vpnConnected(vpnExec) {
		return fmt.Errorf("VPN disconnection failed")
	}

	return nil
}

// getVPNExec gets the VPN executable path from context or auto-detects it
func getVPNExec(cmd *cli.Command) (string, error) {
	vpnExec := cmd.String("vpn-exec")
	if vpnExec == "" {
		return findVPNExec()
	}
	return vpnExec, nil
}

// connectAction handles the connect command
func connectAction(ctx context.Context, cmd *cli.Command) error {
	username := cmd.String("username")
	vpnHost := cmd.String("vpn-host")
	method := cmd.String("method")
	verbose := cmd.Bool("verbose")

	if username == "" {
		return fmt.Errorf("--username is required for connect command")
	}
	if vpnHost == "" {
		return fmt.Errorf("--vpn-host is required for connect command")
	}

	vpnExec, err := getVPNExec(cmd)
	if err != nil {
		return err
	}

	err = connectVPN(vpnExec, vpnHost, username, method, verbose)
	if err != nil {
		return err
	}

	fmt.Println("VPN connection successful")
	return nil
}

// disconnectAction handles the disconnect command
func disconnectAction(ctx context.Context, cmd *cli.Command) error {
	verbose := cmd.Bool("verbose")

	vpnExec, err := getVPNExec(cmd)
	if err != nil {
		return err
	}

	err = disconnectVPN(vpnExec, verbose)
	if err != nil {
		return err
	}

	fmt.Println("VPN disconnection successful")
	return nil
}

// statusAction handles the status command
func statusAction(ctx context.Context, cmd *cli.Command) error {
	vpnExec, err := getVPNExec(cmd)
	if err != nil {
		return err
	}

	// FIXME: this code is duplicated
	// Start spinner for connection process
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Checking VPN Status..."
	s.Start()
	defer s.Stop()

	connected := vpnConnected(vpnExec)

	s.Stop()
	if connected {
		fmt.Println("VPN Connected: Yes")
	} else {
		fmt.Println("VPN Connected: No")
	}
	return nil
}

func main() {
	// Set default method from environment variable
	defaultMethod := os.Getenv("VPN_METHOD")
	if defaultMethod == "" {
		defaultMethod = "push"
	}

	cmd := &cli.Command{
		Name:  "seccli",
		Usage: "CLI wrapper around Cisco Secure Client",
		Commands: []*cli.Command{
			{
				Name:  "connect",
				Usage: "Connect to VPN",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "username",
						Aliases:  []string{"u"},
						Usage:    "Your VPN username",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "vpn-host",
						Aliases:  []string{"h"},
						Usage:    "VPN URL",
						Required: true,
					},
					&cli.StringFlag{
						Name:    "method",
						Aliases: []string{"m"},
						Usage:   "Authentication method",
						Value:   defaultMethod,
					},
					&cli.StringFlag{
						Name:  "vpn-exec",
						Usage: "Path to VPN executable (auto-detected if not provided)",
					},
					&cli.BoolFlag{
						Name:    "verbose",
						Aliases: []string{"v"},
						Usage:   "Show verbose output from VPN tool",
					},
				},
				Action: connectAction,
			},
			{
				Name:  "disconnect",
				Usage: "Disconnect from VPN",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "vpn-exec",
						Usage: "Path to VPN executable (auto-detected if not provided)",
					},
					&cli.BoolFlag{
						Name:    "verbose",
						Aliases: []string{"v"},
						Usage:   "Show verbose output from VPN tool",
					},
				},
				Action: disconnectAction,
			},
			{
				Name:  "status",
				Usage: "Show VPN connection status",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "vpn-exec",
						Usage: "Path to VPN executable (auto-detected if not provided)",
					},
				},
				Action: statusAction,
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
