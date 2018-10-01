package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
	sigsci "github.com/signalsciences/go-sigsci"
)

const (
	defaultConfigFile = "themes/sigsci-sounds-osx.conf"
	apiURL            = "https://dashboard.signalsciences.net/api/v0"
	loginEndpoint     = apiURL + "/auth/login"
	interval          = 300
)

// Config Configuration for sigsci-sounds
type Config struct {
	Username string
	Password string
	CorpName string
	SiteName string
	Tags     []struct {
		Name  string
		Sound string
	}
}

// Timeseries Timeseries data from signal sciences
type Timeseries struct {
	Data []struct {
		Data  []int
		From  int
		Inc   int
		Label string
		Meta  struct {
			Lookup   int
			Quantize int
		}
		SummaryCount int
		TotalPoints  int
		Type         string
		Until        int
	}
}

// initConfig reads the configuration file and returns a config object.
func initConfig(initVariables bool) Config {
	// get configuration file path
	configFile := os.Getenv("SIGSCI_SOUNDS_CONFIG")

	if len(configFile) == 0 {
		configFile = defaultConfigFile
	}

	// verify the configuration file path
	_, err := os.Stat(configFile)

	if err != nil {
		log.Fatal("Config file is missing (see readme file for instructions): ", configFile)
	}

	// read the configuration file
	file, err := ioutil.ReadFile(configFile)

	if err != nil {
		log.Fatal("Error reading file!")
	}

	// decode json and load config object
	var c Config

	err := json.Unmarshal(file, &c)
	if err != nil {
		log.Fatal(err)
	}

	if initVariables {
		// override with ENV variable configs
		email := os.Getenv("SIGSCI_EMAIL")
		password := os.Getenv("SIGSCI_PASSWORD")
		corp := os.Getenv("SIGSCI_CORP")
		site := os.Getenv("SIGSCI_SITE")

		if len(email) != 0 {
			c.Username = email
		}

		if len(password) != 0 {
			c.Password = password
		}

		if len(corp) != 0 {
			c.CorpName = corp
		}

		if len(site) != 0 {
			c.SiteName = site
		}
	}

	return c
}

// APIRequest authenticates and makes a request to specified endpoint.
func APIRequest(username string, password string, corp string, site string, query url.Values, ch chan<- []sigsci.Timeseries) {
	sc, err := sigsci.NewClient(username, password)
	if err != nil {
		log.Fatal(err)
	}

	data, err := sc.GetTimeseries(corp, site, query)
	if err != nil {
		log.Fatal(err)
	}

	ch <- data
}

func playWAV(sound string) {
	f, _ := os.Open(sound)
	s, format, _ := wav.Decode(f)
	playing := make(chan struct{})

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	speaker.Play(beep.Seq(s, beep.Callback(func() {
		// Callback after the stream Ends
		close(playing)
	})))
	<-playing

	f.Close()
}

func playMP3(sound string) {
	f, _ := os.Open(sound)
	s, format, _ := mp3.Decode(f)
	playing := make(chan struct{})

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	speaker.Play(beep.Seq(s, beep.Callback(func() {
		// Callback after the stream Ends
		close(playing)
	})))
	<-playing

	f.Close()
}

func testConfig() {
	fmt.Println("Testing sound files for " + os.Getenv("SIGSCI_SOUNDS_CONFIG"))

	var config Config

	// initialize configuration
	config = initConfig(false)

	// for each tag in configuration launch a goroutine
	for i := range config.Tags {
		// get tag configuration
		tag := config.Tags[i].Name
		sound := config.Tags[i].Sound

		// verify sound file exists
		_, fileErr := os.Stat(sound)

		if fileErr != nil {
			log.Fatal("Sound file is missing: ", sound)
		}

		fmt.Println("Playing sound for " + tag + " (" + sound + ")")

		if strings.HasSuffix(sound, ".mp3") {
			playMP3(sound)
		} else {
			playWAV(sound)
		}
	}

	fmt.Println("Done!")
}

func main() {
	var test = false
	if len(os.Args) > 1 {

		if os.Args[1] == "test" {
			test = true
		}
	}
	var wg sync.WaitGroup

	fmt.Println("Initiating SigSci Sounds!")

	if test {
		testConfig()
	} else {
		fmt.Println("Enjoy the soothing sounds of attacks and anomalies...")
		fmt.Println("Press Ctrl+C to terminate.")
		runtime.GOMAXPROCS(2)

		var config Config

		// initialize configuration
		config = initConfig(true)

		// add WaitGroups for the number of tags in the configuration
		// concurrency implementation based on https://www.goinggo.net/2014/01/concurrency-goroutines-and-gomaxprocs.html
		wg.Add(len(config.Tags))

		apiResponseChannel := make(chan []sigsci.Timeseries)

		var now = int32(time.Now().Unix())

		for {
			var from = int(now - interval)
			var until = int(now)

			// for each tag in configuration launch a goroutine
			for i := range config.Tags {
				// get tag configuration
				tag := config.Tags[i].Name
				sound := config.Tags[i].Sound

				query := url.Values{}
				query.Add("from", strconv.Itoa(from))
				query.Add("until", strconv.Itoa(until))
				query.Add("tag", tag)

				go APIRequest(config.Username, config.Password, config.CorpName, config.SiteName, query, apiResponseChannel)

				var payload = <-apiResponseChannel

				_, fileErr := os.Stat(sound)

				if fileErr != nil {
					log.Fatal("Sound file is missing: ", sound)
				}

				if 0 != len(payload[0].Data) {
					for i := range payload[0].Data {
						time.Sleep(time.Second)

						if payload[0].Data[i] > 0 {

							if strings.HasSuffix(sound, ".mp3") {
								playMP3(sound)
							} else {
								playWAV(sound)
							}
						}
					}
				}
			}

			// sleep for interval before doing it all over again
			time.Sleep(time.Second * interval)
		}

	}

	// wait for WaitGroups
	wg.Wait()
	fmt.Println("\nTerminating SigSci Sounds!")
}
