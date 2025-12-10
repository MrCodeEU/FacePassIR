package main

import (
	"compress/bzip2"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/MrCodeEU/facepass/pkg/logging"
)

func cmdDownloadModels(args []string) error {
	modelDir := cfg.Recognition.ModelPath
	if len(args) > 0 {
		modelDir = args[0]
	}

	logging.Infof("Downloading models to: %s", modelDir)

	if err := os.MkdirAll(modelDir, 0755); err != nil {
		return fmt.Errorf("failed to create model directory: %w", err)
	}

	models := []struct {
		Name string
		URL  string
	}{
		{
			Name: "shape_predictor_5_face_landmarks.dat",
			URL:  "http://dlib.net/files/shape_predictor_5_face_landmarks.dat.bz2",
		},
		{
			Name: "dlib_face_recognition_resnet_model_v1.dat",
			URL:  "http://dlib.net/files/dlib_face_recognition_resnet_model_v1.dat.bz2",
		},
		{
			Name: "mmod_human_face_detector.dat",
			URL:  "http://dlib.net/files/mmod_human_face_detector.dat.bz2",
		},
	}

	for _, model := range models {
		targetPath := filepath.Join(modelDir, model.Name)
		if _, err := os.Stat(targetPath); err == nil {
			logging.Infof("Model %s already exists, skipping", model.Name)
			continue
		}

		logging.Infof("Downloading %s...", model.Name)
		if err := downloadAndExtract(model.URL, targetPath); err != nil {
			return fmt.Errorf("failed to download %s: %w", model.Name, err)
		}
		logging.Infof("Successfully downloaded %s", model.Name)
	}

	logging.Info("All models downloaded successfully!")
	return nil
}

func downloadAndExtract(url, targetPath string) error {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Minute,
	}

	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Create output file
	out, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

	// Create bzip2 reader
	bz2Reader := bzip2.NewReader(resp.Body)

	// Copy to file
	_, err = io.Copy(out, bz2Reader)
	return err
}
