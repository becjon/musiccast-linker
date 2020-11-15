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
	standby     = flag.Bool("standby", false, "set this to power off clients and master")
)

func mainErr() error {
	logger := log.New()
	flag.Parse()

	musiccastClient := musiccastClient.New(logger)

	clientHostnames := strings.Split(strings.ReplaceAll(*clients, "http://", ""), ",")
	masterHostname := strings.ReplaceAll(*master, "http://", "")

	if *standby {
		err := musiccastClient.PowerOff(masterHostname, *masterZone)
		if err != nil {
			return err
		}
		for _, clientHostname := range clientHostnames {
			err = musiccastClient.PowerOff(clientHostname, "main")
			return err
		}
		return nil
	}

	err := musiccastClient.PowerOn(masterHostname, *masterZone)
	if err != nil {
		return err
	}
	err = musiccastClient.Link(masterHostname, clientHostnames)
	if err != nil {
		return err
	}

	if len(*masterInput) > 0 {
		err = musiccastClient.ChangeInput(masterHostname, *masterZone, *masterInput)
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
