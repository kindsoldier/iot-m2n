
package pmdrivers

import (
    "encoding/json"
    "errors"
    "context"
    "sync"
    "time"


    "app/mqtrans"
    "app/pmtools"
    "app/pmlog"
)

type Driverer interface {
    InitializeDriver() error
    ConnectDriver() error
    StartDriver() error
    StopDriver() error

    SetConfig(name string, value []byte) error
    GetConfig(name string) (value []byte, err error)

    //ToJson() []byte
}

type SubjectType string

const (
    mqttClassId    string = "53db7bc7-b406-11eb-900d-68f72872406b"

    ConfigMqttUrlName string = "MqttUrl"
    loopPeriod  time.Duration = 1000 // ms

    subjectTypeUnknown      SubjectType = "unknown"
    subjectTypeMqttTopic    SubjectType = "mqttTopic"

)

type MqttDriver struct {
    Id          string          `json:"id"          db:"id"`
    ClassId     string          `json:"classId"     db:"class_id"`
    Name        string          `json:"name"        db:"name"`
    ClassName   string          `json:"className"   db:"-"`

    Enabled     bool            `json:"enabled"     db:"enabled"`
    Hidden      bool            `json:"hidden"      db:"hidden"`

    Configs     []*Config       `json:"configs"     db:"-"`
    Indicators  []*Indicator    `json:"indicators"  db:"-"`
    Controls    []*Control      `json:"controls"    db:"-"`
    Subjects    []*Subject      `json:"subjects"    db:"-"`

    mqt     *mqtrans.Transport  `json:"-"           db:"-"`
    context context.Context     `json:"-"           db:"-"`
    cancel  context.CancelFunc  `json:"-"           db:"-"`
    wg      sync.WaitGroup      `json:"-"           db:"-"`
}

func (this *MqttDriver) ToJson() []byte {
    jsonBytes, _ := json.Marshal(this)
    return jsonBytes
}

func (this *MqttDriver) InitializeDriver() error {
    var err error
    this.Id           = pmtools.GetNewUUID()
    this.ClassId      = mqttClassId
    this.Enabled      = true
    this.Hidden       = false

    this.Configs      = make([]*Config, 0)
    this.Indicators   = make([]*Indicator, 0)
    this.Controls     = make([]*Control, 0)
    this.Subjects     = make([]*Subject, 0)

    this.Configs = append(this.Configs, this.NewMqttConfig())

    this.context, this.cancel = context.WithCancel(context.Background())

    this.mqt = mqtrans.NewTransport()
    return err
}

func (this *MqttDriver) NewMqttConfig() *Config {
    config := NewConfig()
    config.Name     = ConfigMqttUrlName
    config.Id       = pmtools.GetNewUUID()
    config.DriverId = this.Id
    return config
}

func (this *MqttDriver) ConnectDriver() error {
    var err error
    url, err := this.GetConfig(ConfigMqttUrlName)
    if err != nil {
        return err
    }
    this.mqt.Bind(string(url))
    return err
}

func (this *MqttDriver) StartDriver() error {
    var err error

    this.StartLoop()
    return err
}

func (this *MqttDriver) StartLoop() error {
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
        }
    }
    go loopFunc()
    return err
}

func (this *MqttDriver) StopDriver() error  {
    var err error
    this.cancel()
    this.wg.Wait()
    pmlog.LogInfo("driver", this.Id, "loop stopped")
    return err
}

func (this *MqttDriver) SetConfig(name string, value []byte) error {
    var err error
    for i := range this.Configs {
        if this.Configs[i].Name == name {
            this.Configs[i].Value = value
            return err
        }
    }
    err = errors.New("config not found")
    return err
}

func (this *MqttDriver) GetConfig(name string) ([]byte, error) {
    var err     error
    var result  []byte
    for i := range this.Configs {
        if this.Configs[i].Name == name {
            result = this.Configs[i].Value
            return result, err
        }
    }
    err = errors.New("config not found")
    return result, err
}

func (this *MqttDriver) SetIndicator(name string, value []byte) error {
    var err error
    for i := range this.Indicators {
        if this.Indicators[i].Name == name {
            this.Indicators[i].Value = value
            return err
        }
    }
    err = errors.New("config not found")
    return err
}

func (this *MqttDriver) GetIndicator(name string) ([]byte, error) {
    var err     error
    var result  []byte
    for i := range this.Indicators {
        if this.Indicators[i].Name == name {
            result = this.Indicators[i].Value
            return result, err
        }
    }
    err = errors.New("indicator not found")
    return result, err
}

func NewMqttDriver() *MqttDriver {
    var driver MqttDriver
    driver.InitializeDriver()
    return &driver
}
//
// Config
//
type Config struct {
    Id          string          `json:"id"          db:"id"`
    Value       []byte          `json:"value"       db:"value"`
    DriverId    string          `json:"driver_id"   db:"driver_id"`
    Hidden      bool            `json:"hidden"      db:"hidden"`
    Name        string          `json:"name"        db:"-"`
}

func NewConfig() *Config {
    var config Config
    config.Value = make([]byte, 0)
    return &config
}

func (this *Config) ToJson() []byte {
    jsonBytes, _ := json.Marshal(this)
    return jsonBytes
}
//
// Indicator
//
type Indicator struct {
    Id          string          `json:"id"          db:"id"`
    DriverId    string          `json:"driver_id"   db:"driver_id"`

    Name        string          `json:"name"        db:"-"`
    Value       []byte          `json:"value"       db:"-"`
    Proactive   bool            `json:"-"           db:"-"`
    Interval    int             `json:"-"           db:"-"`
    Enabled     bool            `json:"enabled"     db:"-"`
    Hidden      bool            `json:"hidden"      db:"-"`
}

func NewIndicator() *Indicator {
    var indicator Indicator
    return &indicator
}

func (this *Indicator) ToJson() []byte {
    jsonBytes, _ := json.Marshal(this)
    return jsonBytes
}
//
// Control
//
type Control struct {
    Id          string          `json:"id"          db:"id"`
    DriverId    string          `json:"driver_id"   db:"driver_id"`

    Name        string          `json:"name"        db:"-"`
    Enabled     bool            `json:"enabled"     db:"-"`
    Hidden      bool            `json:"hidden"      db:"-"`
}

func NewControl() *Control {
    var control Control
    return &control
}

func (this *Control) ToJson() []byte {
    jsonBytes, _ := json.Marshal(this)
    return jsonBytes
}
//
// Subject
//
type Subject struct {
    Type        SubjectType     `json:"type"        db:"-"`
    Id          string          `json:"id"          db:"id"`
    DriverId    string          `json:"driver_id"   db:"driver_id"`

    Value       []byte          `json:"value"       db:"value"`
    Enabled     bool            `json:"enabled"     db:"enabled"`
    Name        string          `json:"name"        db:"-"`
}

func NewSubject() *Subject {
    var subject Subject
    subject.Value   = make([]byte, 0)
    subject.Type    = subjectTypeUnknown
    return &subject
}

func (this *Subject) ToJson() []byte {
    jsonBytes, _ := json.Marshal(this)
    return jsonBytes
}
//EOF
