package main

import (
	"fmt"
	"log"

	"github.com/go-daq/canbus"
)

func main() {
	recv, err := canbus.New()
	if err != nil {
		log.Fatal(err)
	}
	defer recv.Close()

	send, err := canbus.New()
	if err != nil {
		log.Fatal(err)
	}
	defer send.Close()

	err = recv.Bind("vcan0")
	if err != nil {
		log.Fatalf("could not bind recv socket: %+v", err)
	}

	err = send.Bind("vcan0")
	if err != nil {
		log.Fatalf("could not bind send socket: %+v", err)
	}

	for i := 0; i < 5; i++ {
		_, err := send.Send(canbus.Frame{
			ID:   0x123,
			Data: []byte(fmt.Sprintf("data-%02d", i)),
			Kind: canbus.SFF,
		})
		if err != nil {
			log.Fatalf("could not send frame %d: %+v", i, err)
		}
	}

	for i := 0; i < 5; i++ {
		frame, err := recv.Recv()
		if err != nil {
			log.Fatalf("could not recv frame %d: %+v", i, err)
		}
		fmt.Printf("frame-%02d: %q (id=0x%x)\n", i, frame.Data, frame.ID)
	}

}