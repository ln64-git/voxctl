# voxctl

voxel is a terminal-based text-to-speech interface designed with modularity in mind. Once the server is running after inital launch, users can send provide input via command flag or API call. Sentences are parsed from user input sequentually for speedy response, audio clips are then queued respectively.

## Installation

1. Clone and build repository:

```
git clone https://github.com/ln64-git/voxctl.git
cd voxctl
go build cmd/voxctl.go
```

## Usage

```
./voxctl [flags]
```

### Flags

- `-input`: Input text to play
- `-port`: Port number to connect or serve (default: 8080)
- `-quit`: Exit application after request
- `-status`: Request info

### Examples

1. Launch server on a specified port:

```
./voxctl -port 4202
```

2. Send input to server:

```
./voxctl -input "Hello Server!!" -port 4202 -quit
```

## Configuration

The program requires an Azure Subscription Key and Region for the Speech Services. You can set these values in a configuration file (`voxctl.json`) located in the project directory. The file should have the following structure:

```json
{
  "AzureSubscriptionKey": "your_azure_subscription_key",
  "AzureRegion": "your_azure_region",
  "VoiceGender": "Female",
  "VoiceName": "en-US-JennyNeural"
}
```

Replace `your_azure_subscription_key` and `your_azure_region` with your actual Azure Speech Services credentials. You can also customize the voice gender and name by modifying the `VoiceGender` and `VoiceName` fields.

## Contributing

Contributions to voxctl are welcome! If you find any issues or have suggestions for improvements, please open an issue or submit a pull request on the GitHub repository.

## License

This project is licensed under the [MIT License](LICENSE).
