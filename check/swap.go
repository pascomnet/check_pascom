package check

import (
	"log"
	"strconv"
	"strings"

	"github.com/pascomnet/check_pascom/nagios"
)

func Swap() nagios.CheckFunc {
	return func(n *nagios.Nagios) int {
		o := n.ExecRemoteCommand("free | sed -e 1d | tail -n1")
		var usage int

		/*
		 Sample output o for remote command: (total,used,free)
		 Swap:       4191228           0     4191228
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
			log.Fatal("ERROR: check.Memory: Failed to convert available memory to integer!")
		}
		usage = int((1 - float64(avail)/float64(total)) * 100)

		return usage
	}
}
