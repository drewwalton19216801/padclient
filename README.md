# Padserve Client

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE) [![Go](https://github.com/drewwalton19216801/padclient/actions/workflows/go.yml/badge.svg)](https://github.com/drewwalton19216801/padclient/actions/workflows/go.yml)

A secure messaging client that communicates over Tailscale, using the Bubble Tea TUI framework. This client is designed to work with the [Padserve secure messaging server](https://github.com/drewwalton19216801/padserve) and provides a terminal-based user interface for sending and receiving encrypted messages.

## Table of Contents

- [Features](#features)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Usage](#usage)
- [Commands](#commands)
- [Encryption Details](#encryption-details)
- [Project Structure](#project-structure)
- [Contributing](#contributing)
- [License](#license)
- [Acknowledgements](#acknowledgements)

## Features

- Secure communication over Tailscale networks.
- End-to-end encryption using AES and OTP (XOR cipher).
- Terminal-based user interface built with Bubble Tea.
- Operator support with special commands.
- Message broadcasting to all connected clients.

## Prerequisites

- [Go](https://golang.org/dl/) 1.23 or higher.
- [Tailscale](https://tailscale.com/) installed and connected.
- A Padserve [secure messaging server](https://github.com/drewwalton19216801/padserve) set up on your Tailscale network.
- Tailscale network configured with the server and clients.

## Installation

1. **Clone the Repository**

   ```sh
   git clone https://github.com/drewwalton19216801/padclient.git
   cd padclient
   ```

2. **Install Dependencies**

   Ensure you have the necessary Go packages installed:

   ```sh
   go mod tidy
   ```

## Usage

### Running the Client

```sh
go run . <YourID> <TailscaleServer>
```

- `<YourID>`: A unique identifier for your client (e.g., your username).
- `<TailscaleServer>`: The Tailscale IP address or hostname of the messaging server.

### Example

```sh
go run . Alice 100.101.102.103
```

### Connecting to Tailscale

Ensure you are connected to your Tailscale network before running the client:

```sh
tailscale up
```

## Commands

Once connected, you can use the following commands within the client:

- `SEND <RecipientID|ALL> <Message>`: Send a message to a specific client or broadcast to all clients.
- `HELP`: Display help information about available commands.
- `LIST`: List all connected clients.
- `SERVERHELP`: Display help information about the available server commands.
- `EXIT`: Exit the client program.

### Operator Commands

If you are the server operator, you may have access to additional commands (consult the server documentation for details):

- `KICK <ClientID>`: Remove a client from the server.
- `BAN <ClientID>`: Ban a client from the server.
- `UNBAN <ClientID>`: Remove a ban on a client.
- `LISTBANS`: List all banned clients.

## Encryption Details

- **Broadcast Messages**: Encrypted using AES with a shared secret derived from ECDH key exchange.
- **Direct Messages**: Encrypted using a One-Time Pad (OTP) generated for each message and XOR cipher.

### Key Exchange

- The client performs an ECDH key exchange with the server to establish a shared secret.
- The shared secret is hashed using SHA-256 to derive a symmetric key for AES encryption.

### Encryption Algorithms

- **AES Encryption**: Used for broadcasting messages to all clients securely.
- **OTP (XOR Cipher)**: Used for direct messages between two clients.

## Project Structure

- `main.go`: Initializes the client and handles the main loop using Bubble Tea.
- `client.go`: Manages client setup, registration, and key exchange with the server.
- `encryption.go`: Contains encryption functions for AES and XOR ciphers.
- `message_handler.go`: Reads and processes messages from the server.

## Contributing

Contributions are welcome! Please visit [CONTRIBUTING.md](docs/CONTRIBUTING.md) for more information.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE.txt) file for details.

## Acknowledgements

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) for the TUI framework.
- [Tailscale](https://tailscale.com/) for the secure network overlay.
- [tailutils](https://github.com/drewwalton19216801/tailutils) for Tailscale utilities.