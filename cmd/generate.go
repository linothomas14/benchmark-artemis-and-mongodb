/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"benchmark-mongo-and-artemis/util"
	"context"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/go-stomp/stomp"
	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate data to Artemis, with atribute on config.yaml",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		DB := cmd.Flag("DB").Value.String()
		generate(DB)
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.Flags().String("DB", "artemis", "Type of database to use")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// generateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// generateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func generate(DB string) {
	// Load configuration
	err := util.LoadConfig(".")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	switch DB {
	case "mongo":
		generateToMongo()

	case "artemis":
		generateToArtemis()

	}

	fmt.Println("Finished sending messages")
}

func generateToArtemis() {
	hostString := fmt.Sprintf("%s:%d", util.Configuration.Artemis.Host, util.Configuration.Artemis.Port)

	// Connection parameters
	conn, err := stomp.Dial("tcp", hostString, stomp.ConnOpt.Login(util.Configuration.Artemis.User, util.Configuration.Artemis.Password))
	if err != nil {
		log.Fatalf("Failed to connect to ActiveMQ: %v", err)
	}
	defer conn.Disconnect()

	// Destination queue
	queueName := util.Configuration.Artemis.QueueName

	logFile, err := openLogFile("log/log-generate.txt")

	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	// Send messages
	for i := 1; i <= util.Configuration.Artemis.NumberOfData; i++ {
		message := fmt.Sprintf("{\"id\": %d }", i)
		err = conn.Send(queueName, "application/json", []byte(message), stomp.SendOpt.Header("persistent", "true"), stomp.SendOpt.Header("destination-type", "ANYCAST"))
		if err != nil {
			log.Printf("Failed to send message %d: %v", i, err)
		} else {

			log.Printf("Message %d sent successfully", i)
		}
		// time.Sleep(10 * time.Millisecond) // Throttle the message sending rate
	}
}

func generateToMongo() {
	// Connect to mongo

	cfg := util.Configuration.MongoDB

	optionsStr := fmt.Sprintf("&%s=%s", "authSource", url.QueryEscape(cfg.AuthSource))

	// Set the MongoDB client options
	clientOptions := options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%s@%s:%d/?%s",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, optionsStr))

	// Customize additional options if needed
	clientOptions.MaxPoolSize = &cfg.MaxPoolSize
	timeout := time.Duration(cfg.ConnectTimeout) * time.Second
	clientOptions.ConnectTimeout = &timeout

	// Connect to the MongoDB server
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, clientOptions)

	if err != nil {
		log.Printf("Failed to connect to mongo: %v", err)
		log.Fatal(err)
		return
	}

	defer client.Disconnect(context.TODO())

	var result bson.M
	err = client.Database("admin").RunCommand(context.TODO(), bson.D{{Key: "ping", Value: 1}}).Decode(&result)
	if err != nil {
		log.Printf("Failed to ping mongo: %v", err)
		return
	}

	// Send data to mongo coll
	coll := client.Database(cfg.Database).Collection(cfg.Collection)
	for i := 1; i <= cfg.NumberOfData; i++ {
		_, err := coll.InsertOne(context.TODO(), bson.M{"id": i})
		if err != nil {
			log.Printf("Failed to send message %d: %v", i, err)
		} else {
			log.Printf("Message %d sent successfully", i)
		}
	}
}
