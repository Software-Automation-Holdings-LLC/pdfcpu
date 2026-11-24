/*
Copyright 2018 The pdfcpu Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package model

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/pdfcpu/pdfcpu/internal/fileutil"
	"github.com/pdfcpu/pdfcpu/pkg/font"
	"github.com/pdfcpu/pdfcpu/pkg/log"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// NewStreamDictForFile creates a streamDict for filename.
func (xRefTable *XRefTable) NewStreamDictForFile(filename string) (*types.StreamDict, error) {
	buf, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return xRefTable.NewStreamDictForBuf(buf)
}

// NewEmbeddedFileStreamDict returns an embeddedFileStreamDict containing the file "filename".
func (xRefTable *XRefTable) NewEmbeddedFileStreamDict(filename string) (*types.IndirectRef, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}

	return xRefTable.NewEmbeddedStreamDict(f, fi.ModTime())
}

func initializeConfigurationFile(path string) error {
	return fileutil.WriteFile(path, defaultConfigurationFileBytes(), 0o600)
}

func readOpenConfigurationFile(f *os.File, path string) (conf *Configuration, err error) {
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			err = errors.Join(err, closeErr)
		}
	}()
	return readConfiguration(f, path)
}

func readConfigurationFile(path string) (*Configuration, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return readOpenConfigurationFile(f, path)
}

func readConfigurationAt(root string) (*Configuration, error) {
	if root == "" {
		return nil, errors.New("missing configuration root")
	}
	configDir := filepath.Join(root, "pdfcpu")
	path := filepath.Join(configDir, "config.yml")
	if _, err := preflightConfigurationSchema(path, ConfigurationSchemaVersionCurrent); err != nil {
		return nil, err
	}
	conf, err := readConfigurationFile(path)
	if err != nil {
		return nil, err
	}
	conf.resources = resourcesForConfigurationDir(configurationResourceModeReadOnly, configDir)
	return conf, nil
}

// LoadConfigurationReadOnly loads an existing configuration tree rooted at root without modifying filesystem or package
// loader state.
func LoadConfigurationReadOnly(root string) (*Configuration, error) {
	return readConfigurationAt(root)
}

func loadOrInitializeConfigurationFile(path string) (*Configuration, bool, error) {
	if _, err := preflightConfigurationSchema(path, ConfigurationSchemaVersionCurrent); err == nil {
		conf, err := readConfigurationFile(path)
		return conf, false, err
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, false, err
	}
	if err := initializeConfigurationFile(path); err != nil {
		return nil, false, err
	}
	conf, err := readConfigurationFile(path)
	return conf, true, err
}

func ensureConfigFileAt(path string, override bool) (*Configuration, error) {
	if !override {
		f, err := os.Open(path)
		if err == nil {
			conf, err := readOpenConfigurationFile(f, path)
			if err != nil {
				return nil, err
			}
			conf.resources = resourcesForConfigurationDir(configurationResourceModeAuto, filepath.Dir(path))
			return conf, nil
		}
	}

	if err := initializeConfigurationFile(path); err != nil {
		return nil, err
	}
	conf, err := readConfigurationFile(path)
	if err != nil {
		return nil, err
	}
	conf.resources = resourcesForConfigurationDir(configurationResourceModeAuto, filepath.Dir(path))
	return conf, nil
}

func onlyHidden(files []os.DirEntry) bool {
	for _, file := range files {
		if !strings.HasPrefix(file.Name(), ".") {
			return false
		}
	}
	return true
}

// ensureFontDirInitializedAt sets up the font directory without loading fonts.
// Font loading is deferred until fonts are actually needed.
func ensureFontDirInitializedAt(userFontDir string) error {
	files, err := os.ReadDir(userFontDir)
	if err != nil {
		return err
	}
	if onlyHidden(files) {
		// Ensure Roboto font for form filling.
		fontname := "Roboto-Regular"
		if log.DebugEnabled() && log.CLIEnabled() {
			log.CLI.Printf("installing user font: %s\n", fontname)
		}
		if err := font.InstallFontFromBytesQuiet(userFontDir, fontname, robotoFontFileBytes); err != nil {
			return err
		}
	}

	return nil
}

func initCertificatesAt(trustedCertDir string) error {
	files, err := os.ReadDir(trustedCertDir)
	if err != nil {
		return err
	}
	if !onlyHidden(files) {
		return nil
	}
	if !bundledDefaultCertificates {
		return nil
	}

	MarkCertificateStoreChanged()
	return installDefaultCertificates(trustedCertDir)
}

func initializeConfigurationResourcesAt(configDir string) error {
	// Initialize the user-font directory and install Roboto for Unicode form filling.
	// Other user fonts are installed with `pdfcpu font install` or copied from another configuration directory.
	// User fonts are loaded into memory lazily.
	userFontDir := filepath.Join(configDir, "fonts")
	if err := os.MkdirAll(userFontDir, 0o700); err != nil {
		return err
	}
	if err := ensureFontDirInitializedAt(userFontDir); err != nil {
		return err
	}

	// Initialize the certificate directory and install build defaults. Certificates are loaded into memory lazily.
	trustedCertDir := filepath.Join(configDir, "certs")
	if err := os.MkdirAll(trustedCertDir, 0o700); err != nil {
		return err
	}
	return initCertificatesAt(trustedCertDir)
}

func loadConfiguration(root string) (*Configuration, bool, error) {
	if root == "" {
		return nil, false, errors.New("missing configuration root")
	}

	configDir := filepath.Join(root, "pdfcpu")
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		return nil, false, err
	}
	conf, created, err := loadOrInitializeConfigurationFile(filepath.Join(configDir, "config.yml"))
	if err != nil {
		return nil, false, err
	}
	conf.resources = resourcesForConfigurationDir(configurationResourceModeAutoIsolated, configDir)
	if err := initializeConfigurationResourcesAt(configDir); err != nil {
		return nil, false, err
	}
	return conf, created, nil
}

// InitializeConfiguration loads or initializes a configuration tree and reports whether config.yml was created.
func InitializeConfiguration(root string) (*Configuration, bool, error) {
	return loadConfiguration(root)
}

// LoadConfiguration loads or initializes a configuration tree rooted at root without changing compatibility loader
// state.
func LoadConfiguration(root string) (*Configuration, error) {
	conf, _, err := loadConfiguration(root)
	return conf, err
}

// ResetConfiguration replaces config.yml with the built-in configuration at root without changing compatibility loader
// state. Existing font and certificate resources are preserved.
func ResetConfiguration(root string) (*Configuration, error) {
	if root == "" {
		return nil, errors.New("missing configuration root")
	}

	configDir := filepath.Join(root, "pdfcpu")
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		return nil, err
	}
	path := filepath.Join(configDir, "config.yml")
	if err := initializeConfigurationFile(path); err != nil {
		return nil, err
	}
	conf, err := readConfigurationFile(path)
	if err != nil {
		return nil, err
	}
	conf.resources = resourcesForConfigurationDir(configurationResourceModeAutoIsolated, configDir)
	if err := initializeConfigurationResourcesAt(configDir); err != nil {
		return nil, err
	}
	return conf, nil
}

// EnsureDefaultConfigAt tries to load the default configuration from path.
// If path/pdfcpu/config.yml is not found, it will be created.
func EnsureDefaultConfigAt(path string, override bool) error {
	defaultConfigMu.Lock()
	defer defaultConfigMu.Unlock()
	return ensureDefaultConfigAt(path, override)
}

// ensureDefaultConfigAt requires defaultConfigMu to be held for writing.
func ensureDefaultConfigAt(path string, override bool) error {
	configDir := filepath.Join(path, "pdfcpu")
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		return err
	}
	conf, err := ensureConfigFileAt(filepath.Join(configDir, "config.yml"), override)
	if err != nil {
		return err
	}

	if err := initializeConfigurationResourcesAt(configDir); err != nil {
		return err
	}

	font.UserFontDir = conf.resources.userFontDir
	TrustedCertDir = conf.resources.trustedCertDir
	loadedDefaultConfig = conf
	return nil
}
