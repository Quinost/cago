package simple

import (
	"cago/internal"
	"context"
	"fmt"
	"net"
	"strconv"
	"sync"
)

type SimpleServer struct {
	cfg      *internal.Config
	cachesrv *internal.CacheService
	wg       sync.WaitGroup
	ctx      context.Context
}

func NewSimpleProtocolServer(cfg *internal.Config, cacheService *internal.CacheService, ctx context.Context) *SimpleServer {
	return &SimpleServer{
		cfg:      cfg,
		cachesrv: cacheService,
		ctx:      ctx,
	}
}

func (s *SimpleServer) Run() error {
	addr := net.JoinHostPort(s.cfg.Host, strconv.Itoa(s.cfg.Port))
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	defer listener.Close()

	fmt.Printf("Cago cache server listening on %s\n", addr)
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
				fmt.Printf("Connection with client error: %v\n", err)
				continue
			}
		}

		s.wg.Add(1)
		go func(client net.Conn) {
			defer conn.Close()
			defer s.wg.Done()
			//handleConnection(client)
		}(conn)
	}
}

func (s *SimpleServer) Shutdown() {
	s.wg.Wait()
	fmt.Println("Server shutdown complete")
}
