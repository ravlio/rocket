package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/rs/zerolog/log"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"rockets/internal/http"
	"rockets/internal/rocket"
)

// run initializes the HTTP server and starts listening for requests.
func run() error {
	portPtr := flag.Int("port", 8088, "HTTP Server Port")
	flag.Parse()

	ctx := context.Background()
	echo := http.NewEcho()
	logger, err := zap.NewProduction()
	if err != nil {
		return err
	}

	// Initialize the Rocket service with an in-memory store
	var rocketSvc rocket.Service
	{
		var store = rocket.NewInMemoryRocketStore(logger)
		rocketSvc = rocket.NewRocketService(store, logger)
	}

	opts := http.ServerOpts{
		Echo:   echo,
		Rocket: rocketSvc,
	}
	_, e := http.NewServer(&opts)
	g, ctx := errgroup.WithContext(ctx)

	// Start the HTTP server
	g.Go(http.ListenEchoServer(ctx, echo, fmt.Sprintf(":%d", *portPtr)))
	g.Go(http.ShutDownEchoServer(ctx, e))
	err = g.Wait()
	if err != nil {
		return err
	}
	log.Info().Msg("Service is down gracefully")
	return nil
}

func main() {
	err := run()
	if err != nil {
		log.Err(err).Msg("error")
	}
}
