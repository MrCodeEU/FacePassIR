package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/MrCodeEU/facepass/pkg/camera"
	"github.com/MrCodeEU/facepass/pkg/config"
	"github.com/MrCodeEU/facepass/pkg/liveness"
	"github.com/MrCodeEU/facepass/pkg/logging"
	"github.com/MrCodeEU/facepass/pkg/recognition"
	"github.com/MrCodeEU/facepass/pkg/storage"
)

const version = "0.2.0"

// Command represents a CLI command.
type Command struct {
	Name        string
	Description string
	Usage       string
	Run         func(args []string) error
}

var (
	cfg        *config.Config
	commands   map[string]*Command
	recognizer *recognition.DlibRecognizer
	store      *storage.FileStorage
)

// Enrollment angles to capture
var enrollmentAngles = []string{"front", "left", "right", "up", "down"}

func init() {
	commands = map[string]*Command{
		"enroll": {
			Name:        "enroll",
			Description: "Enroll a new face (captures 5 angles)",
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
		"cameras": {
			Name:        "cameras",
			Description: "List available cameras",
			Usage:       "facepass cameras",
			Run:         cmdCameras,
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
		"download-models": {
			Name:        "download-models",
			Description: "Download required dlib models",
			Usage:       "facepass download-models [directory]",
			Run:         cmdDownloadModels,
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
	for _, name := range []string{"enroll", "add-face", "test", "remove", "list", "cameras", "config", "download-models", "version", "help"} {
		cmd := commands[name]
		fmt.Printf("  %-12s %s\n", cmd.Name, cmd.Description)
	}
	fmt.Println("\nExamples:")
	fmt.Println("  facepass enroll john       # Enroll user 'john'")
	fmt.Println("  facepass test john         # Test recognition for 'john'")
	fmt.Println("  facepass -debug enroll me  # Enroll with debug output")
	fmt.Println("\nRun 'facepass help <command>' for more information on a command.")
}

// initRecognizer initializes the face recognizer.
func initRecognizer() error {
	if recognizer != nil && recognizer.IsLoaded() {
		return nil
	}

	recognizer = recognition.NewRecognizer()
	recognizer.SetTolerance(cfg.Recognition.Tolerance)

	if err := recognizer.LoadModels(cfg.Recognition.ModelPath); err != nil {
		return fmt.Errorf("failed to load face recognition models: %w\n\nPlease ensure dlib models are installed in: %s\n\nRequired files:\n  - shape_predictor_5_face_landmarks.dat\n  - dlib_face_recognition_resnet_model_v1.dat\n\nDownload from: http://dlib.net/files/", err, cfg.Recognition.ModelPath)
	}

	return nil
}

// initStorage initializes the storage system.
func initStorage() error {
	if store != nil {
		return nil
	}

	var err error
	store, err = storage.NewFileStorage(cfg.Storage.DataDir, cfg.Storage.EncryptionEnabled)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	return nil
}

// waitForEnter waits for user to press Enter.
func waitForEnter(prompt string) {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	_, _ = reader.ReadString('\n')
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

	// Initialize storage
	if err := initStorage(); err != nil {
		return err
	}

	// Check if user already enrolled
	if store.UserExists(username) {
		return fmt.Errorf("user '%s' is already enrolled. Use 'facepass add-face %s' to add more angles or 'facepass remove %s' first", username, username, username)
	}

	// Initialize recognizer
	if err := initRecognizer(); err != nil {
		return err
	}
	defer func() { _ = recognizer.Close() }()

	// Initialize camera
	cam := camera.NewCamera()
	_ = cam.SetResolution(cfg.Camera.Width, cfg.Camera.Height)

	// Select camera device
	device := cfg.Camera.Device
	if cfg.Camera.PreferIR {
		// Try IR camera first
		if _, err := os.Stat(cfg.Camera.IRDevice); err == nil {
			device = cfg.Camera.IRDevice
		}
	}

	if err := cam.Open(device); err != nil {
		return fmt.Errorf("failed to open camera %s: %w", device, err)
	}
	defer func() { _ = cam.Close() }()

	// Enable IR emitter if available
	if cam.HasIREmitter() && cfg.Camera.IREmitterEnabled {
		_ = cam.EnableIREmitter()
		defer func() { _ = cam.DisableIREmitter() }()
	}

	// Start streaming for faster capture
	if err := cam.StartStreaming(); err != nil {
		logging.Warnf("Failed to start streaming, falling back to single capture: %v", err)
	}
	defer func() { _ = cam.StopStreaming() }()

	fmt.Printf("\nStarting enrollment for '%s'...\n", username)
	fmt.Println("Please ensure good lighting and face the camera.")
	fmt.Println("You will be prompted to capture 5 different angles.")

	embeddings := make([]recognition.Embedding, 0, len(enrollmentAngles))

	for i, angle := range enrollmentAngles {
		prompt := getAnglePrompt(angle)
		fmt.Printf("[%d/%d] %s\n", i+1, len(enrollmentAngles), prompt)
		waitForEnter("      Press Enter when ready...")

		// Capture frame (uses ReadFrame which handles streaming)
		fmt.Print("      Capturing... ")

		// We capture a few frames to ensure we get a good one, but for enrollment
		// we just take the first valid one for now.
		// In the future we could average them.
		var frame *camera.Frame
		var err error

		// Try to read up to 5 frames to clear buffer/get fresh frame
		for k := 0; k < 5; k++ {
			frame, err = cam.ReadFrame()
			if err == nil {
				break
			}
		}

		if err != nil {
			fmt.Printf("FAILED: %v\n", err)
			fmt.Println("      Skipping this angle, continuing...")
			continue
		}

		// Detect and recognize face
		embedding, err := recognizer.RecognizeFace(frame.Data, angle)
		if err != nil {
			fmt.Printf("FAILED: %v\n", err)
			switch err {
			case recognition.ErrNoFaceDetected:
				fmt.Println("      No face detected. Please ensure your face is visible.")
			case recognition.ErrMultipleFaces:
				fmt.Println("      Multiple faces detected. Please ensure only you are in frame.")
			}
			fmt.Println("      Skipping this angle, continuing...")
			continue
		}

		embeddings = append(embeddings, *embedding)
		fmt.Println("OK")
	}

	if len(embeddings) < 3 {
		return fmt.Errorf("enrollment failed: only %d angles captured (minimum 3 required)", len(embeddings))
	}

	// Save user data
	metadata := map[string]string{
		"camera":      device,
		"version":     version,
		"enrolled_by": "cli",
	}

	if err := store.CreateUser(username, embeddings, metadata); err != nil {
		return fmt.Errorf("failed to save enrollment data: %w", err)
	}

	fmt.Printf("\nEnrollment complete! %d angles captured.\n", len(embeddings))
	fmt.Printf("User '%s' is now enrolled.\n", username)
	fmt.Println("\nTip: Test recognition with: facepass test", username)

	return nil
}

func cmdAddFace(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("username required\nUsage: facepass add-face <username>")
	}
	username := args[0]

	// Initialize storage
	if err := initStorage(); err != nil {
		return err
	}

	// Check if user is enrolled
	if !store.UserExists(username) {
		return fmt.Errorf("user '%s' is not enrolled. Use 'facepass enroll %s' first", username, username)
	}

	// Initialize recognizer
	if err := initRecognizer(); err != nil {
		return err
	}
	defer func() { _ = recognizer.Close() }()

	// Initialize camera
	cam := camera.NewCamera()
	_ = cam.SetResolution(cfg.Camera.Width, cfg.Camera.Height)

	device := cfg.Camera.Device
	if cfg.Camera.PreferIR {
		if _, err := os.Stat(cfg.Camera.IRDevice); err == nil {
			device = cfg.Camera.IRDevice
		}
	}

	if err := cam.Open(device); err != nil {
		return fmt.Errorf("failed to open camera: %w", err)
	}
	defer func() { _ = cam.Close() }()

	if cam.HasIREmitter() && cfg.Camera.IREmitterEnabled {
		_ = cam.EnableIREmitter()
		defer func() { _ = cam.DisableIREmitter() }()
	}

	// Start streaming for faster capture
	if err := cam.StartStreaming(); err != nil {
		logging.Warnf("Failed to start streaming, falling back to single capture: %v", err)
	}
	defer func() { _ = cam.StopStreaming() }()

	fmt.Printf("\nAdding face angles for '%s'...\n", username)
	fmt.Println("Position yourself and press Enter when ready.")

	waitForEnter("Press Enter to capture... ")

	fmt.Print("Capturing... ")

	// Read a few frames to clear buffer
	var frame *camera.Frame
	var err error
	for k := 0; k < 5; k++ {
		frame, err = cam.ReadFrame()
		if err == nil {
			break
		}
	}

	if err != nil {
		return fmt.Errorf("capture failed: %w", err)
	}

	embedding, err := recognizer.RecognizeFace(frame.Data, "additional")
	if err != nil {
		return fmt.Errorf("face recognition failed: %w", err)
	}

	if err := store.AddEmbedding(username, *embedding); err != nil {
		return fmt.Errorf("failed to save embedding: %w", err)
	}

	fmt.Println("OK")
	fmt.Printf("Additional face angle added for '%s'.\n", username)

	return nil
}

func cmdTest(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("username required\nUsage: facepass test <username>")
	}
	username := args[0]

	// Initialize storage
	if err := initStorage(); err != nil {
		return err
	}

	// Check if user is enrolled
	if !store.UserExists(username) {
		return fmt.Errorf("user '%s' is not enrolled. Use 'facepass enroll %s' first", username, username)
	}

	// Load user embeddings
	storedEmbeddings, err := store.GetAllEmbeddings(username)
	if err != nil {
		return fmt.Errorf("failed to load user data: %w", err)
	}

	// Initialize recognizer
	if err := initRecognizer(); err != nil {
		return err
	}
	defer func() { _ = recognizer.Close() }()

	// Initialize camera
	cam := camera.NewCamera()
	_ = cam.SetResolution(cfg.Camera.Width, cfg.Camera.Height)

	device := cfg.Camera.Device
	if cfg.Camera.PreferIR {
		if _, err := os.Stat(cfg.Camera.IRDevice); err == nil {
			device = cfg.Camera.IRDevice
		}
	}

	if err := cam.Open(device); err != nil {
		return fmt.Errorf("failed to open camera: %w", err)
	}
	defer func() { _ = cam.Close() }()

	if cam.HasIREmitter() && cfg.Camera.IREmitterEnabled {
		_ = cam.EnableIREmitter()
		defer func() { _ = cam.DisableIREmitter() }()
	}

	fmt.Printf("\nTesting face recognition for '%s'...\n", username)
	fmt.Println("Look at the camera and press Enter.")

	waitForEnter("Press Enter when ready... ")

	fmt.Println("Capturing and analyzing (capturing multiple frames)... ")

	// Start streaming for faster capture
	if err := cam.StartStreaming(); err != nil {
		logging.Warnf("Failed to start streaming, falling back to single capture: %v", err)
	}
	defer func() { _ = cam.StopStreaming() }()

	// Initialize liveness detector
	livenessCfg := liveness.DefaultConfig()
	livenessCfg.Level = liveness.LevelStandard
	// Map thresholds from config
	livenessCfg.MovementThreshold = cfg.Liveness.Thresholds.Movement
	livenessCfg.DepthThreshold = cfg.Liveness.Thresholds.Depth
	livenessCfg.ConsistencyThreshold = cfg.Liveness.Thresholds.Consistency
	detector := liveness.NewDetector(livenessCfg)

	var frames []liveness.Frame
	var embeddings []recognition.Embedding

	fmt.Println("\nFrame Analysis (Debug):")
	fmt.Println("Frame | Face | EAR   | Blink? | Action")
	fmt.Println("------+------+-------+--------+-------")

	// Pipeline Architecture:
	// 1. Capture Goroutine -> rawFramesChan
	// 2. Worker Pool -> resultsChan
	// 3. Main Thread -> Collects results

	type captureJob struct {
		index int
		frame *camera.Frame
		err   error
	}

	type processedFrame struct {
		index     int
		liveFrame liveness.Frame
		embedding *recognition.Embedding
		logMsg    string
	}

	// Performance Tuning:
	// We capture 30 frames (approx 1 second at 30fps) to ensure we catch blinks and movement.
	// However, we only process every 3rd frame (10 frames total) to reduce CPU load.
	// This gives us a good tradeoff: 1s temporal coverage but 3x faster processing.
	const captureCount = 30
	const processInterval = 3

	// Calculate how many frames we will actually process
	processCount := (captureCount + processInterval - 1) / processInterval

	rawFramesChan := make(chan captureJob, processCount)
	resultsChan := make(chan processedFrame, processCount)

	// Start Capture Goroutine
	startCapture := time.Now()
	go func() {
		defer close(rawFramesChan)
		for i := 0; i < captureCount; i++ {
			frameStart := time.Now()
			camFrame, err := cam.ReadFrame()

			// Log slow frames to debug
			duration := time.Since(frameStart)
			if duration > 100*time.Millisecond {
				logging.Debugf("Slow frame capture: %v", duration)
			}

			// Only process every Nth frame
			if i%processInterval == 0 {
				rawFramesChan <- captureJob{index: i, frame: camFrame, err: err}
			}
		}
	}()

	// Start Worker Pool
	numWorkers := runtime.NumCPU()
	var wg sync.WaitGroup

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range rawFramesChan {
				if job.err != nil {
					logging.Warnf("Failed to capture frame %d: %v", job.index, job.err)
					continue
				}

				camFrame := job.frame
				liveFrame := liveness.Frame{
					Data:      camFrame.Data,
					IsIR:      cam.GetDeviceInfo().IsIR,
					Timestamp: camFrame.Timestamp,
					FaceFound: false,
				}

				var emb *recognition.Embedding
				var logMsg string

				// Detect face
				face, err := recognizer.DetectSingleFace(camFrame.Data)
				if err == nil {
					liveFrame.FaceFound = true
					embVal := recognizer.GetEmbedding(face, "test")
					emb = &embVal
					liveFrame.Embedding = embVal // Copy embedding to frame

					// Convert landmarks
					var landmarks []liveness.Point
					for _, p := range face.Landmarks {
						landmarks = append(landmarks, liveness.Point{X: float64(p.X), Y: float64(p.Y)})
					}
					liveFrame.Landmarks = landmarks

					// Calculate EAR
					if len(landmarks) >= 5 {
						leftEye := landmarks[0:2]
						rightEye := landmarks[2:4]
						leftEAR := liveness.CalculateEyeAspectRatio(leftEye)
						rightEAR := liveness.CalculateEyeAspectRatio(rightEye)
						liveFrame.EyeAspectRatio = (leftEAR + rightEAR) / 2.0
					}
					logMsg = fmt.Sprintf(" %4d | Yes  | %.3f |        |\n", job.index, liveFrame.EyeAspectRatio)
				} else {
					logMsg = fmt.Sprintf(" %4d | No   | ----- |        |\n", job.index)
				}

				resultsChan <- processedFrame{index: job.index, liveFrame: liveFrame, embedding: emb, logMsg: logMsg}
			}
		}()
	}

	// Wait for workers in background to close results channel
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results as they come in
	processedResults := make([]processedFrame, 0, processCount)
	for res := range resultsChan {
		processedResults = append(processedResults, res)
	}

	captureDuration := time.Since(startCapture)
	_ = cam.StopStreaming() // Stop streaming immediately to save resources

	fmt.Printf("Captured %d frames, processed %d frames in %v.\n",
		captureCount, len(processedResults), captureDuration)

	// Sort by index to maintain order
	sort.Slice(processedResults, func(i, j int) bool {
		return processedResults[i].index < processedResults[j].index
	})

	// Output results and build final lists
	frames = make([]liveness.Frame, 0, len(processedResults))
	for _, res := range processedResults {
		frames = append(frames, res.liveFrame)
		if res.embedding != nil {
			embeddings = append(embeddings, *res.embedding)
		}
		fmt.Print(res.logMsg)
	}

	if len(embeddings) == 0 {
		fmt.Println("FAILED: No face detected in any frame")
		return nil
	}

	// Liveness Check
	livenessResult := detector.Detect(frames)

	// Recognition (use average embedding)
	avgEmbedding := recognition.AverageEmbedding(embeddings)
	_, distance, matched := recognizer.FindBestMatch(avgEmbedding, storedEmbeddings)

	fmt.Println("Done")
	fmt.Println()

	// Calculate confidence (inverse of distance, normalized)
	confidence := 1.0 - (distance / 1.0)
	if confidence < 0 {
		confidence = 0
	}
	if confidence > 1 {
		confidence = 1
	}

	fmt.Println("Results:")
	fmt.Printf("  Distance:   %.4f\n", distance)
	fmt.Printf("  Confidence: %.1f%%\n", confidence*100)
	fmt.Printf("  Threshold:  %.2f\n", cfg.Recognition.Tolerance)
	fmt.Printf("  Liveness:   %v (Score: %.2f)\n", livenessResult.IsLive, livenessResult.Score)
	if !livenessResult.IsLive {
		fmt.Printf("  Liveness Reason: %s\n", livenessResult.Reason)
	}
	fmt.Println()

	if matched {
		if livenessResult.IsLive {
			fmt.Printf("SUCCESS: Face matches user '%s' and liveness confirmed\n", username)
			logging.Infof("Face recognition test PASSED for user: %s (distance: %.4f, liveness: %.2f)", username, distance, livenessResult.Score)
		} else {
			fmt.Printf("WARNING: Face matches user '%s' BUT liveness check failed\n", username)
			logging.Warnf("Face recognition test MATCHED but LIVENESS FAILED for user: %s (distance: %.4f, reason: %s)", username, distance, livenessResult.Reason)
		}
	} else {
		fmt.Printf("FAILED: Face does not match user '%s'\n", username)
		logging.Warnf("Face recognition test FAILED for user: %s (distance: %.4f)", username, distance)
	}

	// Update last used timestamp
	_ = store.UpdateLastUsed(username)

	return nil
}

func cmdRemove(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("username required\nUsage: facepass remove <username>")
	}
	username := args[0]

	// Initialize storage
	if err := initStorage(); err != nil {
		return err
	}

	if !store.UserExists(username) {
		return fmt.Errorf("user '%s' is not enrolled", username)
	}

	// Confirm deletion
	fmt.Printf("Are you sure you want to remove face data for '%s'? [y/N]: ", username)
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	if response != "y" && response != "yes" {
		fmt.Println("Cancelled.")
		return nil
	}

	logging.Infof("Removing face data for user: %s", username)

	if err := store.DeleteUser(username); err != nil {
		return fmt.Errorf("failed to remove user data: %w", err)
	}

	fmt.Printf("Face data for '%s' has been removed.\n", username)
	return nil
}

func cmdList(args []string) error {
	logging.Debug("Listing enrolled users")

	// Initialize storage
	if err := initStorage(); err != nil {
		return err
	}

	users, err := store.ListUsers()
	if err != nil {
		return fmt.Errorf("failed to list users: %w", err)
	}

	if len(users) == 0 {
		fmt.Println("No users enrolled.")
		return nil
	}

	fmt.Println("Enrolled users:")
	for _, username := range users {
		user, err := store.LoadUser(username)
		if err != nil {
			fmt.Printf("  - %s (error loading data)\n", username)
			continue
		}
		fmt.Printf("  - %s (%d embeddings, enrolled: %s)\n",
			username,
			len(user.Embeddings),
			user.EnrolledAt.Format("2006-01-02"))
	}
	fmt.Printf("\nTotal: %d user(s)\n", len(users))

	return nil
}

func cmdCameras(args []string) error {
	fmt.Println("Detecting cameras...")

	cameras, err := camera.ListCameras()
	if err != nil {
		return fmt.Errorf("failed to list cameras: %w", err)
	}

	if len(cameras) == 0 {
		fmt.Println("No cameras found.")
		return nil
	}

	fmt.Println("\nAvailable cameras:")
	for _, cam := range cameras {
		irLabel := ""
		if cam.IsIR {
			irLabel = " [IR]"
		}
		fmt.Printf("  %s: %s%s\n", cam.Path, cam.Name, irLabel)
		if cam.Driver != "" {
			fmt.Printf("       Driver: %s\n", cam.Driver)
		}
	}

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
	fmt.Println()
	fmt.Println("Components:")
	fmt.Println("  - Face Recognition: dlib/go-face")
	fmt.Println("  - Encryption: NaCl secretbox")
	fmt.Println("  - Camera: V4L2")
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

	switch cmdName {
	case "enroll":
		fmt.Println("\nEnrollment Process:")
		fmt.Println("  1. Position yourself in front of the camera")
		fmt.Println("  2. Ensure good lighting")
		fmt.Println("  3. Follow prompts to capture 5 face angles:")
		fmt.Println("     - Front (looking straight)")
		fmt.Println("     - Left (head turned left)")
		fmt.Println("     - Right (head turned right)")
		fmt.Println("     - Up (head tilted up)")
		fmt.Println("     - Down (head tilted down)")
		fmt.Println("  4. Face data is encrypted and stored locally")
	case "test":
		fmt.Println("\nTesting Process:")
		fmt.Println("  1. Look at the camera")
		fmt.Println("  2. The system captures your face")
		fmt.Println("  3. Compares against stored embeddings")
		fmt.Println("  4. Shows match confidence and result")
	case "config":
		fmt.Println("\nConfiguration Locations:")
		fmt.Println("  System: /etc/facepass/facepass.yaml")
		fmt.Println("  User:   ~/.config/facepass/facepass.yaml")
		fmt.Println("\nUse -config flag to specify a custom config file.")
	}

	return nil
}

// getAnglePrompt returns the instruction for capturing a specific angle.
func getAnglePrompt(angle string) string {
	prompts := map[string]string{
		"front": "Look directly at the camera",
		"left":  "Turn your head slightly to the LEFT",
		"right": "Turn your head slightly to the RIGHT",
		"up":    "Tilt your head slightly UP",
		"down":  "Tilt your head slightly DOWN",
	}
	if prompt, ok := prompts[angle]; ok {
		return prompt
	}
	return "Position your face"
}

// Unused but kept for potential future use
var _ = time.Now
