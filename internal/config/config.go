package config

import (
	"os"
	"log"
	"fmt"
	"encoding/json"
)

const (
	configFileName = ".gatorconfig.json"
)

type Config struct {
	DbURL string `json:"db_url"`
	CurrentUsername string `json:"current_user_name"`
}

/* Reads the json configfile from the home directory and unmarshals the json data into a Config struct and returns the struct. */
func Read() Config {
	configFilePath := getConfigPath()
	data, err := os.ReadFile(configFilePath)
	if err != nil {
		log.Fatal(err)
	}
	var readConfig Config
	err = json.Unmarshal(data, &readConfig)
	if err != nil {
		log.Fatal(err)
	}
	return readConfig


}

/*Export a SetUser method on the Config struct that writes the config struct to the JSON file after setting the current_user_name field.*/
func (c *Config) SetUser(username string) {
	log.Printf("SetUser %s", username)
	if len(username) < 1 {
		return
	}
	c.CurrentUsername = username
	byte, err := json.Marshal(&c)
	if err != nil {
		log.Fatal(err)
	}
	os.WriteFile(getConfigPath(), byte, 0666)

}

func getConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	return fmt.Sprintf("%s/%s", homeDir, configFileName)
}
