# HTML Image Embedder

## Overview

This program processes an HTML file and embeds images directly into the file as base64 encoded strings. It's designed
for situations where standalone HTML files are required, without external dependencies on image files.

### Running the Program

In the program directory, execute:
```bash
go run main.go
```

Parameters:

    inputFile: Path to the HTML file to be processed.
    outputFile: Path for saving the processed HTML file.

These are currently set in the main function.

### Supports embedding the following image formats:

    JPEG
    PNG
    GIF
    SVG (basic)

For SVGs, the program looks for <svg tags or .svg file extensions.

### Notes

- Ensure legal rights to use and embed images.
- Embedding many or large images can significantly increase HTML file size.