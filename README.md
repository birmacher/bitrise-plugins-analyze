# App Analyzer Plugin for [Bitrise CLI](https://github.com/bitrise-io/bitrise)

A powerful tool for analyzing iOS app bundles, providing detailed insights about size, content, and potential optimizations.

## Installation

Can be run directly with the Bitrise CLI.

```bash
bitrise plugin install https://github.com/birmacher/bitrise-plugins-analyze.git
```

## Usage

Basic command structure:
```bash
bitrise :analyze [path] [flags]
```

### Arguments

- `path`: Path to the app bundle (.app), archive (.xcarchive), or IPA file (.ipa)

### Flags

- `--html`: Generate an interactive HTML visualization report
- `--json`: Generate a detailed JSON report
- `--markdown`: Generate a markdown report with key insights
- `--output-dir`: Directory where the output files will be generated (default: current directory)

### Output Files

All generated files will use the app's bundle ID as the base filename:
- HTML report: `<bundle_id>.html`
- JSON report: `<bundle_id>.json`
- Markdown report: `<bundle_id>.md`

### Examples

1. Basic analysis of an .app bundle:
```bash
bitrise :analyze MyApp.app
```

2. Generate HTML visualization:
```bash
bitrise :analyze MyApp.ipa --html
```

3. Generate all report formats:
```bash
bitrise :analyze MyApp.xcarchive --html --json --markdown
```

4. Specify output directory:
```bash
bitrise :analyze MyApp.app --html --output-dir=/path/to/reports
```

### Report Contents

The analysis provides detailed information about:
- Basic app information (bundle ID, version, size)
- Top 10 largest modules
- Top 10 largest files
- Duplicate content (both in file system and asset catalogs)

## Requirements

- macOS (required for iOS app bundle analysis)
- Bitrise CLI installed
- For analyzing .ipa files: ability to extract and process iOS app bundles