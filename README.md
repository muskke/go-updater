# Go Updater

Go Updater is a command-line tool for managing and updating Go tools installed in your `$GOPATH/bin` or `$GOBIN` directory. It automatically scans for executables, checks for the latest versions on their respective repositories, and provides an easy way to update them all at once.

## Features

- **Automatic Discovery**: Scans your `GOBIN` directory to find all installed Go tools.
- **Update Checks**: For each tool, it checks if a newer version is available.
- **Smart Versioning**: Correctly identifies module paths and versions, even for major version updates (e.g., `v1` to `v2`).
- **Interactive Updates**: Prompts before updating and shows a clear overview of which tools have updates available.
- **Concurrent Operations**: Performs checks and updates concurrently for faster execution.
- **Cross-Platform**: Works on both Windows and Unix-like systems (Linux, macOS).

## How It Works

1.  **Scan**: The tool starts by finding your `GOBIN` path from your Go environment. It then scans this directory for all executable files.
2.  **Analyze**: For each executable, it runs `go version -m` to extract its module information. This allows it to identify the tool's package path (e.g., `github.com/go-delve/delve/cmd/dlv`) and its current version.
3.  **Check for Updates**: With the module path, it uses `go list -m -u -json [module-path]@latest` to query for the latest available version of the module. It includes special logic to check for the existence of new major versions (e.g., if a tool is on `v2`, it will check if a `v3` is available).
4.  **Prompt and Update**: After checking all tools, it lists all the ones with available updates and prompts the user for confirmation. If confirmed, it runs `go install [package-path]@latest` for each tool to update it to the newest version.

## Usage

### Prerequisites

- Go 1.18 or higher installed.
- Your `GOBIN` environment variable must be set, or your `$GOPATH/bin` must be in your system's `PATH`.

### Building from Source

1.  Clone the repository:
    ```sh
    git clone https://github.com/muskke/go-updater.git
    cd go-updater
    ```

2.  Build the executable:
    ```sh
    go build -o go-updater .
    ```

### Running the Tool

Simply run the executable from your terminal:

```sh
./go-updater
```

The tool will automatically scan, check for updates, and prompt you if any actions are needed.

## License

This project is licensed under the MIT License. See the `LICENSE` file for details.