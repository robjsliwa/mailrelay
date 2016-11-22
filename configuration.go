package main

import (
	"encoding/json"
	"os"
)

// The first letter of the struct elements must be upper case in order to export them
// The JSON decoder will not use struct elements that are not exported
// This struct will be used to unmarshal the configuration file read at startup
type Configuration struct {
	Log_file       string
	Ssl_cert_file  string
	Ssl_key_file   string
	Server_port    string
	Server_host    string
	Mailrelay_fqdn string
	Max_retries    string
}

var emailConfigurationInstance *Configuration = nil

func EmailConfigurationInstance() *Configuration {
	if emailConfigurationInstance == nil {

		emailConfiguration := &Configuration{
			Log_file:       "",
			Ssl_cert_file:  "",
			Ssl_key_file:   "",
			Server_port:    "",
			Server_host:    "",
			Mailrelay_fqdn: "",
			Max_retries:    "",
		}

		emailConfigurationInstance = emailConfiguration
	}

	return emailConfigurationInstance
}

func (c *Configuration) GetConfiguration(config_file string) (err error) {
	file, err := os.Open(config_file)
	if err == nil {
		decoder := json.NewDecoder(file)
		err = decoder.Decode(c)
	}
	return
}
