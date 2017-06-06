package main

import (
	"flag"
	"log"
	"time"

	"github.com/pascomnet/check_pascom/check"
	"github.com/pascomnet/check_pascom/nagios"
)

func main() {

	debug := flag.Bool("debug", false, "Enable debug")
	user := flag.String("user", "root", "SSH user of monitored pascom system")
	host := flag.String("host", "", "Hostname of the pascom system you want to monitor")
	timeout := flag.Int("timeout", 15, "Overall check_pascom timeout in seconds")

	flag.Parse()

	if *host == "" {
		log.Fatal("Please specify a host with the -host option. Try -h for help.")
	}

	timeoutChan := time.NewTimer(time.Second * time.Duration(*timeout)).C

	go func() {

		n := nagios.New(*user, *host, *debug)

		n.Connect()

		n.AddCheck("Memory", check.Memory(), "80", "90")
		n.AddCheck("Swap", check.Swap(), "80", "90")
		n.AddCheck("Ramdisk /", check.Disk("/"), "75", "90")
		n.AddCheck("Disk /SYSTEM", check.Disk("/SYSTEM"), "75", "90")
		n.AddCheck("Load", check.Load(), "80", "90")
		n.AddCheck("Container pg", check.ContainerState("pg"), "0:2", "0:2")
		n.AddCheck("Container controller", check.ContainerState("controller"), "0:2", "0:2")

		containerList := n.ExecRemoteCommand("lxc-ls --running -1")

		for _, container := range containerList {

			n.AddCheck("Container mem "+container, check.ContainerMem(container), "90", "95")

		}

		n.DoChecks()
		n.Exit()

	}()

	for {
		select {
		case <-timeoutChan:
			log.Fatalf("check_pascom timeout after %d seconds. See -h for help.", *timeout)
		}
	}

}
