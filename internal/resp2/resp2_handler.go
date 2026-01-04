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
	case "DEL":
		return h.handleDel(args, writer)
	case "EXISTS":
		return h.handleExists(args, writer)
	case "EXPIRE":
		return h.handleExpire(args, writer)
	case "TTL":
		return h.handleTTL(args, writer)
	case "KEYS":
		return h.handleKeys(args, writer)
	default:
		return writer.WriteError(fmt.Sprintf("ERR unknown command '%s'", command))
	}
}

// RESP: *1\r\n$4\r\nPING\r\n
// RESP: *2\r\n$4\r\nPING\r\n$5\r\nhello\r\n
// Pattern: PING [message]
// Example: PING → PONG
// Example: PING "hello" → "hello"
func (h *RESPHandler) handlePing(args []Value, writer *RESPWriter) error {
	if len(args) == 0 {
		return writer.WriteSimpleString("PONG")
	}

	if args[0].Type != BulkString {
		return writer.WriteError(ERRWrongArgumentType)
	}

	return writer.WriteBulkString(args[0].Bulk)
}

// RESP: *3\r\n$3\r\nSET\r\n$5\r\nmykey\r\n$5\r\nhello\r\n
// RESP: *5\r\n$3\r\nSET\r\n$5\r\nmykey\r\n$5\r\nhello\r\n$2\r\nEX\r\n$2\r\n10\r\n
// Pattern: SET key value [EX seconds]
// Example: SET mykey "hello" → OK
// Example: SET mykey "hello" EX 10 → OK (expires in 10s)
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
		return writer.WriteError(formatError(err))
	}

	return writer.WriteSimpleString("OK")
}

// RESP: *2\r\n$3\r\nGET\r\n$5\r\nmykey\r\n
// Pattern: GET key
// Example: GET mykey → "hello"
// Example: GET nonexistent → (nil)
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
		return writer.WriteError(formatError(err))
	}
	if !exists {
		return writer.WriteNull()
	}
	return writer.WriteBulkString(value)
}

// RESP: *2\r\n$3\r\nDEL\r\n$4\r\nkey1\r\n
// RESP: *4\r\n$3\r\nDEL\r\n$4\r\nkey1\r\n$4\r\nkey2\r\n$4\r\nkey3\r\n
// Pattern: DEL key [key ...]
// Example: DEL key1 → 1 (deleted)
// Example: DEL key1 key2 nonexistent → 2 (deleted 2 out of 3)
// Returns: number of keys deleted
func (h *RESPHandler) handleDel(args []Value, writer *RESPWriter) error {
	if len(args) < 1 {
		return writer.WriteError("ERR wrong number of arguments for 'DEl' commanmd")
	}

	deleted := int64(0)

	for _, arg := range args {
		if arg.Type != BulkString {
			return writer.WriteError(ERRWrongArgumentType)
		}

		success, err := h.cachesrv.Delete(arg.Bulk)
		if err != nil {
			return writer.WriteError(formatError(err))
		}

		if success {
			deleted++
		}
	}

	return writer.WriteInteger(deleted)
}

// RESP: *2\r\n$6\r\nEXISTS\r\n$4\r\nkey1\r\n
// RESP: *4\r\n$6\r\nEXISTS\r\n$4\r\nkey1\r\n$4\r\nkey2\r\n$4\r\nkey3\r\n
// Pattern: EXISTS key [key ...]
// Example: EXISTS key1 → 1 (exists) or 0 (doesn't exist)
// Example: EXISTS key1 key2 nonexistent → 2 (2 out of 3 exist)
// Returns: count of how many keys exist (not which ones)
func (h *RESPHandler) handleExists(args []Value, writer *RESPWriter) error {
	if len(args) < 1 {
		return writer.WriteError("ERR wrong number of arguments for 'EXISTS' command")
	}

	count := int64(0)
	for _, arg := range args {
		if arg.Type != BulkString{
			return writer.WriteError(ERRWrongArgumentType)
		}

		exists, err := h.cachesrv.Exists(arg.Bulk)
		if err != nil {
			return writer.WriteError(formatError(err))
		}

		if exists {
			count++
		}
	}

	return writer.WriteInteger(count)
}

// RESP: *3\r\n$6\r\nEXPIRE\r\n$5\r\nmykey\r\n$2\r\n10\r\n
// Pattern: EXPIRE key seconds
// Example: EXPIRE mykey 10 → 1 (TTL set)
// Example: EXPIRE nonexistent 10 → 0 (key doesn't exist)
// Returns: 1 if TTL was set, 0 if key doesn't exist
func (h *RESPHandler) handleExpire(args []Value, writer *RESPWriter) error {
	if len(args) != 2 {
		return writer.WriteError("ERR wrong number of arguments for 'EXPIRE' command")
	}

	if args[0].Type != BulkString || args[1].Type != BulkString {
		return writer.WriteError(ERRWrongArgumentType)
	}

	key := args[0].Bulk
	seconds, err :=strconv.Atoi(args[1].Bulk)
	if err != nil {
		return writer.WriteError("ERR value is not an integer or out of range")
	}

	ttl := time.Duration(seconds) * time.Second
	err = h.cachesrv.Expire(key, ttl)
	if err != nil {
		if err == internal.ErrKeyNotFound {
			return writer.WriteInteger(0)
		}
		return writer.WriteError(formatError(err))
	}

	return writer.WriteInteger(1)
}

// RESP: *2\r\n$3\r\nTTL\r\n$5\r\nmykey\r\n
// Pattern: TTL key
// Example: TTL mykey → 10 (10 seconds remaining)
// Example: TTL nonexistent → -2 (key doesn't exist)
// Example: TTL persistkey → -1 (key exists but has no TTL)
// Returns: TTL in seconds, -2 if key doesn't exist, -1 if no TTL
func (h *RESPHandler) handleTTL(args []Value, writer *RESPWriter) error {
	if len(args) != 1 {
		return writer.WriteError("ERR wrong number of arguments for 'TTL' command")
	}

	if args[0].Type != BulkString {
		return writer.WriteError(ERRWrongArgumentType)
	}

	key := args[0].Bulk
	ttl, err := h.cachesrv.TTL(key)
	if err != nil {
		return writer.WriteError(formatError(err))
	}

	if ttl == -2 * time.Second {
		return writer.WriteInteger(-2)
	}
	if ttl == -1 * time.Second {
		return writer.WriteInteger(-1)
	}

	seconds := int64(ttl.Seconds())
	return writer.WriteInteger(seconds)
}

// RESP: *2\r\n$4\r\nKEYS\r\n$1\r\n*\r\n
// RESP: *2\r\n$4\r\nKEYS\r\n$6\r\nuser:*\r\n
// Pattern: KEYS pattern
// Example: KEYS * → all keys
// Example: KEYS user:* → ["user:1", "user:2"]
// Example: KEYS *:temp → ["cache:temp", "session:temp"]
// Returns: array of matching keys
func (h *RESPHandler) handleKeys(args []Value, writer *RESPWriter) error {
	if len(args) != 1 {
		return writer.WriteError("ERR wrong number of arguments for 'KEYS' command")
	}

	if args[0].Type != BulkString {
		return writer.WriteError(ERRWrongArgumentType)
	}

	pattern := args[0].Bulk
	keys, err := h.cachesrv.Keys(pattern)
	if err != nil {
		return writer.WriteError(formatError(err))
	}

	if err := writer.WriteArray(len(keys)); err != nil {
		return err
	}

	for _, key := range keys {
		if err := writer.WriteBulkString(key); err != nil {
			return err
		}
	}

	return nil
}

func formatError(err error) string {
	return fmt.Sprintf("ERR %s", err.Error())
}
