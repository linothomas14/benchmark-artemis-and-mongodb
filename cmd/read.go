/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"benchmark-mongo-and-artemis/util"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/go-stomp/stomp"
	"github.com/spf13/cobra"
)

// readCmd represents the read command
var readCmd = &cobra.Command{
	Use:   "read",
	Short: "Read messages from the queue",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		count, _ := cmd.Flags().GetInt("count")

		stopChannel := make(chan int)
		// var wg sync.WaitGroup
		// CATAT WAKTU MULAI
		start := time.Now()

		var messageCount int

		for i := 1; i <= count; i++ {
			// wg.Add(1)

			// defer wg.Done()
			fmt.Println("Consumer ID: " + strconv.Itoa(i))
			go read(strconv.Itoa(i), stopChannel)

		}

		// go func() {
		// 	wg.Wait()
		//
		// }()

		// MENERIMA DATA DARI CHANNEL, JIKA ADA DATA, HENTIKAN PROGRAM
		for range stopChannel {
			messageCount++
			if messageCount >= util.Configuration.Artemis.NumberOfData {
				close(stopChannel)
				break
			}
		}
		fmt.Println("stop")
		end := time.Now()
		fmt.Printf("Total time taken: %v\n", end.Sub(start))

	},
}

func init() {
	rootCmd.AddCommand(readCmd)

	// Define and bind flag
	readCmd.Flags().Int("count", 1, "Number of consumer")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// readCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// readCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func read(count string, stopChannel chan int) {
	// Load configuration

	err := util.LoadConfig(".")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	hostString := fmt.Sprintf("%s:%d", util.Configuration.Artemis.Host, util.Configuration.Artemis.Port)

	// Connection parameters
	conn, err := stomp.Dial("tcp", hostString, stomp.ConnOpt.Login(util.Configuration.Artemis.User, util.Configuration.Artemis.Password))
	if err != nil {
		log.Fatalf("Failed to connect to ActiveMQ: %v", err)
	}
	defer conn.Disconnect()

	// Destination queue
	queueName := util.Configuration.Artemis.QueueName

	// Subscribe to the specified queue
	sub, err := conn.Subscribe(queueName, stomp.AckClientIndividual)
	if err != nil {
		log.Fatalf("Failed to subscribe to queue: %v", err)
	}
	defer sub.Unsubscribe()

	fmt.Printf("consumer %s Start listening to %s...\n", count, queueName)

	// Continuously receive messages and send them to the channel

	for msg := range sub.C {

		processMessages(conn, msg, stopChannel)

		if msg.Err != nil {
			log.Fatalf("Failed to receive message: %v", msg.Err)
		}

	}
}

func processMessages(conn *stomp.Conn, msg *stomp.Message, stopChannel chan int) {

	log.Printf("Received message: %s\n", string(msg.Body))

	// time.Sleep(100 * time.Millisecond)

	if err := conn.Ack(msg); err != nil {
		log.Fatalf("Failed to acknowledge message: %v", err)
	}

	stopChannel <- 1

}

func openLogFile(fileName string) (*os.File, error) {
	logFilePath := filepath.Join(".", fileName)

	// Check if the file exists
	if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
		// Create the file if it does not exist
		logFile, err := os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return nil, err
		}
		return logFile, nil
	}

	// Clear the file contents if it exists
	err := os.Truncate(logFilePath, 0)
	if err != nil {
		return nil, err
	}

	// Open the file with read/write and append mode
	logFile, err := os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	return logFile, nil
}
