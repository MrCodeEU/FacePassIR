package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/MrCodeEU/facepass/pkg/config"
	"github.com/MrCodeEU/facepass/pkg/logging"
)

const version = "0.1.0"

// Command represents a CLI command.
type Command struct {
	Name        string
	Description string
	Usage       string
	Run         func(args []string) error
}

var (
	cfg      *config.Config
	commands map[string]*Command
)

func init() {
	commands = map[string]*Command{
		"enroll": {
			Name:        "enroll",
			Description: "Enroll a new face (captures 5-7 angles)",
			Usage:       "facepass enroll <username>",
			Run:         cmdEnroll,
		},
		"add-face": {
			Name:        "add-face",
			Description: "Add additional face angles to existing enrollment",
			Usage:       "facepass add-face <username>",
			Run:         cmdAddFace,
		},
		"test": {
			Name:        "test",
			Description: "Test face recognition for a user",
			Usage:       "facepass test <username>",
			Run:         cmdTest,
		},
		"remove": {
			Name:        "remove",
			Description: "Remove a user's face data",
			Usage:       "facepass remove <username>",
			Run:         cmdRemove,
		},
		"list": {
			Name:        "list",
			Description: "List all enrolled users",
			Usage:       "facepass list",
			Run:         cmdList,
		},
		"config": {
			Name:        "config",
			Description: "Show current configuration",
			Usage:       "facepass config",
			Run:         cmdConfig,
		},
		"version": {
			Name:        "version",
			Description: "Show version information",
			Usage:       "facepass version",
			Run:         cmdVersion,
		},
		"help": {
			Name:        "help",
			Description: "Show help information",
			Usage:       "facepass help [command]",
			Run:         cmdHelp,
		},
	}
}

func main() {
	// Parse global flags
	configFile := flag.String("config", "", "Path to configuration file")
	debug := flag.Bool("debug", false, "Enable debug logging")
	flag.Parse()

	// Get remaining args after flags
	args := flag.Args()

	// Load configuration
	var err error
	if *configFile != "" {
		cfg, err = config.Load(*configFile)
	} else {
		cfg, err = config.LoadDefault()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not load config: %v\n", err)
		cfg = config.DefaultConfig()
	}

	// Expand paths in config
	cfg.ExpandPaths()

	// Initialize logging
	logLevel := cfg.Logging.Level
	if *debug {
		logLevel = "debug"
	}
	if err := logging.Init(logLevel, cfg.Logging.File); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not initialize file logging: %v\n", err)
	}

	logging.Debugf("FacePass v%s starting", version)
	logging.Debugf("Config loaded, storage dir: %s", cfg.Storage.DataDir)

	// Show usage if no command provided
	if len(args) < 1 {
		printUsage()
		os.Exit(0)
	}

	// Find and run command
	cmdName := args[0]
	cmd, ok := commands[cmdName]
	if !ok {
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", cmdName)
		printUsage()
		os.Exit(1)
	}

	// Run the command
	if err := cmd.Run(args[1:]); err != nil {
		logging.WithError(err).Errorf("Command '%s' failed", cmdName)
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("FacePass - Face Recognition Authentication for Linux")
	fmt.Printf("Version: %s\n\n", version)
	fmt.Println("Usage: facepass [options] <command> [arguments]")
	fmt.Println("\nOptions:")
	fmt.Println("  -config <file>   Path to configuration file")
	fmt.Println("  -debug           Enable debug logging")
	fmt.Println("\nCommands:")
	for _, name := range []string{"enroll", "add-face", "test", "remove", "list", "config", "version", "help"} {
		cmd := commands[name]
		fmt.Printf("  %-12s %s\n", cmd.Name, cmd.Description)
	}
	fmt.Println("\nExamples:")
	fmt.Println("  facepass enroll john       # Enroll user 'john'")
	fmt.Println("  facepass test john         # Test recognition for 'john'")
	fmt.Println("  facepass -debug enroll me  # Enroll with debug output")
	fmt.Println("\nRun 'facepass help <command>' for more information on a command.")
}

// Command implementations

func cmdEnroll(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("username required\nUsage: facepass enroll <username>")
	}
	username := args[0]

	logging.Infof("Starting enrollment for user: %s", username)

	// Ensure directories exist
	if err := cfg.EnsureDirectories(); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Check if user already enrolled
	userPath := cfg.GetUserDataPath(username)
	if _, err := os.Stat(userPath); err == nil {
		return fmt.Errorf("user '%s' is already enrolled. Use 'facepass add-face %s' to add more angles or 'facepass remove %s' first", username, username, username)
	}

	fmt.Printf("Starting enrollment for '%s'...\n", username)
	fmt.Println("Please ensure good lighting and face the camera.")
	fmt.Println()

	// TODO: Implement actual enrollment
	// 1. Initialize camera
	// 2. Capture 5-7 angles
	// 3. Extract embeddings
	// 4. Store encrypted

	fmt.Println("[1/5] Look directly at camera... (not implemented)")
	fmt.Println("[2/5] Turn head slightly left... (not implemented)")
	fmt.Println("[3/5] Turn head slightly right... (not implemented)")
	fmt.Println("[4/5] Tilt head slightly up... (not implemented)")
	fmt.Println("[5/5] Tilt head slightly down... (not implemented)")
	fmt.Println()
	fmt.Printf("Enrollment for '%s' not yet implemented.\n", username)
	fmt.Println("Face recognition core needs to be implemented first (Phase 2).")

	return nil
}

func cmdAddFace(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("username required\nUsage: facepass add-face <username>")
	}
	username := args[0]

	// Check if user is enrolled
	userPath := cfg.GetUserDataPath(username)
	if _, err := os.Stat(userPath); os.IsNotExist(err) {
		return fmt.Errorf("user '%s' is not enrolled. Use 'facepass enroll %s' first", username, username)
	}

	logging.Infof("Adding face angles for user: %s", username)
	fmt.Printf("Adding additional face angles for '%s'...\n", username)
	fmt.Println("(Not implemented yet)")

	return nil
}

func cmdTest(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("username required\nUsage: facepass test <username>")
	}
	username := args[0]

	// Check if user is enrolled
	userPath := cfg.GetUserDataPath(username)
	if _, err := os.Stat(userPath); os.IsNotExist(err) {
		return fmt.Errorf("user '%s' is not enrolled. Use 'facepass enroll %s' first", username, username)
	}

	logging.Infof("Testing recognition for user: %s", username)
	fmt.Printf("Testing face recognition for '%s'...\n", username)
	fmt.Println("(Not implemented yet)")

	return nil
}

func cmdRemove(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("username required\nUsage: facepass remove <username>")
	}
	username := args[0]

	userPath := cfg.GetUserDataPath(username)
	if _, err := os.Stat(userPath); os.IsNotExist(err) {
		return fmt.Errorf("user '%s' is not enrolled", username)
	}

	logging.Infof("Removing face data for user: %s", username)

	if err := os.Remove(userPath); err != nil {
		return fmt.Errorf("failed to remove user data: %w", err)
	}

	fmt.Printf("Face data for '%s' has been removed.\n", username)
	return nil
}

func cmdList(args []string) error {
	logging.Debug("Listing enrolled users")

	usersDir := cfg.Storage.DataDir + "/users"
	entries, err := os.ReadDir(usersDir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No users enrolled.")
			return nil
		}
		return fmt.Errorf("failed to read users directory: %w", err)
	}

	var users []string
	for _, entry := range entries {
		if !entry.IsDir() && len(entry.Name()) > 5 {
			// Remove .json extension
			name := entry.Name()
			if len(name) > 5 && name[len(name)-5:] == ".json" {
				users = append(users, name[:len(name)-5])
			}
		}
	}

	if len(users) == 0 {
		fmt.Println("No users enrolled.")
		return nil
	}

	fmt.Println("Enrolled users:")
	for _, user := range users {
		fmt.Printf("  - %s\n", user)
	}
	fmt.Printf("\nTotal: %d user(s)\n", len(users))

	return nil
}

func cmdConfig(args []string) error {
	logging.Debug("Showing configuration")

	fmt.Println("Current Configuration:")
	fmt.Println("======================")
	fmt.Println()
	fmt.Println("[Camera]")
	fmt.Printf("  Device:          %s\n", cfg.Camera.Device)
	fmt.Printf("  Resolution:      %dx%d @ %d FPS\n", cfg.Camera.Width, cfg.Camera.Height, cfg.Camera.FPS)
	fmt.Printf("  Prefer IR:       %t\n", cfg.Camera.PreferIR)
	fmt.Printf("  IR Device:       %s\n", cfg.Camera.IRDevice)
	fmt.Printf("  IR Emitter:      %t\n", cfg.Camera.IREmitterEnabled)
	fmt.Println()
	fmt.Println("[Recognition]")
	fmt.Printf("  Confidence:      %.2f\n", cfg.Recognition.ConfidenceThreshold)
	fmt.Printf("  Tolerance:       %.2f\n", cfg.Recognition.Tolerance)
	fmt.Printf("  Model Path:      %s\n", cfg.Recognition.ModelPath)
	fmt.Println()
	fmt.Println("[Liveness Detection]")
	fmt.Printf("  Level:           %s\n", cfg.Liveness.Level)
	fmt.Printf("  Blink Required:  %t\n", cfg.Liveness.BlinkRequired)
	fmt.Printf("  Min Score:       %.2f\n", cfg.Liveness.MinLivenessScore)
	fmt.Println()
	fmt.Println("[Authentication]")
	fmt.Printf("  Timeout:         %d seconds\n", cfg.Auth.Timeout)
	fmt.Printf("  Max Attempts:    %d\n", cfg.Auth.MaxAttempts)
	fmt.Printf("  Fallback:        %t\n", cfg.Auth.FallbackEnabled)
	fmt.Println()
	fmt.Println("[Storage]")
	fmt.Printf("  Data Dir:        %s\n", cfg.Storage.DataDir)
	fmt.Printf("  Encryption:      %t\n", cfg.Storage.EncryptionEnabled)
	fmt.Println()
	fmt.Println("[Logging]")
	fmt.Printf("  Level:           %s\n", cfg.Logging.Level)
	fmt.Printf("  File:            %s\n", cfg.Logging.File)

	return nil
}

func cmdVersion(args []string) error {
	fmt.Printf("FacePass v%s\n", version)
	fmt.Println("Face Recognition Authentication for Linux")
	fmt.Println()
	fmt.Println("Build Information:")
	fmt.Printf("  Go version: %s\n", "1.21+")
	fmt.Printf("  Platform:   linux/amd64\n")
	return nil
}

func cmdHelp(args []string) error {
	if len(args) == 0 {
		printUsage()
		return nil
	}

	cmdName := args[0]
	cmd, ok := commands[cmdName]
	if !ok {
		return fmt.Errorf("unknown command: %s", cmdName)
	}

	fmt.Printf("Command: %s\n", cmd.Name)
	fmt.Printf("Description: %s\n", cmd.Description)
	fmt.Printf("Usage: %s\n", cmd.Usage)

	// Add specific help for each command
	switch cmdName {
	case "enroll":
		fmt.Println("\nEnrollment Process:")
		fmt.Println("  1. Position yourself in front of the camera")
		fmt.Println("  2. Ensure good lighting")
		fmt.Println("  3. Follow prompts to capture 5-7 face angles")
		fmt.Println("  4. Face data is encrypted and stored locally")
	case "test":
		fmt.Println("\nTesting Process:")
		fmt.Println("  1. Look at the camera")
		fmt.Println("  2. The system will attempt to recognize you")
		fmt.Println("  3. Results show match confidence")
	case "config":
		fmt.Println("\nConfiguration Locations:")
		fmt.Println("  System: /etc/facepass/facepass.yaml")
		fmt.Println("  User:   ~/.config/facepass/facepass.yaml")
		fmt.Println("\nUse -config flag to specify a custom config file.")
	}

	return nil
}
