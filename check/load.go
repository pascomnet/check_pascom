package check

import (
	"log"
	"strconv"
	"strings"

	"github.com/pascomnet/check_pascom/nagios"
)

func Load() nagios.CheckFunc {
	return func(n *nagios.Nagios) int {

		var usage int

		// Get Number of CPUs to calculate the load later
		o := n.ExecRemoteCommand("cat /proc/cpuinfo | grep 'model name' | wc -l")
		/*
		   Sample output o for remote command: (2 CPUS)
		   2
		*/

		if n.Debug {
			log.Print(o)
		}

		cpusStr := o[0]

		cpus, err := strconv.Atoi(cpusStr)

		if err != nil {
			log.Fatal("ERROR: check.Load: Failed to convert numbers of cpu to integer!")
		}

		// Get load from uptime. We use the 5 min average because most nagios
		// installations check every 5 min per default
		o = n.ExecRemoteCommand("uptime")
		if n.Debug {
			log.Print(o)
		}
		/*
		 Sample output o for remote command:
		 10:40  up 10 days,  2:08, 2 users, load averages: 1.44 1.67 1.73
		 or if shorter online then one day
		 16:27:18 up  7:29,  1 user,  load average: 0.11, 0.06, 0.01
		*/
		field := strings.Fields(o[0])

		load5 := field[len(field)-2]

		load5 = strings.Trim(load5, " ,")

		if n.Debug {
			log.Printf("Extracted load (5min average): %s", load5)
		}

		load5float, err := strconv.ParseFloat(load5, 64)
		if err != nil {
			log.Fatal("ERROR: check.Load: Failed to convert cpu load to float!")
		}

		load := load5float * 100
		maxLoad := float64(100 * cpus)

		usage = int(load / maxLoad * 100)

		return usage
	}
}
