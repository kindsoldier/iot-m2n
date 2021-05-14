
package pmdrivers

import (
    //"container/list"
    "time"
    "encoding/json"
    "sync"

    //"app/pmtools"
    "app/pmlog"
)

const (
    MG1ClassId    string = "53db7bc7-b406-11eb-900d-68f72872406b"

    statusTopicName  string = "StatusTopic"
    statusTopicValue string = "/gw/ac233fc0025f/status"
)

type MG1Driver struct {
    MqttDriver
    IBeacons    *IBeacons    `json:"iBeacons"`
}

func NewMG1Driver() *MG1Driver {
    var driver MG1Driver
    driver.InitializeDriver()
    return &driver
}

func (this *MG1Driver) InitializeDriver() error {
    var err error
    this.MqttDriver.InitializeDriver()
    this.ClassId = MG1ClassId
    this.IBeacons = NewIBeacons()

    this.Subjects = append(this.Subjects, this.NewStatusSubject())
    return err
}

func (this *MG1Driver) NewStatusSubject() *Subject {
    subject := NewSubject()
    subject.Name    = statusTopicName
    subject.Value   = []byte(statusTopicValue)
    subject.Type    = subjectTypeMqttTopic
    return subject
}

func (this *MG1Driver) StartDriver() error {
    var err error

    handler := func(subject string, payload []byte) {
        pmlog.LogDebug("driver", this.Id, "handled subject", subject)
        iBeacons := make([]IBeacon, 0)
        _ = json.Unmarshal(payload, &iBeacons)
        for i := range iBeacons {
            this.IBeacons.Add(&iBeacons[i])
        }
        pmlog.LogDebug(string(this.IBeacons.ToJson()))
    }

    //this.ConnectSubjects()
    for i := range this.Subjects {
        if this.Subjects[i].Type == subjectTypeMqttTopic {
            topic := string(this.Subjects[i].Value)
            this.mqt.Subscribe(topic, handler)
            pmlog.LogDebug("driver", this.Id, "subsribed to topic", topic)
        }
    }

    this.StartLoop()
    return err
}

func (this *MG1Driver) StartLoop() error {
    var err error

    loopFunc := func() {
        this.wg.Add(1)
        defer this.wg.Done()

        pmlog.LogInfo("driver", this.Id, "loop started")
        timer := time.NewTicker(loopPeriod * time.Millisecond)
        for _ = range timer.C {
            select {
                case <- this.context.Done():
                    pmlog.LogInfo("driver", this.Id, "loop canceled")
                    return
                default:
            }

            now := int64(time.Now().Unix())
            switch {
                case now % 5 == 0:
                    pmlog.LogInfo("driver", this.Id, "is alive")
            }

            this.IBeacons.Clean()
        }
    }
    go loopFunc()
    return err
}


//
// IBeacon
//
type IBeacon struct {
    Timestamp      time.Time `json:"timestamp"`
    Type           string    `json:"type"`
    Mac            string    `json:"mac"`
    //GatewayFree    int       `json:"gatewayFree,omitempty"`
    //GatewayLoad    float64   `json:"gatewayLoad,omitempty"`
    //BleName        string    `json:"bleName,omitempty"`
    Rssi           int       `json:"rssi,omitempty"`
    //RawData        string    `json:"rawData,omitempty"`
    //IbeaconUUID    string    `json:"ibeaconUUID,omitempty"`
    //IbeaconMajor   int       `json:"ibeaconMajor,omitempty"`
    //IbeaconMinor   int       `json:"ibeaconMinor,omitempty"`
    //IbeaconTxPower int       `json:"ibeaconTxPower,omitempty"`
    Battery        int       `json:"battery"`
}

func (this *IBeacon) ToJson() []byte {
    jsonBytes, _ := json.Marshal(this)
    return jsonBytes
}

func NewIBeacon() *IBeacon{
    return &IBeacon{}
}


//
// IBeacons
//
const (
    gatewayTypeLabel string = "Gateway"
)

type IBeacons struct {
    GatewayMac  string          `json:"gatewayMac"`
    List        []*IBeacon      `json:"list"`
    listMutex   sync.Mutex      `json:"-"`
}

func NewIBeacons() *IBeacons {
    var iBeacons IBeacons
    iBeacons.List = make([]*IBeacon, 0)
    return &iBeacons
}

func (this *IBeacons) ToJson() []byte {
    jsonBytes, _ := json.MarshalIndent(this, "", "    ")
    return jsonBytes
}

func (this *IBeacons) Add(beacon *IBeacon) {
    if beacon.Type == gatewayTypeLabel {
        this.GatewayMac = beacon.Mac
        return
    }
    this.listMutex.Lock()
    defer this.listMutex.Unlock()
    for i := range this.List {
        if this.List[i].Mac == beacon.Mac {
            this.List[i] = beacon
            return
            //this.List = append(this.List[:i], this.List[i:]...)
            //break
        }
    }
    this.List = append(this.List, beacon)
}

func (this *IBeacons) Delete(beacon *IBeacon) {
    beacon.Timestamp = time.Now()
    if beacon.Type == gatewayTypeLabel {
        return
    }
    this.listMutex.Lock()
    defer this.listMutex.Unlock()
    for i := range this.List {
        if this.List[i].Mac == beacon.Mac {
            this.List = append(this.List[:i], this.List[i:]...)
            return
        }
    }
}

func (this *IBeacons) Clean() {
    this.listMutex.Lock()
    defer this.listMutex.Unlock()

    tmpList := make([]*IBeacon, 0)
    for i := range this.List {
        //pmlog.LogDebug(time.Since(this.List[i].Timestamp))
        if time.Since(this.List[i].Timestamp) < (10 * time.Second) {
            tmpList = append(tmpList, this.List[i])
        }
    }
    if len(this.List) != len(tmpList) {
        this.List = tmpList
    }
}


//EOF
