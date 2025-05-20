# Bitrise App Analyzer

A Go-based tool for analyzing iOS and Android application files (.app, .ipa, .xcarchive, .apk).

## Features

- iOS app analysis (.app, .ipa, .xcarchive)
  - Extract and analyze Info.plist
  - Analyze embedded frameworks
  - Check entitlements
  - Analyze provisioning profiles
  - Extract and analyze embedded resources
- Android app analysis (coming soon)
  - Analyze AndroidManifest.xml
  - Extract and analyze resources
  - Check signing information

## Requirements

- Go 1.21 or later
- macOS (for iOS app analysis)

## Installation

```bash
go install github.com/yourusername/bitrise-app-analyze@latest
```

## Usage

```bash
# Analyze an iOS app
bitrise-app-analyze analyze --platform ios --path /path/to/your.app

# Analyze an iOS archive
bitrise-app-analyze analyze --platform ios --path /path/to/your.xcarchive

# Analyze an IPA file
bitrise-app-analyze analyze --platform ios --path /path/to/your.ipa
```

## Development

1. Clone the repository
2. Install dependencies: `go mod tidy`
3. Run tests: `go test ./...`
4. Build: `go build`

## License

MIT
