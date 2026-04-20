package handler

import (
	"bufio"
	"context"
	"io"
	"log"
	"net"
	"time"

	"redis-like-golang/internal/adapter/protocol"
	"redis-like-golang/internal/domain/command"
	"redis-like-golang/internal/usecase"
)

// TCPHandler handles TCP connections
type TCPHandler struct {
	commandHandler *usecase.CommandHandler
	parser         *protocol.Parser
}

// NewTCPHandler creates a new TCP handler
func NewTCPHandler(commandHandler *usecase.CommandHandler, parser *protocol.Parser) *TCPHandler {
	return &TCPHandler{
		commandHandler: commandHandler,
		parser:         parser,
	}
}

// HandleConnection handles a single TCP connection
func (h *TCPHandler) HandleConnection(conn net.Conn) {
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("Error closing connection: %v", err)
		}
	}()

	// Create context for this connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Increment connection counter (if we had access to stats, we'd do it here)
	// For now, this is handled at a higher level if needed

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		// Check context cancellation
		if ctx.Err() != nil {
			return
		}

		line := scanner.Text()
		if line == "" {
			continue
		}

		// Parse command
		cmd, err := h.parser.ParseCommand(line)
		if err != nil {
			response := h.parser.FormatError(err.Error())
			h.writeResponse(conn, response)
			continue
		}

		// Handle QUIT command
		if cmd.Type == command.QUIT {
			h.writeResponse(conn, h.parser.FormatOK())
			return
		}

		// Execute command with context
		response := h.commandHandler.ExecuteCommand(ctx, cmd)
		h.writeResponse(conn, response)
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		log.Printf("Error reading from connection: %v", err)
	}
}

func (h *TCPHandler) writeResponse(conn net.Conn, response string) {
	data := []byte(response + "\n")
	n, err := conn.Write(data)
	if err != nil {
		log.Printf("error writing response: %v", err)
		return
	}
	if n != len(data) {
		log.Printf("incomplete write: wrote %d of %d bytes", n, len(data))
	}
}
