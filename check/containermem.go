package check

import (
	"log"
	"strconv"
	"strings"

	"github.com/pascomnet/check_pascom/nagios"
)

func ContainerMem(container string) nagios.CheckFunc {
	return func(n *nagios.Nagios) int {
		o := n.ExecRemoteCommand("lxc-attach -n " + container + " -- free | sed -e 1d | head -n1")
		var usage int

		/*
		 Sample output o for remote command: (total, used, free, shared, buffers, cached)
		 Mem:        262144      54224     207920      91904          0      20660
		*/

		if n.Debug {
			log.Print(o)
		}

		field := strings.Fields(o[0])
		total, err := strconv.Atoi(field[1])
		if err != nil {
			log.Fatal("ERROR: check.Memory: Failed to convert total memory to integer!")
		}
		avail, err := strconv.Atoi(field[3])
		if err != nil {
			log.Fatal("ERROR: check.Memory: Failed to convert free memory to integer!")
		}
		usage = int((1 - float64(avail)/float64(total)) * 100)

		return usage
	}
}
