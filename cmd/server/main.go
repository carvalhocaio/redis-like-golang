package server

import (
	"context"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"redis-like-golang/internal/container"
	"redis-like-golang/internal/infrastructure/persistence"
	"syscall"
	"time"
)

var (
	port      = flag.String("port", "6379", "Port to listen on")
	enableAOF = flag.Bool("aof", false, "Enable append-only file persistence")
)

func main() {
	flag.Parse()

	opt := persistence.AOFProviderOption{
		EnableAOF: *enableAOF,
		FilePath:  "data.aof",
	}
	ctn, cleanup, err := container.InitializeContainer(opt)
	if err != nil {
		log.Fatalf("Failed to create container: %v", err)
	}
	defer func() {
		if cleanup != nil {
			cleanup()
		}
		ctn.Close()
	}()

	if *enableAOF && ctn.Persistence != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := ctn.Persistence.Replay(ctx, ctn.Store); err != nil {
			log.Printf("Warning: failed to replay AOF: %v", err)
		} else {
			log.Printf("Replayed AOF")
		}
	}

	ctn.Store.StartCleanup(1000)
	defer ctn.Store.StopCleanup()

	listener, err := net.Listen("tcp", ":"+*port)
	if err != nil {
		log.Fatalf("Error listening on port %v: %v", *port, err)
	}
	defer listener.Close()

	log.Printf("Server listening on port %v (AOF: %v)", *port, *enableAOF)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Printf("Caught interrupt, shutting down...")

		time.Sleep(5 * time.Second)
		listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			break

			go ctn.TCPHandler.HandleConnection(conn)
		}
	}
}
