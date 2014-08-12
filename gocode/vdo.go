package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/codegangsta/cli"
)

type Settings struct {
	Provider Provider `json:"provider"`
	Override Override `json:"override"`
}

type Provider struct {
	ClientId   string `json:"client_id"`
	ApiKey     string `json:"api_key"`
	SshKeyName string `json:"ssh_key_name"`
	Token      string `json:"token"`
}

type Override struct {
	Ssh Ssh `json:"ssh"`
}

type Ssh struct {
	PrivateKeyPath string `json:"private_key_path"`
}

func main() {
	app := cli.NewApp()
	app.Name = "vdo"
	app.Usage = "Vagrantfile for DigitalOcean"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "input, i",
			Usage: "",
		},
		cli.StringFlag{
			Name:  "hostname",
			Usage: "",
		},
	}
	app.Action = func(c *cli.Context) {
		filePath := c.String("input")
		hostname := c.String("hostname")
		out, err := createVagrantfile(filePath, hostname)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println(string(out))
	}

	app.Commands = []cli.Command{
		{
			Name:      "up",
			ShortName: "u",
			Usage:     "Up",
			Action: func(c *cli.Context) {
			},
		},
		{
			Name:      "destroy",
			ShortName: "d",
			Usage:     "Destroy",
			Action: func(c *cli.Context) {
			},
		},
	}
	app.Run(os.Args)
}

func createVagrantfile(filePath string, hostname string) (string, error) {
	settings := new(Settings)
	if err := settings.open(filePath); err != nil {
		return "", err
	}
	if len(hostname) == 0 {
		return "", errors.New("Input hostname to create directory and use it for droplet")
	}

	provider := settings.Provider
	override := settings.Override
	dest := "./droplets/" + hostname

	// Check directory
	if err := os.Chdir(dest); err == nil {
		return "", errors.New("Directory is already exists")
	}

	// Make directory
	if err := os.Mkdir(dest, 0755); err != nil {
		return "", err
	}

	// open input file
	fi, err := os.Open("./template/Vagrantfile")
	if err != nil {
		return "", err
	}
	defer fi.Close()

	text := ""
	scanner := bufio.NewScanner(fi)
	if err := scanner.Err(); err != nil {
		return "", err
	}
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.Replace(line, "__CONFIG__VM__HOSTNAME__", hostname, 1)
		line = strings.Replace(line, "__PROVIDER__CLIENT_ID__", provider.ClientId, 1)
		line = strings.Replace(line, "__PROVIDER__API_KEY__", provider.ApiKey, 1)
		line = strings.Replace(line, "__PROVIDER__SSH_KEY_NAME__", provider.SshKeyName, 1)
		line = strings.Replace(line, "__PROVIDER__TOKEN__", provider.Token, 1)
		line = strings.Replace(line, "__OVERRIDE__SSH__PRIVATE_KEY_PATH__", override.Ssh.PrivateKeyPath, 1)
		text += line + "\n"
	}

	// open output file
	fo, err := os.Create(dest + "/Vagrantfile")
	if err != nil {
		return "", err
	}
	defer fo.Close()

	// make a write buffer
	w := bufio.NewWriter(fo)
	_, err = w.WriteString(text)
	defer w.Flush()

	result := "Create " + dest + "\n    $ cd " + dest + "\n    $ vagrant up --provider=digital_ocean --provision && vagrant ssh" + "\n    $ vagrant destroy"
	return result, nil
}

func (settingsPtr *Settings) open(filePath string) error {
	// open config file settings
	fi, err := os.Open(filePath)
	if err != nil {
		return err
	}
	// close fi on exit and check for its returned error
	defer func() {
		if err := fi.Close(); err != nil {
			panic(err)
		}
	}()

	jsonParser := json.NewDecoder(fi)
	if err = jsonParser.Decode(&settingsPtr); err != nil {
		return err
	}
	return nil
}
