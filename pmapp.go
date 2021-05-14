/*
 * Copyright: Oleg Borodin <onborodin@gmail.com>
 */

package main

import (
    "context"
    //"fmt"
    //"io"
    //"net/http"
    "os"
    //"path/filepath"
    "sync"
    "time"

    //"github.com/gin-gonic/gin"
    //"github.com/gorilla/websocket"

    "app/pmconfig"
    "app/pmdrivers"
    //"app/pmdaemon"
    "app/pmlog"
)

func main() {
    var err error
    app := NewApp()

    err = app.AppStart()
    if err != nil {
        pmlog.LogError("application error:", err)
        os.Exit(1)
    }
}

const (
    loopPeriod time.Duration    = 1000 // ms
    mqttUrl string = "mqtt://device:qwerty@v7.unix7.org:1883"
)

type Application struct {
    config      *pmconfig.Config
    drivers     []pmdrivers.Driverer
    context     context.Context
    cancel      context.CancelFunc
    wg          sync.WaitGroup
}

func NewApp() *Application {
    var app Application
    app.context, app.cancel = context.WithCancel(context.Background())
    app.config  = pmconfig.NewConfig()
    app.drivers = make([]pmdrivers.Driverer, 0)
    return &app
}

func (this *Application) AppStart() error {
    var err error

    //err = this.config.Setup()
    //if err != nil {
        //return err
    //}

    err = this.startDriver()
    if err != nil {
        return err
    }
    err = this.startLoop()
    if err != nil {
        return err
    }
    return err
}

func (this *Application) startDriver() error {
    var err error

    driverN1 := pmdrivers.NewMG1Driver()

    this.drivers = make([]pmdrivers.Driverer, 0)
    this.drivers = append(this.drivers, driverN1)

    for i := range this.drivers {
        err = this.drivers[i].InitializeDriver()
        if err != nil {
            return err
        }
    }

    for i := range this.drivers {
        err = this.drivers[i].SetConfig(pmdrivers.ConfigMqttUrlName, []byte(mqttUrl))
        if err != nil {
            return err
        }
    }

    for i := range this.drivers {
        err = this.drivers[i].ConnectDriver()
        if err != nil {
            return err
        }
        err = this.drivers[i].StartDriver()
        if err != nil {
            return err
        }
    }
    return err
}

func (this *Application) startLoop() error {
    var err error
    loopFunc := func() {
        pmlog.LogInfo("application loop started")
        timer := time.NewTicker(loopPeriod * time.Millisecond)
        for _ = range timer.C {
            now := int64(time.Now().Unix())
            switch {
                case now % 5 == 0:
                    pmlog.LogInfo("application is alive")
            }

        }
    }
    loopFunc()
    return err
}
//EOF
