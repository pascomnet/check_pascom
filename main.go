package main

import (
	"flag"
	"log"
	"strings"
	"time"

	"github.com/pascomnet/check_pascom/check"
	"github.com/pascomnet/check_pascom/nagios"
)

func main() {

	debug := flag.Bool("debug", false, "Enable debug")
	user := flag.String("user", "root", "SSH user of monitored pascom system")
	host := flag.String("host", "", "Hostname of the pascom system you want to monitor")
	mode := flag.String("mode", "system", "Check 'system', 'all_customers' or 'single_customer'")
	excludes := flag.String("exclude", "", "List of comma separated customers which should be excluded. Only in mode 'all_customers'")
	customer := flag.String("customer", "", "Name of a customer to check. Only in mode 'single_customer'")
	timeout := flag.Int("timeout", 15, "Overall check_pascom timeout in seconds")

	flag.Parse()

	if *host == "" {
		log.Fatal("Please specify a host with the -host option. Try -h for help.")
	}

	timeoutChan := time.NewTimer(time.Second * time.Duration(*timeout)).C

	go func() {

		n := nagios.New(*user, *host, *debug)

		n.Connect()

		n.GetContainerInfo()

		switch *mode {
		case "system":
			n.AddCheck("Memory", check.Memory(), "80", "90")
			n.AddCheck("Swap", check.Swap(), "80", "90")
			n.AddCheck("Ramdisk /", check.Disk("/"), "75", "90")
			n.AddCheck("Disk /SYSTEM", check.Disk("/SYSTEM"), "75", "90")
			n.AddCheck("Load", check.Load(), "80", "90")

			for _, container := range n.Containers {
				switch container.ImageName {
				case "cs-proxy":
				case "cs-controller":
				case "cs-postgresql":
				default:
					continue
				}

				n.AddCheck("Container state "+container.Name, check.ContainerState(container.Name), "0:2", "0:2")
				n.AddCheck("Container mem "+container.Name, check.ContainerMem(container.Name), "90", "95")
			}

		case "single_customer":
			if *customer == "" {
				log.Fatal("Please specify a 'customer' in mode 'single_customer'. Try -h for help.")
			}
			n.AddCheck("Container state "+*customer, check.ContainerState(*customer), "0:2", "0:2")
			n.AddCheck("Container mem "+*customer, check.ContainerMem(*customer), "90", "95")

		case "all_customers":

			if *excludes != "" {
				for _, exclude := range strings.Split(*excludes, ",") {
					n.AddExclude(exclude)
				}
			}

			for _, container := range n.Containers {
				if container.ImageName != "mobydick" {
					continue
				}
				if n.Excludes[container.Name] == true {
					continue
				}
				if container.Running == false {
					continue
				}

				n.AddCheck("Container state "+container.Name, check.ContainerState(container.Name), "0:2", "0:2")
				n.AddCheck("Container mem "+container.Name, check.ContainerMem(container.Name), "90", "95")
			}

		default:
			log.Fatal("Please specify a valid mode with the -mode option. Try -h for help.")

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
