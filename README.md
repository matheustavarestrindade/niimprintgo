# NiimprintGO 

Welcome to NiimprintGO, a command-line tool designed for easy and efficient label printing with the Niimbot D11 printer. This document provides an overview of the available command-line parameters to customize your printing tasks.

## Supported printers

- Niimbot D11

## Installation

Before using NiimprintGO, ensure that you have the Niimbot D11 printer drivers installed on your system and the printer is properly connected. For installation instructions, refer to the Niimbot D11 documentation.

## Command-Line Flags

NiimprintGO supports a range of command-line flags to tailor the printing process to your needs. Here are the available flags:

### Logger Flags

- `--debug`: Enable or disable debug logs. (default: `false`)
- `--info`: Enable or disable info logs. (default: `true`)
- `--error`: Enable or disable error logs. (default: `true`)
- `--colors`: Enable or disable colors in logs. (default: `true`)

### Printing Flags

- `--labelType`: Set the label type. Valid values are `1`, `2`, or `3`. (default: `1`)
- `--labelDensity`: Set the label density. Valid values are `1`, `2`, or `3`. (default: `2`)
- `--quantity`: Specify the quantity of labels to print. (default: `1`)
- `--comPort`: Specify the COM port used for the printer connection.
- `--imagePath`: Specify the path to the image file to be printed on the label.

**Image requirements**: The image must have a maximum width of 96px and a maximum height of 600px.

Example usage:

```sh
NiimprintGO --labelType=2 --labelDensity=3 --quantity=5 --comPort=COM3 --imagePath="/path/to/image.png"
```

## Best Practices

- **Label Type and Density**: Experiment with different label types and densities to find the best combination for your specific labels and printer.
- **COM Port**: Ensure the `--comPort` flag is set to the correct port that your Niimbot D11 printer is connected to. You can find this information in your system's device manager.
- **Image Preparation**: Resize your images to fit within the maximum dimensions (96px width x 600px height) before printing to ensure the best quality and compatibility.

Happy Printing!
