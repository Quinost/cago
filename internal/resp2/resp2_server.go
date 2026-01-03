package resp2

import (
	"cago/internal"
	"context"
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"
)

type RESPServer struct {
	cfg     *internal.Config
	handler *RESPHandler
	wg      sync.WaitGroup
	ctx     context.Context
}

func NewRESP2Server(cfg *internal.Config, cacheSrv *internal.CacheService, ctx context.Context) *RESPServer {
	return &RESPServer{
		cfg:     cfg,
		handler: NewRESPHandler(cacheSrv),
		ctx:     ctx,
	}
}

func (s *RESPServer) Run() error {
	addr := net.JoinHostPort(s.cfg.Host, strconv.Itoa(s.cfg.Port))
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	defer listener.Close()

	fmt.Printf("RESP2 server listening on %s\n", addr)
	fmt.Printf("Default TTL: %v, Cleanup interval: %v\n", s.cfg.DefaultTTL, s.cfg.CleanupInterval)

	go func() {
		<-s.ctx.Done()
		listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-s.ctx.Done():
				return nil
			default:
				fmt.Printf("Connection error: %v\n", err)
				continue
			}
		}

		s.wg.Add(1)
		go s.handleConnection(conn)
	}
}

func (s *RESPServer) handleConnection(conn net.Conn) {
	defer conn.Close()
	defer s.wg.Done()

	fmt.Printf("Client connected: %s\n", conn.RemoteAddr())

	parser := NewRESPParser(conn)
	writer := NewRESPWriter(conn)

	for {
		select {
		case <-s.ctx.Done():
			return
		default:
		}

		cmd, err := parser.Parse()
		if err != nil {
			if err == io.EOF {
				fmt.Printf("Client diconnected: %s\n", conn.RemoteAddr())
				return
			}

			fmt.Printf("Parse error: %v\n", err)
			writer.WriteError(fmt.Sprintf("ERR protocol error: %v", err))
			return
		}

		if err := s.handler.HandleCommand(cmd, writer); err != nil {
			fmt.Printf("Handler error: %v\n", err)
			return

		}
	}
}

func (s *RESPServer) Shutdown() {
	s.wg.Wait()
	fmt.Println("RESP server shutdown complete")
}
