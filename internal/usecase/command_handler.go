package usecase

import (
	"context"
	"fmt"
	"redis-like-golang/internal/adapter/protocol"
	"redis-like-golang/internal/domain/command"
	"redis-like-golang/internal/domain/repository"
	"strconv"
	"strings"
)

type CommandHandler struct {
	store   repository.KeyValueRepository
	persist repository.PersistenceRepository
	parser  *protocol.Parser
	stats   *Stats
}

func NewCommandHandler(
	parser *protocol.Parser,
	store repository.KeyValueRepository,
	persist repository.PersistenceRepository,
	stats *Stats) *CommandHandler {
	return &CommandHandler{
		store:   store,
		persist: persist,
		parser:  parser,
		stats:   stats,
	}
}

func (h *CommandHandler) ExecuteCommand(ctx context.Context, cmd *protocol.Command) string {
	h.stats.IncrementalCommands()

	switch cmd.Type {
	case command.SET:
		return h.handleSet(ctx, cmd.Args)
	case command.GET:
		return h.handleGet(ctx, cmd.Args)
	case command.DEL:
		return h.handleDel(ctx, cmd.Args)
	case command.EXPIRE:
		return h.handleExpire(ctx, cmd.Args)
	case command.TTL:
		return h.handleTTL(ctx, cmd.Args)
	case command.PERSIST:
		return h.handlePersist(ctx, cmd.Args)
	case command.KEYS:
		return h.handleKeys(ctx, cmd.Args)
	case command.PING:
		return h.handlePing(ctx, cmd.Args)
	case command.INFO:
		return h.handleInfo(ctx, cmd.Args)
	default:
		return h.parser.FormatError(fmt.Sprintf("unknown command type: %s", cmd.Type))
	}
}

func (h *CommandHandler) handleSet(ctx context.Context, args []string) string {
	if len(args) != 2 {
		return h.parser.FormatError("SET requires at least 2 arguments")
	}

	key := args[0]
	value := strings.Join(args[1:], " ")
	h.store.Set(ctx, key, value)

	if h.persist != nil {
		h.persist.Append(ctx, command.SET.String(), args)
	}

	return h.parser.FormatOK()
}

func (h *CommandHandler) handleGet(ctx context.Context, args []string) string {
	if len(args) != 1 {
		return h.parser.FormatError("GET requires at least 1 arguments")
	}

	value, found := h.store.Get(ctx, args[0])
	if found {
		return value
	}

	return h.parser.FormatNil()
}

func (h *CommandHandler) handleDel(ctx context.Context, args []string) string {
	if len(args) != 1 {
		return h.parser.FormatError("DEL requires at least 1 arguments")
	}

	count := 0
	for _, keys := range args {
		count += h.store.Del(ctx, keys)
	}

	if h.persist != nil && count > 0 {
		for _, key := range args {
			h.persist.Append(ctx, command.DEL.String(), []string{key})
		}
	}

	return h.parser.FormatResponse(count)
}

func (h *CommandHandler) handleExists(ctx context.Context, args []string) string {
	if len(args) != 1 {
		return h.parser.FormatError("EXISTS requires at least 1 arguments")
	}

	count := 0
	for _, key := range args {
		if h.store.Exists(ctx, key) {
			count++
		}
	}

	return h.parser.FormatResponse(count)
}

func (h *CommandHandler) handleExpire(ctx context.Context, args []string) string {
	if len(args) != 1 {
		return h.parser.FormatError("EXPIRE requires at least 1 arguments")
	}

	key := args[0]
	seconds, err := strconv.Atoi(args[1])
	if err != nil {
		return h.parser.FormatError("invalid seconds value")
	}

	success := h.store.Expire(ctx, key, seconds)
	if success {
		if h.persist != nil {
			h.persist.Append(ctx, command.EXPIRE.String(), []string{key})
		}
		return h.parser.FormatOK()
	}

	return h.parser.FormatResponse(success)
}

func (h *CommandHandler) handlePersist(ctx context.Context, args []string) string {
	if len(args) != 1 {
		return h.parser.FormatError("PERSIST requires at least 1 arguments")
	}

	success := h.store.Persist(ctx, args[0])
	if success {
		if h.persist != nil {
			h.persist.Append(ctx, command.PERSIST.String(), []string{args[0]})
		}
		return h.parser.FormatOK()
	}

	return h.parser.FormatResponse(success)
}

func (h *CommandHandler) handleTTL(ctx context.Context, args []string) string {
	if len(args) != 1 {
		return h.parser.FormatError("TTL requires at least 1 arguments")
	}

	ttl := h.store.TTL(ctx, args[0])
	return h.parser.FormatResponse(ttl)
}

func (h *CommandHandler) handleKeys(ctx context.Context, args []string) string {
	pattern := "*"
	if len(args) > 0 {
		pattern = args[0]
	}

	keys := h.store.Keys(ctx, pattern)

	if len(keys) == 0 {
		return ""
	}

	return strings.Join(keys, " ")
}

func (h *CommandHandler) handlePing(ctx context.Context, args []string) string {
	message := "PONG"
	if len(args) > 0 {
		message = strings.Join(args, " ")
	}
	return message
}

func (h *CommandHandler) handleInfo(ctx context.Context, args []string) string {
	return h.stats.GetInfo(ctx)
}
