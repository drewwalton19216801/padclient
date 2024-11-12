// main.go
// Package main initializes the client, handles user input, and manages the main loop of the application using Bubble Tea TUI.
// This version includes a command history feature, allowing users to navigate through previous commands using Up/Down arrows.

package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput" // Text input component
	"github.com/charmbracelet/bubbles/viewport"  // Viewport component for scrolling messages
	tea "github.com/charmbracelet/bubbletea"     // Bubble Tea TUI framework
	"github.com/drewwalton19216801/tailutils"    // Utilities for Tailscale
)

var (
	address string // Server address
)

// Define message types used in the Bubble Tea program
type errMsg struct{ error }
type connectedMsg struct {
	conn         net.Conn
	hashedSecret []byte
	isOperator   bool
}
type serverMsg struct {
	content string
}
type operatorMsg struct {
	content string
}
type kickedMsg struct{}
type bannedMsg struct{}
type disconnectMsg struct{}
type incomingMessage struct {
	senderID    string
	content     string
	isBroadcast bool
}

// Model represents the application's state
type model struct {
	isOperator   bool            // Operator status
	clientID     string          // Client identifier
	conn         net.Conn        // Network connection
	input        textinput.Model // Text input component for user commands
	viewport     viewport.Model  // Viewport for displaying messages
	messages     []string        // All messages to display in the viewport
	history      []string        // Command history
	historyIndex int             // Current index in the history (-1 means not navigating)
	hashedSecret []byte          // Hashed secret for AES encryption
	messageChan  chan tea.Msg    // Channel for incoming messages from the server
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run main.go <YourID> <TailscaleServer>")
		return
	}
	clientID := os.Args[1]
	serverIP := os.Args[2]
	address = serverIP + ":12345"

	// Check if the local IP address belongs to a Tailscale interface
	isTailscale, err := tailutils.HasTailscaleIP()
	if err != nil {
		fmt.Printf("Error checking local IP address: %v\n", err)
		return
	}
	if !isTailscale {
		fmt.Println("Please connect to a Tailscale network.")
		return
	}

	m := &model{
		clientID:     clientID,
		historyIndex: -1, // Initialize history index
	}

	// Initialize the Bubble Tea program with the model
	p := tea.NewProgram(m)
	if err := p.Start(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

// Init initializes the model and starts the connection to the server
func (m *model) Init() tea.Cmd {
	// Initialize the text input component
	m.input = textinput.New()
	m.input.Placeholder = "Type a command"
	m.input.CharLimit = 256
	m.input.Width = 50
	m.updatePrompt() // Set the initial prompt with client ID and operator status
	m.input.Focus()

	// Initialize the viewport for displaying messages
	m.viewport = viewport.New(80, 20) // Width and Height of the viewport
	m.viewport.YPosition = 0
	m.viewport.HighPerformanceRendering = false      // Set to true if flickering occurs
	m.viewport.SetContent("Connecting to server...") // Initial content

	return tea.Batch(
		connectToServer(m.clientID),
		textinput.Blink, // Start blinking cursor
	)
}

// Update handles incoming events (keyboard input, server messages, etc.)
func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle key presses for input and viewport scrolling
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			// Exit the program on Ctrl+C or Esc
			if m.conn != nil {
				m.conn.Close()
			}
			return m, tea.Quit
		case tea.KeyEnter:
			// Handle command input when Enter is pressed
			input := strings.TrimSpace(m.input.Value())
			m.input.SetValue("")
			return m.handleInput(input)
		case tea.KeyUp:
			// Navigate command history backward
			if len(m.history) > 0 {
				if m.historyIndex == -1 {
					m.historyIndex = len(m.history) - 1
				} else if m.historyIndex > 0 {
					m.historyIndex--
				}
				m.input.SetValue(m.history[m.historyIndex])
				m.input.CursorEnd()
			}
		case tea.KeyDown:
			// Navigate command history forward
			if len(m.history) > 0 && m.historyIndex != -1 {
				if m.historyIndex < len(m.history)-1 {
					m.historyIndex++
					m.input.SetValue(m.history[m.historyIndex])
				} else {
					m.historyIndex = -1
					m.input.SetValue("")
				}
				m.input.CursorEnd()
			}
		case tea.KeyPgUp, tea.KeyCtrlU:
			// Scroll viewport up
			m.viewport.LineUp(1)
		case tea.KeyPgDown, tea.KeyCtrlD:
			// Scroll viewport down
			m.viewport.LineDown(1)
		case tea.KeyHome:
			// Go to top of the viewport
			m.viewport.GotoTop()
		case tea.KeyEnd:
			// Go to bottom of the viewport
			m.viewport.GotoBottom()
		default:
			// Update text input component
			m.input, cmd = m.input.Update(msg)
			// Reset history index when typing a new command
			if msg.String() != "" && msg.Runes != nil {
				m.historyIndex = -1
			}
		}
		return m, cmd
	case connectedMsg:
		// Handle successful connection to the server
		m.conn = msg.conn
		m.hashedSecret = msg.hashedSecret
		m.isOperator = msg.isOperator
		m.updatePrompt() // Update the prompt to reflect operator status
		m.messageChan = make(chan tea.Msg)
		go readMessages(m.conn, m.hashedSecret, m.messageChan)
		m.appendMessage("Connected to the server. Type your commands below:")
		if m.isOperator {
			m.appendMessage("You are the server operator. Type HELP to see available commands.")
		} else {
			m.appendMessage("Type HELP to see available commands.")
		}
		return m, waitForServerMessage(m.messageChan)
	case serverMsg:
		// Handle general messages from the server
		m.appendMessage(msg.content)
		return m, waitForServerMessage(m.messageChan)
	case operatorMsg:
		// Handle operator status change
		m.isOperator = true
		m.updatePrompt() // Update the prompt since operator status changed
		m.appendMessage(msg.content)
		return m, waitForServerMessage(m.messageChan)
	case incomingMessage:
		// Handle incoming messages from other clients
		var prefix string
		if msg.isBroadcast {
			prefix = fmt.Sprintf("Broadcast from %s: ", msg.senderID)
		} else {
			prefix = fmt.Sprintf("Message from %s: ", msg.senderID)
		}
		m.appendMessage(prefix + msg.content)
		return m, waitForServerMessage(m.messageChan)
	case kickedMsg:
		// Handle being kicked by the operator
		m.appendMessage("You have been kicked from the server by the operator.")
		if m.conn != nil {
			m.conn.Close()
		}
		return m, tea.Quit
	case bannedMsg:
		// Handle being banned by the operator
		m.appendMessage("You have been banned from the server by the operator.")
		if m.conn != nil {
			m.conn.Close()
		}
		return m, tea.Quit
	case disconnectMsg:
		// Handle disconnection from the server
		m.appendMessage("Disconnected from server.")
		if m.conn != nil {
			m.conn.Close()
		}
		return m, tea.Quit
	case errMsg:
		// Handle errors
		m.appendMessage(fmt.Sprintf("Error: %v", msg.error))
		if m.conn != nil {
			m.conn.Close()
		}
		return m, tea.Quit
	default:
		return m, nil
	}
}

// View renders the UI
func (m *model) View() string {
	return fmt.Sprintf(
		"%s\n%s",
		m.viewport.View(), // Render the viewport above
		m.input.View(),    // Render the input field below
	)
}

// handleInput processes the user input commands
func (m *model) handleInput(input string) (tea.Model, tea.Cmd) {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		// Reset history index if the input is empty
		m.historyIndex = -1
		return m, nil
	}

	// Add the command to history if it's not empty
	if input != "" {
		m.history = append(m.history, input)
	}
	m.historyIndex = -1 // Reset history index

	switch parts[0] {
	case "SEND":
		// Handle the SEND command to send messages
		if len(parts) < 3 {
			m.appendMessage("Invalid SEND command. Use: SEND <RecipientID|ALL> <Message>")
			return m, nil
		}
		recipientID := parts[1]
		messageText := strings.Join(parts[2:], " ")
		if recipientID == "ALL" {
			// Encrypt the message using AES with the shared secret
			encryptedData, err := encryptAES(m.hashedSecret, []byte(messageText))
			if err != nil {
				m.appendMessage(fmt.Sprintf("Error encrypting message: %v", err))
				return m, nil
			}
			// Encode the encrypted data in hex
			encryptedDataHex := hex.EncodeToString(encryptedData)
			// Send the encrypted message to the server
			fmt.Fprintf(m.conn, "SEND ALL %s\n", encryptedDataHex)
		} else {
			// Generate a one-time pad (OTP) key
			key := make([]byte, len(messageText))
			_, err := rand.Read(key)
			if err != nil {
				m.appendMessage(fmt.Sprintf("Error generating OTP key: %v", err))
				return m, nil
			}

			// Encrypt the message using XOR cipher
			plaintext := []byte(messageText)
			ciphertext := encryptXOR(plaintext, key)

			// Encode key and ciphertext in hex
			keyHex := hex.EncodeToString(key)
			ciphertextHex := hex.EncodeToString(ciphertext)

			// Send the encrypted message in the format: SEND <RecipientID> <key_hex>|<ciphertext_hex>
			encryptedData := keyHex + "|" + ciphertextHex
			fmt.Fprintf(m.conn, "SEND %s %s\n", recipientID, encryptedData)
		}
		return m, nil
	case "HELP":
		// Display help text
		m.appendMessage("Available commands:")
		m.appendMessage("SEND <RecipientID|ALL> <Message> - Send a message")
		m.appendMessage("HELP - Print this help text")
		m.appendMessage("EXIT - Exit the program")
		return m, nil
	case "EXIT":
		// Exit the client program
		fmt.Fprintf(m.conn, "EXIT\n")
		if m.conn != nil {
			m.conn.Close()
		}
		return m, tea.Quit
	default:
		// Pass other commands to the server
		fmt.Fprintf(m.conn, "%s\n", input)
		return m, nil
	}
}

// updatePrompt updates the prompt with the client ID and operator status
func (m *model) updatePrompt() {
	if m.isOperator {
		m.input.Prompt = fmt.Sprintf("%s (op) > ", m.clientID)
	} else {
		m.input.Prompt = fmt.Sprintf("%s > ", m.clientID)
	}
}

// appendMessage adds a message to the viewport and updates the content
func (m *model) appendMessage(msg string) {
	m.messages = append(m.messages, msg)
	content := strings.Join(m.messages, "\n")
	m.viewport.SetContent(content)
	m.viewport.GotoBottom() // Scroll to the bottom to show the new message
}

// connectToServer establishes the connection and performs client setup
func connectToServer(clientID string) tea.Cmd {
	return func() tea.Msg {
		conn, err := net.Dial("tcp", address)
		if err != nil {
			return errMsg{err}
		}
		hashedSecret, isOperator, err := setupClient(conn, clientID)
		if err != nil {
			return errMsg{err}
		}
		return connectedMsg{conn: conn, hashedSecret: hashedSecret, isOperator: isOperator}
	}
}

// waitForServerMessage waits for a message from the server
func waitForServerMessage(messageChan <-chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return <-messageChan
	}
}
