package resp2

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

var (
	ErrInvalidProtocol = errors.New("invalid RESP2 protocol")
	ErrInvalidType     = errors.New("invalid RESP2 type")
)

const (
	SimpleString = '+'
	Error        = '-'
	Integer      = ':'
	BulkString   = '$'
	Array        = '*'
)

type Value struct {
	Type   byte
	Str    string
	Int    int64
	Bulk   string
	Array  []Value
	IsNull bool
}

type RESPParser struct {
	reader *bufio.Reader
}

func NewRESPParser(r io.Reader) *RESPParser {
	return &RESPParser{
		reader: bufio.NewReader(r),
	}
}

func (p *RESPParser) Parse() (*Value, error) {
	typeByte, err := p.reader.ReadByte()
	if err != nil {
		return nil, err
	}

	switch typeByte {
	case SimpleString:
		return p.parseSimpleString()
	case Error:
		return p.parseError()
	case Integer:
		return p.parseInt()
	case BulkString:
		return p.parseBulkString()
	case Array:
		return p.parseArray()
	default:
		return nil, fmt.Errorf("%w: unknown type %c", ErrInvalidType, typeByte)
	}
}

func (p *RESPParser) readLine() (string, error) {
	line, err := p.reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	line = strings.TrimSuffix(line, "\r\n")
	line = strings.TrimSuffix(line, "\n")
	return line, nil
}

func (p *RESPParser) parseSimpleString() (*Value, error) {
	line, err := p.readLine()
	if err != nil {
		return nil, err
	}

	return &Value{Type: SimpleString, Str: line}, nil
}

func (p *RESPParser) parseError() (*Value, error) {
	line, err := p.readLine()
	if err != nil {
		return nil, err
	}

	return &Value{Type: Error, Str: line}, nil
}

func (p *RESPParser) parseInt() (*Value, error) {
	line, err := p.readLine()
	if err != nil {
		return nil, err
	}

	num, err := strconv.ParseInt(line, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid integer", ErrInvalidProtocol)
	}

	return &Value{Type: Integer, Int: num}, nil
}

func (p *RESPParser) parseBulkString() (*Value, error) {
	line, err := p.readLine()
	if err != nil {
		return nil, err
	}

	length, err := strconv.Atoi(line)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid bulk string length", ErrInvalidProtocol)
	}

	if length == -1 {
		return &Value{Type: BulkString, IsNull: true}, nil
	}

	if length < 0 {
		return nil, fmt.Errorf("%w: negative bulk string length", ErrInvalidProtocol)
	}

	bulk := make([]byte, length)
	_, err = io.ReadFull(p.reader, bulk)
	if err != nil {
		return nil, err
	}

	_, err = p.readLine()
	if err != nil {
		return nil, err
	}

	return &Value{Type: BulkString, Bulk: string(bulk)}, nil
}

func (p *RESPParser) parseArray() (*Value, error) {
	line, err := p.readLine()
	if err != nil {
		return nil, err
	}

	count, err := strconv.Atoi(line)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid array length", ErrInvalidProtocol)
	}

	if count == -1 {
		return &Value{Type: Array, IsNull: true}, nil
	}

	if count < 0 {
		return nil, fmt.Errorf("%w: negative array length", ErrInvalidProtocol)
	}

	array := make([]Value, count)
	for i := range count {
		val, err := p.Parse()
		if err != nil {
			return nil, err
		}
		array[i] = *val
	}
	return &Value{Type: Array, Array: array}, nil
}
