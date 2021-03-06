package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/kevinburke/ssh_config"
)

var (
	helpFlag = flag.Bool("help", false, "display help message of the command line")
	hostFlag = flag.String("host", "*", "Set the default host value you want to target")
	listFlag = flag.Bool("list", false, "List the configurations in .ssh/config file")
	confFlag = flag.String("config", ".ssh/config", "default config file of the SSH")
)

func main() {
	// Parse command line flags
	flag.Parse()

	// Print default and terminate program
	if *helpFlag {
		flag.PrintDefaults()
		return
	}

	// Return arguments in command line
	identity := flag.Arg(0)

	// Fetch user home directory
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	// Get current working directory
	wdir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// Appends the Home root folder to the relative route
	if strings.HasPrefix(*confFlag, "./") {
		*confFlag = path.Join(wdir, *confFlag)
	}

	// Appends the Home root folder to the relative route
	if !bytes.Contains([]byte(*confFlag), []byte(home)) {
		*confFlag = path.Join(home, *confFlag)
	}

	// Pull up file and check if there's an error
	f, err := os.Open(*confFlag)
	if err != nil {
		log.Fatal(err)
	}

	// Decode the config file to a struct
	c, err := ssh_config.Decode(f)
	if err != nil {
		log.Fatal(err)
	}

	// Display the list of configurations in the .ssh/config
	if *listFlag {
		fmt.Println(c.String())
		return
	}

	// Check if there is at least one argument in the command line
	if flag.NArg() < 1 {
		panic("no argument given to the command line")
	}

	for _, host := range c.Hosts {
		// Check if hostFlag is not `*` and that host has
		// a match for *, to skip the selection for that host
		if *hostFlag != "*" && host.Matches("*") {
			continue
		}

		// Skip everything that has not matched the `-host` flag
		if !host.Matches(*hostFlag) {
			continue
		}

		for _, node := range host.Nodes {
			switch t := node.(type) {
			case *ssh_config.KV:
				if t.Key != "IdentityFile" {
					continue
				}
				// ssh Identity Path
				identityLocation := path.Join(home, ".ssh", identity)

				// Replace the default value
				t.Value = identityLocation

				// Add a key to the ssh-agent and the keychain
				cmd := exec.Command("ssh-add", "-K", identityLocation)

				// Run the command and check for errors
				if err := cmd.Run(); err != nil {
					log.Fatal(err)
				}
			}
		}

		// Dump changes on the terminal
		fmt.Println(host.String())
	}

	// Marshal text to bytes
	mt, err := c.MarshalText()
	if err != nil {
		log.Fatal(err)
	}

	// Write the changes to file
	err = ioutil.WriteFile(*confFlag, mt, 0644)
	if err != nil {
		log.Fatal(err)
	}
}
