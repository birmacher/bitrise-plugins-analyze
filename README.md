# App Analyzer Plugin for [Bitrise CLI](https://github.com/bitrise-io/bitrise)

A powerful tool for analyzing iOS app bundles, providing detailed insights about size, content, and potential optimizations.

## Installation

Can be run directly with the Bitrise CLI.

```bash
bitrise plugin install https://github.com/birmacher/bitrise-plugins-analyze.git
```

## Usage

The plugin provides two main commands: `analyze` and `diff`.

## Analyze Command

Analyzes an app bundle and provides detailed insights about its size and content.

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

## Diff Command

Compares two JSON analysis reports to identify file changes between app versions.

```bash
bitrise :diff [old.json] [new.json] [flags]
```

### Arguments

- `old.json`: Path to the JSON report of the old app version
- `new.json`: Path to the JSON report of the new app version

### Flags

- `--json`: Path to save the diff results as a JSON file

### Output

The diff command identifies and reports:
- Added files (with sizes)
- Removed files (with sizes)
- Changed files (with size comparisons)

### Examples

1. Compare two JSON reports:
```bash
bitrise :diff old.json new.json
```

2. Save diff results to a JSON file:
```bash
bitrise :diff old.json new.json --json=diff.json
```

## Requirements

- macOS (required for iOS app bundle analysis)
- Bitrise CLI installed
- For analyzing .ipa files: ability to extract and process iOS app bundles
- For analyzing Android images: the `cwebp` command must be installed
