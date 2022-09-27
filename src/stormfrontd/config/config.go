package config

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strconv"
)

const DEFAULT_CONFIG_PATH = "/var/stormfront/config.json"
const ENV_PREFIX = "STORMFRONTD_"

type ConfigObject struct {
	DaemonHost          string   `json:"daemon_host" env:"DAEMON_HOST"`
	DaemonPort          int      `json:"daemon_port" env:"DAEMON_PORT"`
	AllowedIPs          []string `json:"allowed_ips" env:"ALLOWED_IPS"`
	RestrictRequestHost bool     `json:"restrict_request_host" env:"RESTRICT_REQUEST_HOST"`
	ClientPort          int      `json:"client_port" env:"CLIENT_PORT"`
	InterfaceName       string   `json:"interface_name" env:"INTERFACE_NAME"`
}

var Config ConfigObject

func LoadConfig() {
	configPath := os.Getenv(ENV_PREFIX + "CONFIG_PATH")
	if configPath == "" {
		configPath = DEFAULT_CONFIG_PATH
	}

	Config = ConfigObject{
		DaemonHost:          "localhost",
		DaemonPort:          6674,
		AllowedIPs:          []string{"127.0.0.1"},
		RestrictRequestHost: true,
		ClientPort:          6626,
		InterfaceName:       "eth0",
	}

	if _, err := os.Stat(configPath); errors.Is(err, os.ErrNotExist) {
		configData, _ := json.MarshalIndent(Config, "", " ")

		_ = ioutil.WriteFile(configPath, configData, 0644)
	}

	jsonFile, err := os.Open(configPath)
	if err != nil {
		log.Println("Unable to read json file")
		panic(err)
	}

	log.Printf("Successfully Opened %v", configPath)

	byteValue, _ := ioutil.ReadAll(jsonFile)

	json.Unmarshal(byteValue, &Config)

	v := reflect.ValueOf(Config)
	t := reflect.TypeOf(Config)

	for i := 0; i < v.NumField(); i++ {
		field, found := t.FieldByName(v.Type().Field(i).Name)
		if !found {
			continue
		}

		value := field.Tag.Get("env")
		if value != "" {
			val, present := os.LookupEnv(ENV_PREFIX + value)
			if present {
				w := reflect.ValueOf(&Config).Elem().FieldByName(t.Field(i).Name)
				x := getAttr(&Config, t.Field(i).Name).Kind().String()
				if w.IsValid() {
					switch x {
					case "int", "int64":
						i, err := strconv.ParseInt(val, 10, 64)
						if err == nil {
							w.SetInt(i)
						}
					case "int8":
						i, err := strconv.ParseInt(val, 10, 8)
						if err == nil {
							w.SetInt(i)
						}
					case "int16":
						i, err := strconv.ParseInt(val, 10, 16)
						if err == nil {
							w.SetInt(i)
						}
					case "int32":
						i, err := strconv.ParseInt(val, 10, 32)
						if err == nil {
							w.SetInt(i)
						}
					case "string":
						w.SetString(val)
					case "float32":
						i, err := strconv.ParseFloat(val, 32)
						if err == nil {
							w.SetFloat(i)
						}
					case "float", "float64":
						i, err := strconv.ParseFloat(val, 64)
						if err == nil {
							w.SetFloat(i)
						}
					case "bool":
						i, err := strconv.ParseBool(val)
						if err == nil {
							w.SetBool(i)
						}
					default:
						objValue := reflect.New(field.Type)
						objInterface := objValue.Interface()
						err := json.Unmarshal([]byte(val), objInterface)
						obj := reflect.ValueOf(objInterface)
						if err == nil {
							w.Set(reflect.Indirect(obj).Convert(field.Type))
						} else {
							log.Println(err)
						}
					}
				}
			}
		}
	}
}

func getAttr(obj interface{}, fieldName string) reflect.Value {
	pointToStruct := reflect.ValueOf(obj) // addressable
	curStruct := pointToStruct.Elem()
	if curStruct.Kind() != reflect.Struct {
		panic("not struct")
	}
	curField := curStruct.FieldByName(fieldName) // type: reflect.Value
	if !curField.IsValid() {
		panic("not found:" + fieldName)
	}
	return curField
}
