package nagios

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/mitchellh/go-homedir"
	"golang.org/x/crypto/ssh"
)

const (
	CRITICAL = "CRITICAL"
	WARNING  = "WARNING"
	OK       = "OK"
	UNKNOWN  = "UNKNOWN"
)

var exitCodes = map[string]int{
	OK:       0,
	WARNING:  1,
	CRITICAL: 2,
	UNKNOWN:  3,
}

type Nagios struct {
	User       string
	Host       string
	Checks     []Check
	Containers []Container
	Excludes   map[string]bool
	Debug      bool
	client     *ssh.Client
}

type Check struct {
	Name          string
	Check         CheckFunc
	WarnThreshold string
	CritThreshold string
	State         string
}

type Container struct {
	URL          string      `json:"url"`
	ID           int         `json:"id"`
	Name         string      `json:"name"`
	DisplayName  interface{} `json:"display_name"`
	Memory       int         `json:"memory"`
	Running      bool        `json:"running"`
	ImageName    string      `json:"image_name"`
	ImageVersion string      `json:"image_version"`
	Host         int         `json:"host"`
}

type CheckFunc func(*Nagios) int

func New(user string, host string, debug bool) (n *Nagios) {
	n = &Nagios{
		User:  user,
		Host:  host,
		Debug: debug,
	}

	n.Excludes = make(map[string]bool)

	return

}

func (n *Nagios) AddCheck(name string, check CheckFunc, warn string, crit string) {
	n.Checks = append(n.Checks, Check{
		Name:          name,
		Check:         check,
		WarnThreshold: warn,
		CritThreshold: crit,
	})

}

func (n *Nagios) AddExclude(name string) {
	n.Excludes[name] = true
}

func (n *Nagios) GetContainerInfo() {

	var containerListByte []byte

	containerListJSON := n.ExecRemoteCommand("lxc-attach -n controller -- wget --no-check-certificate -O- -q https://localhost/api/v1/containers/")

	for _, line := range containerListJSON {
		containerListByte = append(containerListByte, []byte(line)...)
	}

	err := json.Unmarshal(containerListByte, &n.Containers)

	if err != nil {
		log.Fatal("Cloud not convert json container list to struct", err)
	}

}

func (n *Nagios) DoChecks() {
	for index, check := range n.Checks {
		if n.Debug {
			log.Printf("=== Performing check '%s' (warn: %s , crit: %s)", check.Name, check.WarnThreshold, check.CritThreshold)
		}
		result := check.Check(n)

		state := OK

		if n.testThreshold(result, check.WarnThreshold) {
			state = WARNING
		}

		if n.testThreshold(result, check.CritThreshold) {
			state = CRITICAL
		}

		if result == -1 {
			state = UNKNOWN
		}

		n.Checks[index].State = state
		if n.Debug {
			log.Printf("State: %s, Result: %d", state, result)
		}
	}
}

func (n *Nagios) Connect() {

	if runtime.GOOS == "linux" {
		os.Setenv("HOME", "")
	}

	home, err := homedir.Dir()
	if err != nil {
		log.Fatalf("Unable to read homedir of user: %v", err)
	}

	key, err := ioutil.ReadFile(home + "/.ssh/id_rsa")
	if err != nil {
		log.Fatalf("unable to read private key: %v", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatalf("unable to parse private key: %v", err)
	}

	config := &ssh.ClientConfig{
		User: n.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", n.Host+":22", config)
	n.client = client
	if err != nil {
		log.Fatalf("SSH Connection failed. User: '%s', Host: '%s', Auth: sshkey, Error: '%s'", n.User, n.Host, err.Error())
	}

}

func (n *Nagios) ExecRemoteCommand(command string) []string {
	var output []string
	session, err := n.client.NewSession()
	if err != nil {
		log.Fatal("Failed to create session: ", err)
	}
	defer session.Close()

	var b bytes.Buffer
	session.Stdout = &b
	if err := session.Run(command); err != nil {
		log.Fatal("Failed to run: " + err.Error())
	}

	scanner := bufio.NewScanner(&b)

	for scanner.Scan() {
		output = append(output, scanner.Text())
	}

	return output

}

func (n *Nagios) Exit() {

	var output string
	var multiline string

	overallState := UNKNOWN

	for _, check := range n.Checks {

		multiline += check.Name + " is " + check.State + "\n"

		switch check.State {
		case CRITICAL:
			overallState = CRITICAL

		case WARNING:
			if overallState != CRITICAL {
				overallState = WARNING
			}

		case OK:
			if overallState != CRITICAL && overallState != WARNING {
				overallState = OK
			}
		}
	}

	if overallState == OK {
		output = overallState + ": All checks are ok"
	} else {
		output = overallState + ": At least one check is in " + strings.ToLower(overallState) + " state"
	}

	output += "\n" + multiline

	fmt.Print(output)
	os.Exit(exitCodes[overallState])

}

func (n *Nagios) testThreshold(value int, threshold string) bool {

	const MaxUint = ^uint(0)
	const MinUint = 0
	const MaxInt = int(MaxUint >> 1)
	const MinInt = -MaxInt - 1

	rangeStart := 0
	rangeEnd := 0
	insideRange := false
	match := false

	if strings.HasPrefix(threshold, "@") {
		insideRange = true
		threshold = strings.TrimLeft(threshold, "@")
	}

	if strings.Contains(threshold, ":") {
		fields := strings.Split(threshold, ":")
		start := fields[0]
		end := fields[1]

		if end == "" {
			end = strconv.Itoa(MaxInt)
		}
		if start == "~" {
			start = strconv.Itoa(MinInt)
		}

		startInt, err := strconv.Atoi(start)
		if err != nil {
			log.Fatalf("Threshold '%s': '%s' is not a valid number!", threshold, start)
		}
		endInt, err := strconv.Atoi(end)
		if err != nil {
			log.Fatalf("Threshold '%s': '%s' is not a valid number!", threshold, end)
		}

		rangeStart = startInt
		rangeEnd = endInt

	} else {
		end := threshold
		endInt, err := strconv.Atoi(end)
		if err != nil {
			log.Fatalf("Threshold '%s': '%s' is not a valid number!", threshold, end)
		}
		rangeEnd = endInt
	}

	if rangeStart > rangeEnd {
		log.Fatalf("Threshold '%s' is not valid because %d is grater then %d!", threshold, rangeStart, rangeEnd)
	}

	if insideRange {
		if rangeStart <= value && value <= rangeEnd {
			match = true
		}
	} else {
		if value < rangeStart || rangeEnd < value {
			match = true
		}
	}

	if n.Debug {
		log.Printf("Threshold '%d:%d', Value '%d', InsideRange '%t', Match '%t'", rangeStart, rangeEnd, value, insideRange, match)
	}

	return match

}
