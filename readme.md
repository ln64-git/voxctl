# voxctl

voxel is a terminal-based text-to-speech interface designed with modularity in mind. Once the server is running after inital launch, users can send provide input via command flag or API call. Sentences are parsed from user input sequentually for a speedy response, audio clips are then queued respectively.

## Installation

Clone and build repository:

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

Launch server on a specified port:

```
./voxctl -port 7000
```

Send input to server:

```
./voxctl -input "Hello Server!!" -port 7000 -quit
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

## How to obtain an Azure API key

1. Go to the Azure Portal (https://portal.azure.com) and sign in with your Microsoft account.
2. Click on "Create a resource" in the top left corner of the portal.
3. Search for "Speech" in the search bar and select "Speech" from the results.
4. Click on "Create" under the "Speech" service.
5. Once you've reached "Create Speech resource", enter the following information:
   - Subscription: Select your Azure subscription.
   - Resource group: Choose an existing resource group or create a new one.
   - Region: Select the region closest to you or the one you prefer.
   - Name: Enter a unique name for your Speech resource.
   - Pricing tier: Select the pricing tier that suits your needs (F0 is the free tier).
6. Click on "Review + create" and then "Create" to create the Speech resource.
7. After the deployment is complete, navigate to the Speech resource you just created.
8. In the left-hand menu, click on "Keys and Endpoint" under the "Resource Management" section.
9. Copy one of the two keys listed as the "Key 1" or "Key 2". This is your Azure Subscription Key.
10. Also, note down the Region you selected during the resource creation process. This is your Azure Region.

You can now use the Azure Subscription Key and Azure Region in your application's configuration file (`voxctl.json`) to authenticate and use the Azure Speech Services.

Please note that the free tier (F0) has certain limitations, such as a maximum of 5 million characters per month. If you require higher limits, you may need to choose a paid pricing tier.

## Contributing

Contributions to voxctl are welcome! If you find any issues or have suggestions for improvements, please open an issue or submit a pull request on the GitHub repository.

## License

This project is licensed under the [MIT License](LICENSE).
