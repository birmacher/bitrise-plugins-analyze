package appbundle

import (
	"bitrise-plugins-analyze/appbundle/android"
	"bitrise-plugins-analyze/appbundle/core"
	androidcore "bitrise-plugins-analyze/appbundle/core/android"
	"bitrise-plugins-analyze/appbundle/ios"
	"bitrise-plugins-analyze/appbundle/python"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func Analyze(bundle_path string) (*core.AppBundle, error) {
	ext := strings.ToLower(filepath.Ext(bundle_path))

	var err error
	var tmp_path string

	switch ext {
	case core.AppExtension, core.IpaExtension, core.XcarchiveExtension:
		tmp_path, err = ios.GetTempIPAPath(bundle_path)
	case core.ApkExtension, core.AabExtension:
		tmp_path, err = android.GetTempAPKPath(bundle_path)
	default:
		return nil, fmt.Errorf("unsupported file extension: %s", ext)
	}

	if err != nil {
		return nil, err
	}
	return analyzeAppBundle(tmp_path)
}

// AnalyzeAppBundle analyzes the provided app bundle directory and returns the analysis results
func analyzeAppBundle(bundlePath string) (*core.AppBundle, error) {
	bundle := &core.AppBundle{}
	var err error

	ext := strings.ToLower(filepath.Ext(bundlePath))
	switch ext {
	case core.AppExtension:
		err = analyzeiOSApp(bundlePath, bundle)
	case core.ApkExtension:
		err = analyzeAndroidApp(bundlePath, bundle)
	}
	return bundle, err
}

func analyzeiOSApp(bundlePath string, bundle *core.AppBundle) error {
	files, err := AnalyzeFile(bundlePath, bundlePath)
	if err != nil {
		return err
	}

	bundle.AppName = filepath.Base(bundlePath)
	bundle.Files = files

	// Calculate download size
	bundle.DownloadSize, err = ios.CalculateDownloadSize(bundlePath)
	if err != nil {
		return err
	}

	// Calculate install size using du command
	bundle.InstallSize, err = ios.CalculateInstallSize(bundlePath)
	if err != nil {
		return err
	}

	// Analyze Info.plist
	err = ios.AnalyzeInfoPlist(bundlePath, bundle)
	if err != nil {
		return err
	}

	pythonEnv := &python.PythonEnvironment{}
	err = pythonEnv.SetupPythonEnvironment()
	if err != nil {
		return err
	}
	defer pythonEnv.Cleanup()

	// Walk through the files and analyze the content further
	err = filepath.Walk(bundlePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		switch filepath.Ext(strings.ToLower(path)) {
		case ".car":
			carInfo, err := ios.ParseCARFile(path, bundlePath)
			if err != nil {
				return fmt.Errorf("failed to analyze %s: %v", path, err)
			}
			bundle.CarFiles = append(bundle.CarFiles, *carInfo)
		default:
			// Run file command to check if it's a Mach-O binary
			cmd := exec.Command("file", path)
			output, err := cmd.Output()
			if err != nil {
				return nil // Skip if file command fails
			}

			// Check if the file is a Mach-O binary
			if !strings.Contains(string(output), "Mach-O") {
				return nil
			}

			liefOutput, err := pythonEnv.AnalyzeLief(path)
			if err != nil {
				return err
			}

			relPath, err := filepath.Rel(bundlePath, path)
			if err != nil {
				return err
			}

			err = core.ParseLiefOutput(relPath, liefOutput, bundle)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk bundle directory: %v", err)
	}

	return nil
}

func analyzeAndroidApp(bundlePath string, bundle *core.AppBundle) error {
	// Analyze AndroidManifest.xml
	manifest, err := android.ParseAndroidManifest(bundlePath)
	if err != nil {
		return fmt.Errorf("failed to parse AndroidManifest.xml: %v", err)
	}

	// Set bundle metadata from manifest
	bundle.AppName = manifest.Application.Label
	bundle.BundleID = manifest.Package
	bundle.Version = manifest.VersionName + " (" + manifest.VersionCode + ")"
	bundle.SupportedPlatforms = []string{"Android"}

	unzipedApkDir, err := core.Unzip(bundlePath)
	defer func() {
		os.RemoveAll(unzipedApkDir)
	}()
	if err != nil {
		return fmt.Errorf("failed to unzip APK: %v", err)
	}

	// Analyze the APK files
	files, err := AnalyzeFile(unzipedApkDir, unzipedApkDir)
	if err != nil {
		return err
	}
	bundle.Files = files

	pythonEnv := &python.PythonEnvironment{}
	err = pythonEnv.SetupPythonEnvironment()
	if err != nil {
		return err
	}
	defer pythonEnv.Cleanup()

	// Walk through the files and analyze the content further
	err = filepath.Walk(unzipedApkDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		switch filepath.Ext(strings.ToLower(path)) {
		case ".dex":
			dexFiles, err := pythonEnv.AnalyzeDex(path)
			if err != nil {
				return err
			}

			err = android.ParseDexOutput(dexFiles, bundle)
			if err != nil {
				return err
			}
		case ".so":
			liefOutput, err := pythonEnv.AnalyzeLief(path)
			if err != nil {
				return err
			}

			relPath, err := filepath.Rel(unzipedApkDir, path)
			if err != nil {
				return err
			}

			err = core.ParseLiefOutput(relPath, liefOutput, bundle)
			if err != nil {
				return err
			}
		case ".arsc":
			// Analyze arsc files
			arscResources, err := pythonEnv.AnalyzeArsc(path)
			if err != nil {
				return fmt.Errorf("failed to analyze ARSC files: %v", err)
			} else {
				for resourceId, res := range arscResources {
					bundle.AsrcFiles = append(bundle.AsrcFiles, androidcore.AsrcFile{
						ResourceId: resourceId,
						Type:       res["type"],
						Name:       res["name"],
					})
				}
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk bundle directory: %v", err)
	}

	// Analyze images
	oversizedImages, err := android.AnalyzeImages(unzipedApkDir, manifest)
	if err != nil {
		return err
	}
	bundle.OversizedImages = oversizedImages

	// Calculate sizes
	bundle.InstallSize = files.Size

	// Get the original APK file size for download size
	apkInfo, err := os.Stat(bundlePath)
	if err != nil {
		return fmt.Errorf("failed to get APK file size: %v", err)
	}
	bundle.DownloadSize = apkInfo.Size()

	return nil
}
