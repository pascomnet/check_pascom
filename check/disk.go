package check

import (
	"log"
	"strconv"
	"strings"

	"github.com/pascomnet/check_pascom/nagios"
)

func Disk(mountPoint string) nagios.CheckFunc {
	return func(n *nagios.Nagios) int {
		o := n.ExecRemoteCommand("df --output=target,pcent | sed -e 1d | grep '^" + mountPoint + " '")

		if len(o) > 1 {
			log.Fatal("ERROR: check.Disk: Expected single line output of remote command, got multiple!")
		}

		if n.Debug {
			log.Print(o)
		}

		/*
		   Sample output o for remote command:
		   /SYSTEM            11%
		*/

		field := strings.Fields(o[0])
		usedString := field[1]

		used, err := strconv.Atoi(strings.Trim(usedString, "%"))

		if err != nil {
			log.Fatal("Failed to convert usage percent to integer: ", err)
		}
		return used
	}
}
