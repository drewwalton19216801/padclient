// main.go
// Package main initializes the client, handles user input, and manages the main loop of the application using bubbletea TUI.
package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/drewwalton19216801/tailutils"
)

var (
	address string // Server address
)

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

type model struct {
	isOperator   bool
	clientID     string
	conn         net.Conn
	input        textinput.Model
	messages     []string
	hashedSecret []byte
	messageChan  chan tea.Msg
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
		clientID: clientID,
	}

	p := tea.NewProgram(m)
	if err := p.Start(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func (m *model) Init() tea.Cmd {
	m.input = textinput.New()
	m.input.Placeholder = "Type a command"
	m.input.Focus()
	m.input.CharLimit = 256
	m.input.Width = 50
	return tea.Batch(
		connectToServer(m.clientID),
		textinput.Blink,
	)
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) { // Corrected type switch
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			if m.conn != nil {
				m.conn.Close()
			}
			return m, tea.Quit
		case tea.KeyEnter:
			input := strings.TrimSpace(m.input.Value())
			m.input.SetValue("")
			return m.handleInput(input)
		default:
			m.input, cmd = m.input.Update(msg)
			return m, cmd
		}
	case connectedMsg:
		m.conn = msg.conn
		m.hashedSecret = msg.hashedSecret
		m.isOperator = msg.isOperator
		m.messageChan = make(chan tea.Msg)
		go readMessages(m.conn, m.hashedSecret, m.messageChan)
		m.messages = append(m.messages, "Connected to the server. Type your commands below:")
		if m.isOperator {
			m.messages = append(m.messages, "You are the server operator. Type HELP to see available commands.")
		} else {
			m.messages = append(m.messages, "Type HELP to see available commands.")
		}
		return m, waitForServerMessage(m.messageChan)
	case serverMsg:
		m.messages = append(m.messages, msg.content)
		return m, waitForServerMessage(m.messageChan)
	case operatorMsg:
		m.isOperator = true
		m.messages = append(m.messages, msg.content)
		return m, waitForServerMessage(m.messageChan)
	case incomingMessage:
		var prefix string
		if msg.isBroadcast {
			prefix = fmt.Sprintf("Broadcast from %s: ", msg.senderID)
		} else {
			prefix = fmt.Sprintf("Message from %s: ", msg.senderID)
		}
		m.messages = append(m.messages, prefix+msg.content)
		return m, waitForServerMessage(m.messageChan)
	case kickedMsg:
		m.messages = append(m.messages, "You have been kicked from the server by the operator.")
		if m.conn != nil {
			m.conn.Close()
		}
		return m, tea.Quit
	case bannedMsg:
		m.messages = append(m.messages, "You have been banned from the server by the operator.")
		if m.conn != nil {
			m.conn.Close()
		}
		return m, tea.Quit
	case disconnectMsg:
		m.messages = append(m.messages, "Disconnected from server.")
		if m.conn != nil {
			m.conn.Close()
		}
		return m, tea.Quit
	case errMsg:
		m.messages = append(m.messages, fmt.Sprintf("Error: %v", msg.error))
		if m.conn != nil {
			m.conn.Close()
		}
		return m, tea.Quit
	default:
		return m, nil
	}
}

func (m *model) View() string {
	s := strings.Builder{}
	for _, msg := range m.messages {
		s.WriteString(msg + "\n")
	}
	s.WriteString("\n")
	s.WriteString(m.input.View())
	s.WriteString("\n")
	return s.String()
}

func (m *model) handleInput(input string) (tea.Model, tea.Cmd) {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return m, nil
	}

	switch parts[0] {
	case "SEND":
		if len(parts) < 3 {
			m.messages = append(m.messages, "Invalid SEND command. Use: SEND <RecipientID|ALL> <Message>")
			return m, nil
		}
		recipientID := parts[1]
		messageText := strings.Join(parts[2:], " ")
		if recipientID == "ALL" {
			// Encrypt the message using AES with the shared secret
			encryptedData, err := encryptAES(m.hashedSecret, []byte(messageText))
			if err != nil {
				m.messages = append(m.messages, fmt.Sprintf("Error encrypting message: %v", err))
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
				m.messages = append(m.messages, fmt.Sprintf("Error generating OTP key: %v", err))
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
		m.messages = append(m.messages, "Available commands:")
		m.messages = append(m.messages, "SEND <RecipientID|ALL> <Message> - Send a message")
		m.messages = append(m.messages, "HELP - Print this help text")
		m.messages = append(m.messages, "EXIT - Exit the program")
		return m, nil
	case "EXIT":
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

func waitForServerMessage(messageChan <-chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return <-messageChan
	}
}
