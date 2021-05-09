package main

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v3"
)

var conf config

// struct to map config.yaml content
type config struct {
	AmqpAddr                string `yaml:"amqpAddr"`
	QueueName               string `yaml:"queueName"`
	TransmissionHost        string `yaml:"transmissionHost"`
	TransmissionPort        uint16 `yaml:"transmissionPort"`
	TransmissionRPCUser     string `yaml:"transmissionRPCUser"`
	TransmissionRPCPassword string `yaml:"transmissionRPCPassword"`
}

// Config - function to read the config.yaml
func Config() {
	yamlFile, err := ioutil.ReadFile("/vault/secrets/config.yaml")
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, &conf)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
}
