package main

import (
	"fmt"
	"os"
)

const version = "0.1.0"

func main() {
	fmt.Println("FacePass v" + version)
	fmt.Println("Face Recognition Authentication for Linux\n")

	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := os.Args[1]

	switch command {
	case "enroll":
		if len(os.Args) < 3 {
			fmt.Println("Error: username required")
			fmt.Println("Usage: facepass enroll <username>")
			return
		}
		fmt.Printf("Enrolling user: %s (not implemented yet)\n", os.Args[2])

	case "test":
		if len(os.Args) < 3 {
			fmt.Println("Error: username required")
			fmt.Println("Usage: facepass test <username>")
			return
		}
		fmt.Printf("Testing recognition for: %s (not implemented yet)\n", os.Args[2])

	case "list":
		fmt.Println("Enrolled users: (not implemented yet)")

	case "remove":
		if len(os.Args) < 3 {
			fmt.Println("Error: username required")
			fmt.Println("Usage: facepass remove <username>")
			return
		}
		fmt.Printf("Removing user: %s (not implemented yet)\n", os.Args[2])

	case "add-face":
		if len(os.Args) < 3 {
			fmt.Println("Error: username required")
			fmt.Println("Usage: facepass add-face <username>")
			return
		}
		fmt.Printf("Adding additional face angles for: %s (not implemented yet)\n", os.Args[2])

	case "config":
		fmt.Println("Configuration: (not implemented yet)")

	case "version":
		fmt.Println("FacePass v" + version)

	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
	}
}

func printUsage() {
	fmt.Println("Usage: facepass [command]")
	fmt.Println("\nCommands:")
	fmt.Println("  enroll <username>     Enroll a new face (captures 5-7 angles)")
	fmt.Println("  add-face <username>   Add additional face angles to existing enrollment")
	fmt.Println("  test <username>       Test face recognition")
	fmt.Println("  remove <username>     Remove face data")
	fmt.Println("  list                  List enrolled users")
	fmt.Println("  config                Show/edit configuration")
	fmt.Println("  version               Show version")
	fmt.Println("\nExamples:")
	fmt.Println("  facepass enroll john")
	fmt.Println("  facepass test john")
	fmt.Println("  facepass add-face john")
}
