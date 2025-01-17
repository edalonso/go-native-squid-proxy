package handler

import (
    "io"
    "net"
    "log"
    "strings"

    logger "proxy-server/pkg/log" // Alias to avoid redeclaration with standard log package
    "github.com/valyala/fasthttp"
    "proxy-server/pkg/metrics"
    "proxy-server/pkg/pool"
)

// HandleRequest handles both normal HTTP requests and HTTP CONNECT requests for HTTPS
func HandleRequest(ctx *fasthttp.RequestCtx) {
    metrics.IncrementRequestCounter()
    // Setup logging
    loggerInstance, err := logger.NewLogger("info")
    if err != nil {
        log.Fatalf("Error setting up logger: %v", err)
    }
    defer loggerInstance.Sync()
    sugar := loggerInstance.Sugar()
    sugar.Infof("Requested URI: %s. User Agent header: %s. Method: %s. From IP: %s. Post Arguments: %s. Host: %s", string(ctx.RequestURI()), string(ctx.Request.Header.Peek("User-Agent")), string(ctx.Method()), ctx.RemoteAddr().String(), ctx.PostArgs().String(), string(ctx.Host()))

    sugar.Info("Printing all request headers:")
    ctx.Request.Header.VisitAll(func (key, value []byte) {
        sugar.Infof("%v: %v", string(key), string(value))
    })

    // Handle HTTP CONNECT method for HTTPS proxying
    if string(ctx.Method()) == fasthttp.MethodConnect {
        sugar.Info("Handle tunneling")
        handleTunneling(ctx)
    } else {
        sugar.Info("No handle tunneling")
        handleHTTP(ctx)
    }
}

// handleHTTP handles standard HTTP requests
func handleHTTP(ctx *fasthttp.RequestCtx) {
    client := pool.GetConnection()
    defer pool.PutConnection(client)

    req := &ctx.Request
    resp := &ctx.Response

    if err := client.Do(req, resp); err != nil {
        ctx.Error("Failed to process request: "+err.Error(), fasthttp.StatusInternalServerError)
        return
    }
}

// handleTunneling handles HTTP CONNECT requests
func handleTunneling(ctx *fasthttp.RequestCtx) {
    // Setup logging
    loggerInstance, err := logger.NewLogger("info")
    if err != nil {
        log.Fatalf("Error setting up logger: %v", err)
    }
    defer loggerInstance.Sync()
    sugar := loggerInstance.Sugar()
    dest := ""
    str_host:=string(ctx.Host())
    if strings.Contains(":", str_host) {
        dest = str_host
    } else {
        dest = str_host + ":443"
    }
    destinationConn, err := net.Dial("tcp", dest)
    if err != nil {
        sugar.Errorf("Failed to connect to destination: %s", err.Error())
        ctx.Error("Failed to connect to destination", fasthttp.StatusServiceUnavailable)
        return
    }

    // Send 200 OK response to the client
    ctx.SetStatusCode(fasthttp.StatusOK)
    ctx.Response.SetBodyRaw([]byte{})

    ctx.Hijack(func(clientConn net.Conn) {
        defer clientConn.Close()
        defer destinationConn.Close()

        go func() {
            _, _ = io.Copy(destinationConn, clientConn)
        }()
        _, _ = io.Copy(clientConn, destinationConn)
    })
}

// transfer copies data between source and destination connections
func transfer(destination net.Conn, source net.Conn) {
    _, _ = io.Copy(destination, source)
}
