package main

import (
	"flag"
	log "github.com/sirupsen/logrus"
	"musiccast-linker/musiccastClient"
	"os"
	"strings"
)

var (
	master      = flag.String("master", "", "master hostname")
	masterZone  = flag.String("master-zone", "zone2", "master zone to link")
	masterInput = flag.String("master-input", "", "(optional) set streaming input for given zone")
	clients     = flag.String("clients", "", "comma separated list of client hostnames")
)

func mainErr() error {
	logger := log.New()
	flag.Parse()

	client := musiccastClient.New(logger)

	clientHostnames := strings.Split(strings.ReplaceAll(*clients, "http://", ""), ",")
	masterHostname := strings.ReplaceAll(*master, "http://", "")

	err := client.PowerOn(masterHostname, *masterZone)
	if err != nil {
		return err
	}
	err = client.Link(masterHostname, clientHostnames)
	if err != nil {
		return err
	}

	if len(*masterInput) > 0 {
		err = client.ChangeInput(masterHostname, *masterZone, *masterInput)
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	if err := mainErr(); err != nil {
		log.Fatalf("%v", err)
		os.Exit(1)
	}
}
