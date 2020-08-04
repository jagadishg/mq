package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/jagadishg/mq/mqtt"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "mq",
		Short: "mq - nice little command line tool for interacting with your MQTT Brokers",
		Long:  `Complete documentation is available at https://github.com/jagadishg/mq`,
	}
)

var (
	host       string
	username   string
	clientID   string
	cmdConnect = &cobra.Command{
		Use:   "connect",
		Short: "Connects to a mqtt broker",
		Long:  `Connects to a mqtt broker, accepts broker host url, username and clientId arguments`,
		Run: func(cmd *cobra.Command, args []string) {

			var password string
			if username != "" {
				prompt := promptui.Prompt{
					Label: "Password",
					Mask:  0x02,
				}

				var err error
				password, err = prompt.Run()
				if err != nil {
					log.Fatalln(err)
				}
			}

			connectParams := mqtt.ConnectParams{
				HostURL:  host,
				Username: username,
				ClientID: clientID,
				Password: password,
			}
			mqtt.Connect(connectParams)
		},
	}
)

var (
	pubTopic   string
	cmdPublish = &cobra.Command{
		Use:   "pub",
		Short: "Publishes a message to mqtt topic",
		Long:  `Publishes a message to mqtt topic`,
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			publishData := mqtt.PublishData{
				Topic:   pubTopic,
				Payload: args[0],
			}
			mqtt.Publish(publishData)
		},
	}
)

var (
	cmdSubscribe = &cobra.Command{
		Use:   "sub",
		Short: "Subscribes to one or more mqtt topics",
		Long:  `Messages received in the subscribed topics will be displayed on the stdout`,
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			subscribeData := mqtt.SubscribeData{
				Topics: args,
			}
			mqtt.Subscribe(subscribeData)
		},
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}

func er(msg interface{}) {
	fmt.Println("Error:", msg)
	os.Exit(1)
}

func init() {
	cobra.OnInitialize(initConfig)

	cmdConnect.Flags().BoolP("help", "", false, "Shows help")
	cmdConnect.Flags().StringVarP(&host, "host", "h", "", "MQTT Broker URL. Examples: wss://broker.mqttdashboard.com:8000, tcp://localhost:1883")
	cmdConnect.Flags().StringVarP(&username, "username", "u", "", "")
	cmdConnect.Flags().StringVarP(&clientID, "client", "c", "", "Uniquely generated clientId (Optional)")
	cmdConnect.MarkFlagRequired("host")

	cmdPublish.Flags().BoolP("help", "", false, "Shows help")
	cmdPublish.Flags().StringVarP(&pubTopic, "topic", "t", "", "Topic to publish the message")
	cmdPublish.MarkFlagRequired("topic")

	cmdSubscribe.Flags().BoolP("help", "", false, "Shows help")

	rootCmd.AddCommand(cmdConnect)
	rootCmd.AddCommand(cmdPublish)
	rootCmd.AddCommand(cmdSubscribe)
}

func initConfig() {
}
