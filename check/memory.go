package check

import (
	"log"
	"strconv"
	"strings"

	"github.com/pascomnet/check_pascom/nagios"
)

func Memory() nagios.CheckFunc {
	return func(n *nagios.Nagios) int {
		o := n.ExecRemoteCommand("free | sed -e 1d | head -n1")
		var usage int

		/*
		 Sample output o for remote command: (total,used,free,shared,buff/cache,available)
		 Mem:        8175352      732812     4883020      225584     2559520     6732128
		*/

		if n.Debug {
			log.Print(o)
		}

		field := strings.Fields(o[0])
		total, err := strconv.Atoi(field[1])
		if err != nil {
			log.Fatal("ERROR: check.Memory: Failed to convert total memory to integer!")
		}
		avail, err := strconv.Atoi(field[6])
		if err != nil {
			log.Fatal("ERROR: check.Memory: Failed to convert available memory to integer!")
		}
		usage = int((1 - float64(avail)/float64(total)) * 100)

		return usage
	}
}
