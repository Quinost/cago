package resp2

import (
	"cago/internal"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	ERRWrongArgumentType = "ERR wrong argument type"
	ERRSyntexError       = "ERR syntax error"

	OK = "OK"
)

type RESPHandler struct {
	cachesrv *internal.CacheService
}

func NewRESPHandler(cachesrv *internal.CacheService) *RESPHandler {
	return &RESPHandler{
		cachesrv: cachesrv,
	}
}

func (h *RESPHandler) HandleCommand(cmd *Value, writer *RESPWriter) error {
	if cmd.Type != Array || len(cmd.Array) == 0 {
		return writer.WriteError("ERR invalid command format")
	}

	if cmd.Array[0].Type != BulkString {
		return writer.WriteError("ERR command must be a bulk string")
	}

	command := strings.ToUpper(cmd.Array[0].Bulk)
	args := cmd.Array[1:]

	switch command {
	case "PING":
		return h.handlePing(args, writer)
	case "SET":
		return h.handleSet(args, writer)
	case "GET":
		return h.handleGet(args, writer)
	default:
		return writer.WriteError(fmt.Sprintf("ERR unknown command '%s'", command))
	}
}

func (h *RESPHandler) handlePing(args []Value, writer *RESPWriter) error {
	if len(args) == 0 {
		return writer.WriteSimpleString("PONG")
	}

	if args[0].Type != BulkString {
		return writer.WriteError(ERRWrongArgumentType)
	}

	return writer.WriteBulkString(args[0].Bulk)
}

func (h *RESPHandler) handleSet(args []Value, writer *RESPWriter) error {
	if len(args) < 2 {
		return writer.WriteError("ERR wrong number of arguments for 'SET' command")
	}

	if args[0].Type != BulkString || args[1].Type != BulkString {
		return writer.WriteError(ERRWrongArgumentType)
	}

	key := args[0].Bulk
	value := args[1].Bulk
	var ttl time.Duration

	for i := 2; i < len(args); i++ {
		if args[i].Type != BulkString {
			return writer.WriteError(ERRSyntexError)
		}

		option := strings.ToUpper(args[i].Bulk)

		if option == "EX" {
			if i+1 >= len(args) {
				return writer.WriteError(ERRSyntexError)
			}

			if args[i+1].Type != BulkString {
				return writer.WriteError("ERR value is not an integer or out of range")
			}

			seconds, err := strconv.Atoi(args[i+1].Bulk)
			if err != nil {
				return writer.WriteError("ERR value is not an integer or out of range")
			}

			ttl = time.Duration(seconds) * time.Second
			i++
		}
	}

	err := h.cachesrv.Set(key, value, ttl)
	if err != nil {
		return writer.WriteError(fmt.Sprintf("ERR %s", err.Error()))
	}

	return writer.WriteSimpleString(OK)
}

func (h *RESPHandler) handleGet(args []Value, writer *RESPWriter) error {
	if len(args) != 1 {
		return writer.WriteError("ERR wrong number of arguments for 'GET' command")
	}

	if args[0].Type != BulkString {
		return writer.WriteError("ERR wrong argument type")
	}

	key := args[0].Bulk
	value, exists, err := h.cachesrv.Get(key)
	if err != nil {
		return writer.WriteError(fmt.Sprintf("ERR %s", err.Error()))
	}
	if !exists {
		return writer.WriteNull()
	}
	return writer.WriteBulkString(value)
}