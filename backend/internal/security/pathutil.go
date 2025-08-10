package security

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ValidateFilePath validates and cleans a file path to prevent directory traversal attacks
// and access to sensitive system directories.
func ValidateFilePath(filePath string) (string, error) {
	// Clean the path to resolve any . and .. elements
	cleanPath := filepath.Clean(filePath)

	// Check for path traversal attempts
	if strings.Contains(cleanPath, "..") {
		return "", fmt.Errorf("invalid file path: path traversal detected")
	}

	// Block access to sensitive system directories
	sensitiveDirectories := []string{
		"/etc/",
		"/root/",
		"/sys/",
		"/proc/",
		"/dev/",
		"/boot/",
		"/usr/bin/",
		"/usr/sbin/",
		"/sbin/",
		"/bin/",
	}

	for _, sensitiveDir := range sensitiveDirectories {
		if strings.HasPrefix(cleanPath, sensitiveDir) {
			return "", fmt.Errorf("invalid file path: access to system directories not allowed")
		}
	}

	return cleanPath, nil
}

// ValidateFilePathPair validates both source and destination file paths
func ValidateFilePathPair(src, dst string) (string, string, error) {
	cleanSrc, err := ValidateFilePath(src)
	if err != nil {
		return "", "", fmt.Errorf("invalid source path: %w", err)
	}

	cleanDst, err := ValidateFilePath(dst)
	if err != nil {
		return "", "", fmt.Errorf("invalid destination path: %w", err)
	}

	return cleanSrc, cleanDst, nil
}

// ValidatePathWithinDirectory ensures a file path is within a specific base directory
func ValidatePathWithinDirectory(filePath, baseDir string) (string, error) {
	cleanPath, err := ValidateFilePath(filePath)
	if err != nil {
		return "", err
	}

	cleanBaseDir := filepath.Clean(baseDir)

	// Ensure the path is within the base directory
	if !strings.HasPrefix(cleanPath, cleanBaseDir) {
		return "", fmt.Errorf("invalid path: outside base directory")
	}

	return cleanPath, nil
}
