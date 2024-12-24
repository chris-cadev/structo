<p align="center">
  <img src="assets/branding-image.webp" alt="branding image" style="width: 200px;" />
</p>

# Structo: File Organizer Tool

This is an experimental file organizer written in Go. It sorts files from an input folder into quarterly subfolders within an output folder, based on their last modification times. The program includes support for optional localization (currently English and Spanish) and configurable folder structure preservation.

## Features

- Organizes files into subfolders by year and quarter (e.g., `2024/Q1_JAN-FEB-MAR`)
- Preserves or flattens the folder structure based on user input
- Localized log messages in English (`en`) or Spanish (`es`)
- Automatically generates and appends log files with detailed operation records

## Getting Started

### Prerequisites

- [Go](https://golang.org/) (version 1.16 or newer)

### Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/your-username/file-organizer.git
   cd file-organizer
   ```

2. Build the executable:
   ```bash
   go build -o file-organizer
   ```

### Usage

Run the program with the required and optional arguments:

```bash
./file-organizer --input /path/to/input-folder --output /path/to/output-folder --lang en --preserve-structure
```

#### Arguments

| Argument               | Description                                                                       | Required | Default           |
| ---------------------- | --------------------------------------------------------------------------------- | -------- | ----------------- |
| `--input`              | Path to the input folder.                                                         | Yes      | None              |
| `--output`             | Path to the output folder.                                                        | No       | Same as `--input` |
| `--lang`               | Language to use for logs and messages (`en` for English, `es` for Spanish).       | No       | `en`              |
| `--preserve-structure` | Preserve the subfolder structure of the input folder under the quarterly folders. | No       | Disabled          |

### Example

Organize files from `/home/user/photos` into quarterly subfolders under `/home/user/sorted` with Spanish logs and preserving subfolder structures:

```bash
./file-organizer --input /home/user/photos --output /home/user/sorted --lang es --preserve-structure
```

## Logging

The program generates log files in the output directory, named in the format `.organizer_<timestamp>.log`. These logs include:

- Input and output folder paths
- Success and error messages for file operations
- Timestamps for operation start and completion

## Warnings

**This tool is experimental!** Please be aware of the following:

1. **Backup your files**: Always create a backup of your files before using this tool. Moving files may result in unintended consequences if the tool encounters unexpected errors.
2. **File handling risks**: There may be potential issues during file renaming or copying, particularly with permission restrictions or large file volumes.
3. **Error reporting**: While the tool provides extensive logs, not all edge cases may be handled perfectly.

## Contributing

Contributions, issues, and feature requests are welcome! Feel free to open a pull request or issue in this repository.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
