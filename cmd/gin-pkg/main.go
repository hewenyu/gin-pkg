package main

// import (
// 	"fmt"
// 	"io"
// 	"io/fs"
// 	"log"
// 	"os"
// 	"os/exec"
// 	"path/filepath"
// 	"runtime"
// 	"strings"

// 	"github.com/spf13/cobra"
// )

// var rootCmd = &cobra.Command{
// 	Use:   "gin-pkg",
// 	Short: "A CLI tool for scaffolding new Gin API projects",
// 	Long:  `gin-pkg is a CLI tool for scaffolding new Go API projects using the Gin framework with built-in JWT authentication, security validation, and user management.`,
// }

// var newCmd = &cobra.Command{
// 	Use:   "new [project-name]",
// 	Short: "Create a new project",
// 	Args:  cobra.ExactArgs(1),
// 	Run: func(cmd *cobra.Command, args []string) {
// 		projectName := args[0]
// 		createNewProject(projectName)
// 	},
// }

// func init() {
// 	rootCmd.AddCommand(newCmd)
// }

// func main() {
// 	if err := rootCmd.Execute(); err != nil {
// 		fmt.Println(err)
// 		os.Exit(1)
// 	}
// }

// func createNewProject(projectName string) {
// 	// Get current working directory
// 	cwd, err := os.Getwd()
// 	if err != nil {
// 		log.Fatalf("Failed to get current working directory: %v", err)
// 	}

// 	// Create project directory
// 	projectPath := filepath.Join(cwd, projectName)
// 	if err := os.MkdirAll(projectPath, 0755); err != nil {
// 		log.Fatalf("Failed to create project directory: %v", err)
// 	}

// 	fmt.Printf("Creating new project: %s\n", projectName)

// 	// Get the template directory (the gin-pkg project directory)
// 	templateDir := getTemplateDir()

// 	// Copy template files to new project
// 	copyTemplateFiles(templateDir, projectPath, projectName)

// 	// Initialize git repository
// 	initGitRepo(projectPath)

// 	// Update module name in go.mod
// 	updateModuleName(projectPath, projectName)

// 	fmt.Printf("\nProject created successfully! 🎉\n\n")
// 	fmt.Printf("To get started:\n\n")
// 	fmt.Printf("  cd %s\n", projectName)
// 	fmt.Printf("  go mod tidy\n")
// 	fmt.Printf("  go run main.go\n\n")
// 	fmt.Printf("The server will be available at http://localhost:8080\n")
// }

// func getTemplateDir() string {
// 	// First, check if we're running from the binary installed via go install
// 	exePath, err := os.Executable()
// 	if err == nil {
// 		// go up two directories from bin/gin-pkg to find the template
// 		return filepath.Join(filepath.Dir(filepath.Dir(exePath)), "github.com", "yueban", "gin-pkg")
// 	}

// 	// If we couldn't find it, or we're running in development, use the current file location
// 	_, filename, _, ok := runtime.Caller(0)
// 	if ok {
// 		// go up two directories from cmd/gin-pkg
// 		return filepath.Join(filepath.Dir(filepath.Dir(filepath.Dir(filename))))
// 	}

// 	// Last resort, try to find from GOPATH
// 	gopath := os.Getenv("GOPATH")
// 	if gopath != "" {
// 		templateDir := filepath.Join(gopath, "src", "github.com", "yueban", "gin-pkg")
// 		if _, err := os.Stat(templateDir); err == nil {
// 			return templateDir
// 		}
// 	}

// 	log.Fatalf("Failed to find template directory")
// 	return ""
// }

// func copyTemplateFiles(templateDir, projectPath, projectName string) {
// 	// List of directories and files to exclude
// 	excludes := []string{
// 		".git",
// 		".cursor",
// 		"cmd/gin-pkg",
// 		"go.sum",
// 		".gitignore",
// 		"LICENSE",
// 		"README.md",
// 	}

// 	err := filepath.Walk(templateDir, func(path string, info fs.FileInfo, err error) error {
// 		if err != nil {
// 			return err
// 		}

// 		// Get the relative path to the template directory
// 		relPath, err := filepath.Rel(templateDir, path)
// 		if err != nil {
// 			return err
// 		}

// 		// Skip excluded paths
// 		for _, exclude := range excludes {
// 			if strings.HasPrefix(relPath, exclude) {
// 				if info.IsDir() {
// 					return filepath.SkipDir
// 				}
// 				return nil
// 			}
// 		}

// 		// Get the target path in the new project
// 		targetPath := filepath.Join(projectPath, relPath)

// 		if info.IsDir() {
// 			// Create directory
// 			return os.MkdirAll(targetPath, info.Mode())
// 		} else {
// 			// Create parent directories if they don't exist
// 			targetDir := filepath.Dir(targetPath)
// 			if err := os.MkdirAll(targetDir, 0755); err != nil {
// 				return err
// 			}

// 			// Copy file
// 			return copyFile(path, targetPath)
// 		}
// 	})

// 	if err != nil {
// 		log.Fatalf("Failed to copy template files: %v", err)
// 	}

// 	// Create new README.md, .gitignore, and LICENSE
// 	createProjectFiles(projectPath, projectName)
// }

// func copyFile(src, dst string) error {
// 	// Open source file
// 	srcFile, err := os.Open(src)
// 	if err != nil {
// 		return err
// 	}
// 	defer srcFile.Close()

// 	// Create destination file
// 	dstFile, err := os.Create(dst)
// 	if err != nil {
// 		return err
// 	}
// 	defer dstFile.Close()

// 	// Copy content
// 	_, err = io.Copy(dstFile, srcFile)
// 	if err != nil {
// 		return err
// 	}

// 	// Copy file mode
// 	srcInfo, err := os.Stat(src)
// 	if err != nil {
// 		return err
// 	}
// 	return os.Chmod(dst, srcInfo.Mode())
// }

// func initGitRepo(projectPath string) {
// 	cmd := exec.Command("git", "init")
// 	cmd.Dir = projectPath
// 	cmd.Stdout = os.Stdout
// 	cmd.Stderr = os.Stderr
// 	if err := cmd.Run(); err != nil {
// 		fmt.Printf("Warning: failed to initialize git repository: %v\n", err)
// 	}
// }

// func updateModuleName(projectPath, projectName string) {
// 	goModPath := filepath.Join(projectPath, "go.mod")

// 	// Read go.mod
// 	content, err := os.ReadFile(goModPath)
// 	if err != nil {
// 		fmt.Printf("Warning: failed to read go.mod: %v\n", err)
// 		return
// 	}

// 	// Replace module name
// 	newContent := strings.Replace(
// 		string(content),
// 		"module github.com/hewenyu/gin-pkg",
// 		fmt.Sprintf("module %s", projectName),
// 		1,
// 	)

// 	// Write updated go.mod
// 	if err := os.WriteFile(goModPath, []byte(newContent), 0644); err != nil {
// 		fmt.Printf("Warning: failed to update go.mod: %v\n", err)
// 		return
// 	}

// 	// Update imports in all Go files
// 	updateImportsInGoFiles(projectPath, projectName)
// }

// func updateImportsInGoFiles(projectPath, projectName string) {
// 	err := filepath.Walk(projectPath, func(path string, info fs.FileInfo, err error) error {
// 		if err != nil {
// 			return err
// 		}

// 		// Only process .go files
// 		if !info.IsDir() && strings.HasSuffix(path, ".go") {
// 			// Read file
// 			content, err := os.ReadFile(path)
// 			if err != nil {
// 				return err
// 			}

// 			// Replace import paths
// 			newContent := strings.Replace(
// 				string(content),
// 				"github.com/hewenyu/gin-pkg",
// 				projectName,
// 				-1,
// 			)

// 			// Write updated file
// 			if err := os.WriteFile(path, []byte(newContent), info.Mode()); err != nil {
// 				return err
// 			}
// 		}

// 		return nil
// 	})

// 	if err != nil {
// 		fmt.Printf("Warning: failed to update imports in Go files: %v\n", err)
// 	}
// }

// func createProjectFiles(projectPath, projectName string) {
// 	// Create README.md
// 	readmeContent := fmt.Sprintf(`# %s

// A Go API service using the Gin framework with built-in JWT authentication, security validation, and user management.

// ## Features

// - **JWT Authentication**: Complete implementation with access and refresh tokens
// - **Request Security Validation**:
//   - Timestamp validation to prevent replay attacks
//   - Server-generated nonce system
//   - Request signing
// - **User Management**: Login, registration, token refresh, and user information
// - **Modern Stack**:
//   - Gin for API routing
//   - Ent for database operations
//   - Redis for cache and nonce management
//   - Viper for configuration

// ## API Specifications

// - Base URL: '/api/v1'
// - Content Type: 'application/json'
// - Character Encoding: UTF-8
// - Time Format: ISO 8601 (e.g., '2023-06-15T08:00:00Z')

// ## Getting Started

// 1. Make sure you have PostgreSQL and Redis running
// 2. Update the configuration in 'config/default.yaml'
// 3. Run the application:

// ```bash
// go mod tidy
// go run main.go
// ```

// The server will be available at http://localhost:8080
// `, projectName)

// 	if err := os.WriteFile(filepath.Join(projectPath, "README.md"), []byte(readmeContent), 0644); err != nil {
// 		fmt.Printf("Warning: failed to create README.md: %v\n", err)
// 	}

// 	// Create .gitignore
// 	gitignoreContent := `# Binaries for programs and plugins
// *.exe
// *.exe~
// *.dll
// *.so
// *.dylib

// # Test binary, built with 'go test -c'
// *.test

// # Output of the go coverage tool, specifically when used with LiteIDE
// *.out

// # Dependency directories (remove the comment below to include it)
// # vendor/

// # Go workspace file
// go.work

// # Environment variables
// .env

// # Compiled binary
// /bin/

// # macOS
// .DS_Store

// # IDE files
// .idea/
// .vscode/
// *.swp
// *.swo
// `

// 	if err := os.WriteFile(filepath.Join(projectPath, ".gitignore"), []byte(gitignoreContent), 0644); err != nil {
// 		fmt.Printf("Warning: failed to create .gitignore: %v\n", err)
// 	}

// 	// Create LICENSE (MIT)
// 	licenseContent := `MIT License

// Copyright (c) 2023

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
// `

// 	if err := os.WriteFile(filepath.Join(projectPath, "LICENSE"), []byte(licenseContent), 0644); err != nil {
// 		fmt.Printf("Warning: failed to create LICENSE: %v\n", err)
// 	}
// }
