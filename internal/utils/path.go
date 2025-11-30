package utils

import (
	"os"
	"path/filepath"
)

// GetUploadsPath returns the absolute path to the uploads directory
// It tries to find the path relative to the executable location
func GetUploadsPath() string {
	// Try to get the executable path
	exePath, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exePath)
		uploadsPath := filepath.Join(exeDir, "uploads")
		// Check if this path exists or if we're in production
		if _, err := os.Stat(uploadsPath); err == nil {
			return uploadsPath
		}
		// In production, the binary is at /www/wwwroot/ecommerce-backend/ecommerce-backend
		// So uploads should be at /www/wwwroot/ecommerce-backend/uploads
		if filepath.Base(exeDir) == "ecommerce-backend" {
			return uploadsPath
		}
	}

	// Fallback: try production path
	productionPath := "/www/wwwroot/ecommerce-backend/uploads"
	if _, err := os.Stat(productionPath); err == nil {
		return productionPath
	}

	// Last resort: use relative path (for local development)
	return "./uploads"
}

// GetUploadsSubPath returns the absolute path to a subdirectory in uploads
func GetUploadsSubPath(subdir string) string {
	return filepath.Join(GetUploadsPath(), subdir)
}

