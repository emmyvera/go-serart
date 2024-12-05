package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
)

const (
	WEB_PORT = "81"
	rpcPort  = "5001"
)

type Config struct {
}

func main() {
	app := Config{}
	// Resgister the RPC server
	err := rpc.Register(new(RPCServer))
	go app.rpcListen()
	if err != nil {
		log.Panic(err)
		return
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", WEB_PORT),
		Handler: app.routes(),
	}

	err = srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}

func (app *Config) rpcListen() error {
	log.Println("Starting RPC server on port ", rpcPort)
	listen, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", rpcPort))
	if err != nil {
		return err
	}
	defer listen.Close()

	for {
		rpcConn, err := listen.Accept()
		if err != nil {
			continue
		}
		go rpc.ServeConn(rpcConn)
		//go app.gRCListen()
	}

}
