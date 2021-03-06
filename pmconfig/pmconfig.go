/*
 * Copyright: Oleg Borodin <onborodin@gmail.com>
 */

package pmconfig

import (
    "fmt"
    "flag"
    "io/ioutil"
    "path/filepath"
    "encoding/json"
    "os"

    "github.com/go-yaml/yaml"
)

type Config struct {
    DataDir             string  `yaml:"datadir"`
    PidPath             string  `yaml:"pidfile"`
    MessageLogPath      string  `yaml:"messagelog"`
    AccessLogPath       string  `yaml:"accesslog"`
    ConfigPath          string  `yaml:"-"`
    LibDir              string  `yaml:"-"`

    DbConfig    DbConfig    `yaml:"dbConfig"    json:"dbConfig"`
    WebConfig   WebConfig   `yaml:"webConfig"   json:"webConfig"`
}

type ProcConfig struct {
    ConfigPath          string  `yaml:"-"`
    LibDir              string  `yaml:"-"`
    DataDir             string  `yaml:"datadir"`
    PidPath             string  `yaml:"pidfile"`
    MessageLogPath      string  `yaml:"messagelog"`
    AccessLogPath       string  `yaml:"accesslog"`

}

type WebConfig struct {
    Port        int         `yaml:"hostname"    json:"hostname"`
}

type DbConfig struct {
    Hostname    string      `yaml:"hostname"    json:"hostname"`
    Port        int         `yaml:"port"        json:"port"`
    Username    string      `yaml:"username"    json:"username"`
    Password    string      `yaml:"password"    json:"password"`
    Database    string      `yaml:"database"    json:"database"`
}

func NewConfig() *Config {
    webConfig := WebConfig{
        Port:       8090,
    }
    dbConfig := DbConfig{
        Hostname:   "localhost",
        Port:       5432,
        Username:   "pgsql",
        Password:   "qwert",
    } 
    return &Config{
        ConfigPath:     "/usr/local/etc/pmapp/pmapp.yml",
        LibDir:         "/usr/local/share/pmapp",
        DataDir:        "/var/db/pmapp",

        PidPath:        "/var/run/pmapp/pmapp.pid",
        MessageLogPath: "/var/log/pmapp/message.log",
        AccessLogPath:  "/var/log/pmapp/access.log",

        DbConfig:   dbConfig,
        WebConfig:  webConfig,
    }
}

func (this *Config) Json() string {
    json, _ := json.MarshalIndent(this, "", "    ")
    return string(json)
}

func (this *Config) Yaml() string {
    json, _ := yaml.Marshal(this)
    return string(json)
}

func (this *Config) Write(fileName string) error {
    var data []byte
    var err error

    os.Rename(fileName, fileName + "~")

    data, err = yaml.Marshal(this)
    if err != nil {
        return err
    }
    return ioutil.WriteFile(fileName, data, 0640)
}

func (this *Config) Read(fileName string) error {
    var data []byte
    var err error

    data, err = ioutil.ReadFile(fileName)
    if  err != nil {
        return err
    }
    return yaml.Unmarshal(data, &this)
}

func (this *Config) Setup() error {
    var err error
    flag.StringVar(&this.DbConfig.Hostname, "host", this.DbConfig.Hostname, "database hostname")
    flag.IntVar(&this.DbConfig.Port, "port", this.DbConfig.Port, "database port")
    flag.StringVar(&this.DbConfig.Username, "user", this.DbConfig.Username, "database username")
    flag.StringVar(&this.DbConfig.Password, "pass", this.DbConfig.Password, "database password")
    flag.StringVar(&this.DbConfig.Database, "data", this.DbConfig.Database, "database name")

    exeName := filepath.Base(os.Args[0])

    flag.Usage = func() {
        fmt.Println(exeName)
        fmt.Println("")
        fmt.Printf("usage: %s command [option]\n", exeName)
        fmt.Println("")
        flag.PrintDefaults()
        fmt.Println("")
    }
    flag.Parse()
    return err
}

func (this *Config) GetDbURL() string {
    return fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
                this.DbConfig.Username,
                this.DbConfig.Password,
                this.DbConfig.Hostname,
                this.DbConfig.Port,
                this.DbConfig.Database)
}

func (this *Config) GetListenParam() string {
    return fmt.Sprintf(":%d", this.WebConfig.Port)
}
//EOF
