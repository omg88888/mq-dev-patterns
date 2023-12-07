/****************************************************************************************/
/*                                                                                      */
/*                                                                                      */
/*  Copyright 2023 IBM Corp.                                                            */
/*                                                                                      */
/*  Licensed under the Apache License, Version 2.0 (the "License");                     */
/*  you may not use this file except in compliance with the License.                    */
/*  You may obtain a copy of the License at                                             */
/*                                                                                      */
/*  http://www.apache.org/licenses/LICENSE-2.0                                          */
/*                                                                                      */
/*  Unless required by applicable law or agreed to in writing, software                 */
/*  distributed under the License is distributed on an "AS IS" BASIS,                   */
/*  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.            */
/*  See the License for the specific language governing permissions and                 */
/*  limitations under the License.                                                      */
/*                                                                                      */
/*                                                                                      */
/****************************************************************************************/
/*                                                                                      */
/*  FILE NAME:      amqpPut.go                                                          */
/*                                                                                      */
/*  DESCRIPTION:    A Go applicaton used to send 10 messages to broker-IBM MQ.          */
/*  AMQP 1.0 in use. Go-azure library is used.                                          */
/*  Refer : https://github.com/Azure/go-amqp                                            */
/*  Documentation: https://pkg.go.dev/github.com/Azure/go-amqp#section-documentation    */
/*                                                                                      */
/*  How to Run:                                                                         */
/*  Usage:  go run amqpPut.go [-t TopicName] [-q QueueName]                             */
/*                                                                                      */
/*  Ex:  go run amqpPut.go -q LOCAL.QUEUE.1                                             */
/*                                                                                      */
/*  Parameters:                                                                         */
/*  -t ==> TopicName                                                                    */
/*  -q ==> QueueName                                                                    */
/****************************************************************************************/

package main

import (
    "log"
    "fmt"
    "context"
    "flag"
    "encoding/json"
    "io/ioutil"
    "os"

    "github.com/Azure/go-amqp"
)

type Env struct {
    Username         string `json:"APP_USER"`
    Password         string `json:"APP_PASSWORD"`
    Host             string `json:"HOST"`
    Port             string `json:"PORT"`
}

type MQEndpoints struct {
    Points []Env `json:"MQ_ENDPOINTS"`
}

var (
    queueName   string
    topicName   string
    valid       bool
    destination string
    isQueue     bool
)
var EnvSettings Env
var MQ_ENDPOINTS MQEndpoints

func init() {
    //Parsing the input
    flag.StringVar(&queueName, "q", "DEMO", "Queue")
    flag.StringVar(&topicName, "t", "TOP", "Topic")
    flag.Parse()
    if flag.NFlag() > 2 {
        valid = false
    } else if flag.NFlag() == 1 {
        flag.Visit(func(f *flag.Flag) {
            if f.Name == "q" || f.Name == "t" {
                valid = true
            }
            if f.Name == "q"{
                destination = queueName
                isQueue = true
            }else{
                destination = topicName
            }
        })
    } else {
        valid = false
    }

    if valid{
        // ***********Configuring MQ Credentials***********
        // Read MQ configuration from env.json
        //Check for file path in env variables else default to ../env.json
        DEFAULT_ENV_FILE := "../../env.json"
        filePath := os.Getenv("ENV_FILE")
        if filePath != ""{
            fmt.Println("ENV file is set to " + filePath)
        } else{
            fmt.Println("Using default ENV file: "+DEFAULT_ENV_FILE)
            filePath = DEFAULT_ENV_FILE
        }
        content, err := ioutil.ReadFile(filePath)
        if err != nil {
            log.Fatal("Error when opening file: ", err)
        }
        // Unmarshall the data
        err = json.Unmarshal(content, &MQ_ENDPOINTS)
        if err != nil {
            log.Fatal("Error during Unmarshal(): ", err)
        }
        if len(MQ_ENDPOINTS.Points) > 0 {
            EnvSettings = MQ_ENDPOINTS.Points[0]
        }
    }
}

func main() {
    if !valid || queueName=="" || topicName==""{
        fmt.Println("Invalid arguments")
        return
    }

    //Defining the AMQP URI
    uri:= "amqp://"+EnvSettings.Username+":"+EnvSettings.Password+"@"+EnvSettings.Host+":"+EnvSettings.Port

    // Set up connection parameters
    conn, err := amqp.Dial(context.TODO(), uri, nil)
    fmt.Println("Connected to host.")	
    if err != nil {
        fmt.Println("Error in creating connection.")
        log.Fatal(err)
    }

    // Open a session
    session, err := conn.NewSession(context.TODO(), nil)
    fmt.Println("Session created.")
    if err != nil {
        fmt.Println("Error in creating Session")
        log.Fatal(err)
    }

    //When the destination is a queue, capability must be set to "queue". If not, topic is selected by default.
    SenderOptions := &amqp.SenderOptions{
        TargetCapabilities: []string{"topic"},
    }
    if isQueue{
        SenderOptions = &amqp.SenderOptions{
            TargetCapabilities: []string{"queue"},
        }
    }

    // Create a sender
    sender, err := session.NewSender(context.TODO(), destination, SenderOptions)
    fmt.Println("Sender mapped to destination.")
    if err != nil {
        fmt.Println("Error in creating Sender")
        log.Fatal(err)
    }
    fmt.Println("Destination :",SenderOptions.TargetCapabilities)

    //Message.Properties.To must contain the destination address (URI)
    addr := uri
    // Create a message
    msg := &amqp.Message{
        Data: [][]byte{[]byte("Hello, World!")},
        Properties: &amqp.MessageProperties{
            To: &addr,
        },
    }

    // Send the messages
    for i:=0;i<10;i++{
        err = sender.Send(context.TODO(), msg, nil)
        if err != nil {
            fmt.Println("Error in MessageSend")
        log.Fatal(err)
        }
    }

    log.Println("Messages sent successfully!")
    fmt.Println("End of Sample amqpPut.go")
}
