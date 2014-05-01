package main

import (
	"bytes"
	"code.google.com/p/goconf/conf"
	"fmt"
	"io"
	"net"
	"os"
	"reflect"
	"strconv"
	"strings"
)

var (
	defaultConfig = AkadokConfig{
		Name:           "Akadok Server",
		Host:           net.IPv4(0, 0, 0, 0),
		Port:           1518,
		MaxPlayers:     16,
		Maps:           []string{"map1", "map2", "map3"},
		Verbose:        false,
		LogFile:        "server.log",
		AutoInsecure:   false,
		RegisterServer: true,
		TickRate:       15,
		Insecure:       false,
	}
	// Simplified default config elements
	writtenElements = map[string]bool{
		"Name":       true,
		"Port":       true,
		"MaxPlayers": true,
		"Maps":       true,
		"LogFile":    true,
	}
)

type AkadokConfig struct {
	Name           string
	Host           net.IP
	Port           int
	MaxPlayers     int
	Maps           []string
	Verbose        bool
	LogFile        string
	AutoInsecure   bool
	RegisterServer bool
	TickRate       int
	Insecure       bool
}

// LoadConfig tries to load config from the disk or creates it if necessary
func LoadConfig() {
	// Check for config file replacement in command line parameters
	args := os.Args
	if len(args) > 1 {
		configFile = os.Args[1]
	}

	// First check for config file existence
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		// Default config creation
		fmt.Println("Config file not found. Creating the default file.")
		WriteDefaultConfig()
	} else if err != nil {
		panic("Error accessing configuration file. Try changing permissions!")
	}

	cfg, err := conf.ReadConfigFile("server.cfg")
	if err != nil {
		panic("Config error!")
	}
	config = new(AkadokConfig)

	// Default config reflection to find fields to read
	reflectedDefCfg := reflect.ValueOf(defaultConfig)
	// New config reflection for setting values
	reflectedCfg := reflect.ValueOf(config).Elem()

	for i := 0; i < reflectedCfg.NumField(); i++ {
		field := reflectedCfg.Field(i)
		fieldName := NormalizeName(reflectedDefCfg.Type().Field(i).Name)
		fieldValue := reflectedDefCfg.Field(i).Interface()
		val := fieldValue

		switch reflect.TypeOf(fieldValue).String() {
		// Easier type checking (especially for non-primitive types) using string conversion

		case "string":
			val, err := cfg.GetString("default", fieldName)
			if err == nil {
				fieldValue = val
			}
			field.Set(reflect.ValueOf(fieldValue.(string)))

		case "int":
			val, err := cfg.GetInt("default", fieldName)
			if err == nil {
				fieldValue = val
			}
			field.Set(reflect.ValueOf(fieldValue.(int)))

		case "bool":
			val, err := cfg.GetBool("default", fieldName)
			if err == nil {
				fieldValue = val
			}
			field.Set(reflect.ValueOf(fieldValue.(bool)))

		case "[]string":
			serializedString, err := cfg.GetString("default", fieldName)
			if err == nil {
				serializedString = strings.Replace(serializedString, ", ", ",", 0)
				val = strings.Split(serializedString, ",")
				field.Set(reflect.ValueOf(val.([]string)))
			}

		case "net.IP":
			serializedIP, err := cfg.GetString("default", fieldName)
			var fieldValue net.IP
			if err != nil || serializedIP == "0.0.0.0" {
				fieldValue = nil
			} else {
				fieldValue = net.ParseIP(serializedIP)
			}
			field.Set(reflect.ValueOf(fieldValue))

		default:
			panic("Unknown configuration directive " + fieldName)
		}
	}
}

// NormalizeName turns an internal config entry name into a better name
// for configuration files (UpperCamelCase to lower_snake_case)
func NormalizeName(name string) string {
	buf := bytes.NewBuffer(nil)
	for i := 0; i < len(name); i++ {
		char := string(name[i])
		if name[i] >= 'A' && name[i] <= 'Z' && i != 0 {
			buf.WriteString("_")
			buf.WriteString(strings.ToLower(char))
		} else if i == 0 {
			buf.WriteString(strings.ToLower(char))
		} else {
			buf.WriteString(char)
		}
	}
	return buf.String()
}

// WriteDefaultConfig generates a clean config file from default options
func WriteDefaultConfig() {
	cfg := conf.NewConfigFile()

	// Default config generation by reflection
	reflectedCfg := reflect.ValueOf(defaultConfig)

	for i := 0; i < reflectedCfg.NumField(); i++ {

		fieldName := reflectedCfg.Type().Field(i).Name
		fieldValue := reflectedCfg.Field(i).Interface()

		if !writtenElements[fieldName] {
			continue
		}

		switch reflect.TypeOf(fieldValue).String() {
		// Same hack as for loadConfig()
		case "string":
			cfg.AddOption("default", fieldName, fieldValue.(string))
		case "int":
			cfg.AddOption("default", fieldName, strconv.Itoa(fieldValue.(int)))
		case "bool":
			stringValue := "off"
			if fieldValue.(bool) {
				stringValue = "on"
			}
			cfg.AddOption("default", fieldName, stringValue)
		case "[]string":
			cfg.AddOption("default", fieldName,
				strings.Join(fieldValue.([]string), ", "))
		default:
			panic("Unknown configuration directive " + fieldName)
		}
	}

	writeBuf := new(ConfigFileCleaner)
	cfg.Write(writeBuf, "Akadok default config. Edit as you want!")
}

// Type to improve config file generation
type ConfigFileCleaner struct {
	w io.Writer
}

// Write cleans up the file and writes it
func (w *ConfigFileCleaner) Write(p []byte) (n int, err error) {
	bufString := string(p)
	// Small cleaning of the written file to fit our needs
	bufString = strings.Replace(bufString, "[default]\n", "", 1)
	bufString = strings.Replace(bufString, "\n\n", "\n", 1)

	// File creation
	var file *os.File
	if file, err = os.Create(configFile); err != nil {
		return 0, err
	}

	buf := bytes.NewBuffer([]byte(bufString))
	buf.WriteTo(file)

	return buf.Len(), file.Close()
}
