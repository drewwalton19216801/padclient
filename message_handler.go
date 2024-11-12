// message_handler.go
// Package main handles reading and processing messages from the server.

package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"net"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// readMessages continuously reads messages from the server and processes them.
func readMessages(conn net.Conn, hashedSecret []byte, messageChan chan<- tea.Msg) {
	reader := bufio.NewReader(conn)
	var inMultiLineResponse bool = false
	var multiLineBuffer []string

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			messageChan <- disconnectMsg{}
			return
		}
		message = strings.TrimRight(message, "\r\n")

		if message == "" {
			continue
		}

		// Handle being registered as operator
		if message == "REGISTERED as operator" {
			messageChan <- operatorMsg{content: "You are registered as the server operator."}
			continue
		}

		// Handle being kicked
		if message == "KICKED You have been kicked by the operator" {
			messageChan <- kickedMsg{}
			return
		}

		// Handle being banned
		if message == "BANNED You have been banned by the operator" {
			messageChan <- bannedMsg{}
			return
		}

		// Detect the start of a multi-line response
		if message == "BEGIN_RESPONSE" {
			inMultiLineResponse = true
			multiLineBuffer = []string{}
			continue // Skip printing the marker
		}

		// Detect the end of a multi-line response
		if message == "END_RESPONSE" {
			inMultiLineResponse = false
			messageChan <- serverMsg{content: strings.Join(multiLineBuffer, "\n")}
			continue // Skip printing the marker
		}

		if inMultiLineResponse {
			multiLineBuffer = append(multiLineBuffer, message)
			continue
		}

		// Handle incoming messages from other clients
		if strings.HasPrefix(message, "MESSAGE from") || strings.HasPrefix(message, "BROADCAST from") {
			parts := strings.SplitN(message, ": ", 2)
			if len(parts) != 2 {
				messageChan <- serverMsg{content: "Invalid message format. Ignoring."}
				continue
			}
			senderInfo := parts[0]
			encryptedData := parts[1]

			// Extract sender ID
			var senderID string
			isBroadcast := false
			if strings.HasPrefix(senderInfo, "MESSAGE from") {
				senderID = strings.TrimPrefix(senderInfo, "MESSAGE from ")
			} else if strings.HasPrefix(senderInfo, "BROADCAST from") {
				senderID = strings.TrimPrefix(senderInfo, "BROADCAST from ")
				isBroadcast = true
			}

			if isBroadcast {
				if strings.Contains(encryptedData, "|") {
					// Encrypted data format: key_hex|ciphertext_hex
					dataParts := strings.SplitN(encryptedData, "|", 2)
					if len(dataParts) != 2 {
						messageChan <- serverMsg{content: fmt.Sprintf("Invalid broadcast message format from %s. Ignoring.", senderID)}
						continue
					}
					keyHex := dataParts[0]
					ciphertextHex := dataParts[1]

					// Decode hex strings
					key, err := hex.DecodeString(keyHex)
					if err != nil {
						messageChan <- serverMsg{content: fmt.Sprintf("Error decoding key from broadcast from %s: %v", senderID, err)}
						continue
					}
					ciphertext, err := hex.DecodeString(ciphertextHex)
					if err != nil {
						messageChan <- serverMsg{content: fmt.Sprintf("Error decoding ciphertext from broadcast from %s: %v", senderID, err)}
						continue
					}

					// Decrypt the message using XOR cipher
					if len(key) != len(ciphertext) {
						messageChan <- serverMsg{content: fmt.Sprintf("Key and ciphertext lengths do not match in broadcast from %s.", senderID)}
						continue
					}
					plaintext := encryptXOR(ciphertext, key)
					messageChan <- incomingMessage{
						senderID:    senderID,
						content:     string(plaintext),
						isBroadcast: true,
					}
				} else {
					// Decrypt broadcast message using AES
					ciphertext, err := hex.DecodeString(encryptedData)
					if err != nil {
						messageChan <- serverMsg{content: fmt.Sprintf("Error decoding broadcast from %s: %v", senderID, err)}
						continue
					}
					plaintext, err := decryptAES(hashedSecret, ciphertext)
					if err != nil {
						messageChan <- serverMsg{content: fmt.Sprintf("Error decrypting broadcast from %s: %v", senderID, err)}
						continue
					}
					messageChan <- incomingMessage{
						senderID:    senderID,
						content:     string(plaintext),
						isBroadcast: true,
					}
				}
			} else {
				// Encrypted data format: key_hex|ciphertext_hex
				dataParts := strings.SplitN(encryptedData, "|", 2)
				if len(dataParts) != 2 {
					messageChan <- serverMsg{content: fmt.Sprintf("Invalid message format from %s. Ignoring.", senderID)}
					continue
				}
				keyHex := dataParts[0]
				ciphertextHex := dataParts[1]

				// Decode hex strings
				key, err := hex.DecodeString(keyHex)
				if err != nil {
					messageChan <- serverMsg{content: fmt.Sprintf("Error decoding key from %s: %v", senderID, err)}
					continue
				}
				ciphertext, err := hex.DecodeString(ciphertextHex)
				if err != nil {
					messageChan <- serverMsg{content: fmt.Sprintf("Error decoding ciphertext from %s: %v", senderID, err)}
					continue
				}

				// Decrypt the message using XOR cipher
				if len(key) != len(ciphertext) {
					messageChan <- serverMsg{content: fmt.Sprintf("Key and ciphertext lengths do not match from %s.", senderID)}
					continue
				}
				plaintext := encryptXOR(ciphertext, key)
				messageChan <- incomingMessage{
					senderID:    senderID,
					content:     string(plaintext),
					isBroadcast: false,
				}
			}
		} else {
			// Handle other server messages
			messageChan <- serverMsg{content: message}
		}
	}
}
