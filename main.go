package main

import (
    "net/http"
    "os"
    "os/signal"
    "syscall"

    "proxy-server/pkg/config"
    logger "proxy-server/pkg/log" // Alias to avoid redeclaration with standard log package
    "proxy-server/pkg/proxy"

    "github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
    // Load configuration
    cfg, err := config.LoadConfig()
    if err != nil {
        logger.Sugar.Fatalf("Error loading config: %v", err)
    }

    // Setup metrics
    http.Handle("/metrics", promhttp.Handler())
    go func() {
        logger.Sugar.Infof("Starting metric server on %s", cfg.MetricServerAddress)
        if err := http.ListenAndServe(":8081", nil); err != nil {
            logger.Sugar.Fatalf("Failed to start metric sever: %v", err)
        }
    }()

    // Initialize and start the proxy server
    proxyServer := proxy.NewProxyServer(cfg, logger.Sugar)
    logger.Sugar.Infof("Starting proxy server on %s", cfg.ServerAddress)

    // Channel to handle OS signals for graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

    go func() {
        if err := proxyServer.Start(); err != nil {
            logger.Sugar.Fatalf("Failed to start proxy server: %v", err)
        }
    }()

    // Wait for interrupt signal to gracefully shut down the server
    <-quit
    logger.Sugar.Info("Shutting down proxy server...")
    if err := proxyServer.Shutdown(); err != nil {
        logger.Sugar.Fatalf("Failed to gracefully shut down proxy server: %v", err)
    }
    logger.Sugar.Info("Proxy server stopped")
}
