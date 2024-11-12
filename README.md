# Padserve Client

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE) [![Go](https://github.com/drewwalton19216801/padclient/actions/workflows/go.yml/badge.svg)](https://github.com/drewwalton19216801/padclient/actions/workflows/go.yml)

A secure messaging client that communicates over Tailscale, using the Bubble Tea TUI framework. This client is designed to work with the [Padserve secure messaging server](https://github.com/drewwalton19216801/padserve) and provides a terminal-based user interface for sending and receiving encrypted messages.

## Table of Contents

- [Features](#features)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Usage](#usage)
- [Commands](#commands)
- [Command History](#command-history)
- [Key Shortcut Actions](#key-shortcut-actions)
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
- Cross-platform support across macOS, Linux, Windows, and (experimentally) FreeBSD.

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

## Command History

The client application includes a command history feature that allows you to navigate through your previously entered commands, similar to a typical terminal experience. This feature enhances productivity by enabling you to quickly reuse or edit past commands without retyping them entirely.

### How to Use Command History

- **Navigate Backward in History**:
  - **Up Arrow Key (`↑`)**: Press the Up arrow key to scroll backward through your command history. Each press will display the previous command in the input field.
- **Navigate Forward in History**:
  - **Down Arrow Key (`↓`)**: After scrolling backward, you can press the Down arrow key to move forward through the history. This allows you to return to more recent commands or to an empty input field.

### Editing Commands from History

- Once a previous command is displayed in the input field, you can edit it before executing.
- This is useful for sending similar messages or commands with slight modifications.

### Example

1. Type a command:

   ```
   SEND ALL Hello, everyone!
   ```

2. Press `Enter` to execute.
3. To resend the same message or modify it:

   - Press the Up arrow key to retrieve the command.
   - Edit the message if desired (e.g., change "Hello" to "Hi").
   - Press `Enter` to send the modified command.

## Key Shortcut Actions

The client application supports several key shortcuts to improve navigation and efficiency. Below is a list of available key shortcuts and their actions.

### Input and Command History Navigation

- **Up Arrow Key (`↑`)**:
  - **Action**: Navigate backward through the command history.
  - **Usage**: Retrieve previous commands to reuse or edit them.
- **Down Arrow Key (`↓`)**:
  - **Action**: Navigate forward through the command history.
  - **Usage**: Move toward more recent commands or return to an empty input field.

### Message Viewport Scrolling

- **Scroll Up**:
  - **Keys**:
    - **Page Up (`PgUp`)**
    - **Control + U (`Ctrl+U`)**
  - **Action**: Scroll up through the message history in the viewport.
  - **Usage**: View earlier messages that have scrolled off the screen.
- **Scroll Down**:
  - **Keys**:
    - **Page Down (`PgDn`)**
    - **Control + D (`Ctrl+D`)**
  - **Action**: Scroll down through the message history.
  - **Usage**: Return to more recent messages after scrolling up.
- **Jump to Top**:
  - **Key**:
    - **Home**
  - **Action**: Jump to the very top of the message history.
  - **Usage**: Quickly view the earliest messages in the session.
- **Jump to Bottom**:
  - **Key**:
    - **End**
  - **Action**: Jump to the bottom of the message history.
  - **Usage**: Return to the most recent messages.

### General Shortcuts

- **Submit Command**:
  - **Key**:
    - **Enter**
  - **Action**: Submit the command or message typed in the input field.
  - **Usage**: Execute commands like `SEND`, `HELP`, or `EXIT`.
- **Exit Application**:
  - **Keys**:
    - **Control + C (`Ctrl+C`)**
    - **Escape (`Esc`)**
  - **Action**: Exit the client application gracefully.
  - **Usage**: Close the application when you are done or need to disconnect.

### Notes

- **Typing New Commands**:
  - When you start typing a new command (i.e., any printable character), the command history navigation resets. This means that pressing the Up arrow key will start from the most recent command again.
- **Editing Input**:
  - Standard text editing keys work within the input field (e.g., Left/Right arrows to move the cursor, Backspace to delete characters).

### Example Usage of Key Shortcuts

- **Scrolling Messages**:
  - To read an earlier message:
    - Press `Ctrl+U` or `Page Up` to scroll up.
    - Continue pressing to scroll further back.
  - To return to the latest messages:
    - Press `Ctrl+D` or `Page Down` to scroll down.
    - Press `End` to jump directly to the bottom.

- **Navigating Command History**:
  - After executing several commands:
    - Press the Up arrow key to access the last command.
    - Press Up again to go further back.
    - Press Down to navigate forward in the history.
    - Edit the retrieved command if needed before executing.

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
