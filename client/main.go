package main

import (
	"fmt"
	"github.com/nats-io/stan.go"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	stanConn, err := stan.Connect(
		"kso_cluster", "stan_client", stan.Pings(60, 2*60),
		stan.SetConnectionLostHandler(func(_ stan.Conn, reason error) {
			log.Fatal("Connection lost, reason: " + reason.Error())
		}),
		stan.NatsURL("localhost:4222"),
	)
	if err != nil {
		log.Fatal("on stan connect", err.Error())
	}
	defer stanConn.Close()

	go func() {
		stanConn.Subscribe("measurement", func(msg *stan.Msg) {
			fmt.Println(msg)
		})
	}()

	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	<-signals
}
