package main

import (
	"encoding/json"
	"fabric-rest-api-go/pkg/api"
	"fabric-rest-api-go/pkg/handlers"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"log"
	"net/http"
	"os"
)

type ApiConfig struct {
	Org struct {
		Admin string `json:"admin"`
		Name  string `json:"name"`
	} `json:"org"`
	User struct {
		Name string `json:"name"`
	} `json:"user"`
	ConfigPath string `json:"configPath"`
}

func LoadConfiguration(file string) (*ApiConfig, error) {
	var config *ApiConfig
	configFile, err := os.Open(file)
	defer configFile.Close()
	if err != nil {
		return nil, errors.WithMessage(err, "Unable to open configuration file")
	}
	jsonParser := json.NewDecoder(configFile)
	err = jsonParser.Decode(&config)
	if err != nil {
		return nil, errors.WithMessage(err, "Unable to parse configuration file JSON")
	}
	return config, nil
}

func main() {

	// TODO merge config files
	var apiConfigPath string
	var sdkConfigPath string
	flag.StringVar(&apiConfigPath, "api-config", "./configs/config.json", "Path to API configuration file (example: -api-config=./config.json)")
	flag.StringVar(&sdkConfigPath, "sdk-config", "./configs/config.yaml", "Path to SDK configuration file (example: -sdk-config=./config.yaml)")
	flag.Parse()

	config, err := LoadConfiguration(apiConfigPath)
	if err != nil {
		panic(err)
	}

	api.FscInstance = api.FabricSdkClient{
		ConfigFile: sdkConfigPath,

		// Org parameters
		OrgAdmin: config.Org.Admin,
		OrgName:  config.Org.Name,

		// User parameters
		UserName: config.User.Name,
	}

	err = api.FscInstance.Initialize()
	if err != nil {
		panic(err)
	}

	fmt.Println("Start listening to localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", Router()))
}

func Router() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", handlers.WelcomeHandler)
	r.HandleFunc("/health", handlers.HealthCheckHandler)

	r.HandleFunc("/chaincodes/install", handlers.PostChaincodesInstallHandler).Methods("POST")
	r.HandleFunc("/chaincodes/instantiate", handlers.PostChaincodesInstantiateHandler).Methods("POST")

	r.HandleFunc("/chaincodes/installed", handlers.GetChaincodesInstalledHandler).Methods("GET")

	r.HandleFunc("/channels/{channelId}/chaincodes/instantiated", handlers.GetChaincodesInstantiatedHandler).Methods("GET") // TODO
	r.HandleFunc("/channels/{channelId}/chaincodes/{chaincodeId}/info", handlers.GetChaincodesInfoHandler).Methods("GET")
	r.HandleFunc("/channels", handlers.GetChannelsHandler).Methods("GET")
	r.HandleFunc("/channels", handlers.PostChannelsHandler).Methods("POST") // TODO

	r.HandleFunc("/channels/{channelId}", handlers.GetChannelsChannelIdHandler).Methods("GET")
	r.HandleFunc("/channels/{channelId}/orgs", handlers.GetChannelsChannelIdOrgsHandler).Methods("GET") // TODO
	r.HandleFunc("/channels/{channelId}/peers", handlers.GetChannelsChannelIdPeersHandler).Methods("GET")

	r.HandleFunc("/channels/{channelId}/chaincodes/{chaincodeId}/query", handlers.GetQueryHandler).Methods("GET")
	r.HandleFunc("/channels/{channelId}/chaincodes/{chaincodeId}/invoke", handlers.PostInvokeHandler).Methods("POST")

	r.HandleFunc("/init_test_fixtures", handlers.InitTestFixturesHandler).Methods("POST") // for test purposes

	return r
}
