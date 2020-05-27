package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hekmon/transmissionrpc"
	"github.com/streadway/amqp"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

// helper function to get environment variables or return error
func getEnv(key string) (string, error) {
	if value, ok := os.LookupEnv(key); ok {
		return value, nil
	}
	return "", fmt.Errorf("%s environment variable not set", key)
}

func randomFileName(length int) string {
	rand.Seed(time.Now().UnixNano())
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789")

	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	return b.String()
}

func deleteFile(path string) error {
	// delete file
	err := os.Remove(path)
	if err != nil {
		return fmt.Errorf("Error: failed to delete %s file", path)
	}

	log.Println("File Deleted")
	return nil
}

func main() {
	// reading environment variables
	amqpAddr, err := getEnv("AMQP_ADDRESS")
	if err != nil {
		log.Fatal("error: ", err)
	}

	queueName, err := getEnv("QUEUE_NAME")
	if err != nil {
		log.Fatal("error: ", err)
	}

	transmissionHost, err := getEnv("TRANSMISSION_HOST")
	if err != nil {
		log.Fatal("error: ", err)
	}

	transmissionPortString, err := getEnv("TRANSMISSION_PORT")
	if err != nil {
		log.Fatal("error: ", err)
	}

	transmissionPort, err := strconv.Atoi(transmissionPortString)
	if err != nil {
		log.Fatal("error: TRANSMISSION_PORT not an int", err)
	}

	transmissionRPCUser, err := getEnv("TRANSMISSION_RPC_USER")
	if err != nil {
		log.Fatal("error: ", err)
	}

	transmissionRPCPassword, err := getEnv("TRANSMISSION_RPC_PASSWORD")
	if err != nil {
		log.Fatal("error: ", err)
	}

	conn, err := amqp.Dial(amqpAddr)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	failOnError(err, "Failed to set QoS")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf("Received a message")
			torType := d.Headers["x-custom-tor"].(string)
			fileName := "/" + randomFileName(8)
			err := ioutil.WriteFile(fileName, d.Body, 0644)
			failOnError(err, "Failed to write binary file")

			// maybe wrong to open connection to Transmission RPC for each file
			transmissionbt, err := transmissionrpc.New(transmissionHost, transmissionRPCUser, transmissionRPCPassword,
				&transmissionrpc.AdvancedConfig{
					Port: uint16(transmissionPort),
				})
			if err != nil {
				log.Printf("Error: not able to connect to Transmission, will try a minute later\n")
				time.Sleep(1 * time.Minute)
				continue
			}
			torrent, err := transmissionbt.TorrentAddFileDownloadDir(fileName, "/"+torType)
			if err != nil {
				log.Printf("Error: not able to push file to Transmission, will try a minute later\n")
				time.Sleep(1 * time.Minute)
				continue
			} else {
				// Only 3 fields will be returned/set in the Torrent struct
				log.Printf("Debug: %d, %s, %s\n", *torrent.ID, *torrent.Name, *torrent.HashString)
				err := deleteFile(fileName)
				if err != nil {
					log.Printf("error: ", err)
				}
			}
			log.Printf("Done")
			d.Ack(false)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
