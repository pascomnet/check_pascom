package check

import (
	"log"
	"strings"

	"github.com/pascomnet/check_pascom/nagios"
)

func ContainerState(container string) nagios.CheckFunc {
	return func(n *nagios.Nagios) int {

		var state int

		o := n.ExecRemoteCommand("lxc-ls -f --filter=^" + container + "$ | tail -n1")

		if o == nil {
			return -1
		}

		/*
		   Sample output o for remote command: (NAME,STATE,AUTOSTART,GROUPS,IPV4,IPV6)
		   controller RUNNING 0         cs     10.0.3.182 -
		*/

		if n.Debug {
			log.Print(o)
		}

		field := strings.Fields(o[0])
		stateStr := field[1]

		if stateStr == "RUNNING" {
			state = 1
		} else {
			state = 0
		}

		return state

	}
}
