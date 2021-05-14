/*
 * Copyright: Oleg Borodin <onborodin@gmail.com>
 */

package mqtrans

import (
    "fmt"
    "time"
    "net/url"

    "app/pmlog"

    mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
    keepalive   time.Duration   = 3 // sec
    waitTimeout time.Duration   = 3 // sec
    pingTimeout time.Duration   = 3 // sec

    QosL1       byte            = 1
    QosL2       byte            = 2
    QosL3       byte            = 4

    clientId    string          = "pm-xxxx"
)

type Transport struct {
    mc          mqtt.Client
    clientId    string
}

func NewTransport() *Transport {
    var transport Transport
    transport.clientId = clientId
    return &transport
}

func (this *Transport) Bind(mqRef string) error {
    var err error

    mqURL, err := url.Parse(mqRef)
    if err != nil {
        return err
    }

    hostname    := mqURL.Hostname()
    port        := mqURL.Port()

    username    := mqURL.User.Username()
    password, _ := mqURL.User.Password()

    opts := mqtt.NewClientOptions()

    uri := fmt.Sprintf("tcp://%s:%s", hostname, port)
    opts.AddBroker(uri)

    opts.SetUsername(username)
    opts.SetPassword(password)
    opts.SetClientID(this.clientId)

    //opts.SetOrderMatters(true)
    opts.SetAutoReconnect(true)

    opts.SetKeepAlive(keepalive)
    opts.SetPingTimeout(pingTimeout)

    onConnectHandler := func(client mqtt.Client) {
        pmlog.LogInfo("mqtt transport connected to broker:", hostname )
    }
    opts.SetOnConnectHandler(onConnectHandler)

    onReconnectHandler := func(client mqtt.Client, opts *mqtt.ClientOptions) {
        pmlog.LogInfo("mqtt transport reconnected to broker", hostname)
        time.Sleep(1 * time.Second)
    }
    opts.SetReconnectingHandler(onReconnectHandler)

    this.mc = mqtt.NewClient(opts)

    token := this.mc.Connect()
    for !token.WaitTimeout(waitTimeout * time.Second) {}

    err = token.Error()
    if err != nil {
        return err
    }
    return err
}

type Handler = func(topic string, payload []byte)

func (this *Transport) Publish(topic string, message string) error {
    var err error
    token := this.mc.Publish(topic, QosL1, false, message)
    for !token.WaitTimeout(waitTimeout * time.Second) {}
    err = token.Error()

    if err != nil {
        return err
    }
    return err
}

func (this *Transport) Subscribe(topic string, callback Handler) error {
    var err error

    mqttHandler := func(mqttClient mqtt.Client, mqttMessage mqtt.Message) {
        callback(mqttMessage.Topic(), mqttMessage.Payload())
    }

    token := this.mc.Subscribe(topic, QosL1, mqttHandler)
    for !token.WaitTimeout(waitTimeout * time.Second) {}

    err = token.Error()
    if err != nil {
        return err
    }
    return err
}

func (this *Transport) Disconnect() error {
    var err error
    if this.mc.IsConnected() {
        this.mc.Disconnect(1)
    }
    return err
}
