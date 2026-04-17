package handler

import (
	"bufio"
	"context"
	"io"
	"log"
	"net"
	"redis-like-golang/internal/adapter/protocol"
	"redis-like-golang/internal/domain/command"
	"redis-like-golang/internal/usecase"
	"time"
)

type TCPHandler struct {
	commandHandler *usecase.CommandHandler
	parser         *protocol.Parser
}

func NewTCPHandler(commandHandler *usecase.CommandHandler, parser *protocol.Parser) *TCPHandler {
	return &TCPHandler{
		commandHandler: commandHandler,
		parser:         parser,
	}
}

func (h *TCPHandler) HandleConnection(conn net.Conn) {
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		if ctx.Err() != nil {
			return
		}

		line := scanner.Text()
		if line == "" {
			continue
		}

		cmd, err := h.parser.ParserCommand(line)
		if err != nil {
			response := h.parser.FormatError(err.Error())
			h.writeResponse(conn, response)
			continue
		}

		if cmd.Type == command.QUIT {
			h.writeResponse(conn, h.parser.FormatOK())
			return
		}

		response := h.commandHandler.ExecuteCommand(ctx, cmd)
		h.writeResponse(conn, response)
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		log.Printf("Error reading from connection: %v", err)
	}
}

func (h *TCPHandler) writeResponse(conn net.Conn, response string) {
	conn.Write([]byte(response + "\n"))
}
