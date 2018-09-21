package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
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
	file, e := ioutil.ReadFile(configFile)

	if e != nil {
		log.Fatal("Error reading file!")
	}

	// decode json and load config object
	var c Config

	jsonErr := json.Unmarshal(file, &c)
	if jsonErr != nil {
		log.Fatal(jsonErr)
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
func APIRequest(username string, password string, endpoint string, ch chan<- string) {
	form := url.Values{
		"email":    []string{username},
		"password": []string{password},
	}

	var session []*http.Cookie
	req, _ := http.NewRequest("POST", loginEndpoint, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{}
	var transport http.RoundTripper = &http.Transport{}
	resp, transportErr := transport.RoundTrip(req)

	if transportErr != nil {
		log.Fatal(fmt.Sprintf("Error connecting to API: %v", transportErr))
	}

	// check for invalid login
	if "/login?p=invalid" == resp.Header["Location"][0] {
		log.Fatal("Invalid Login")
	}

	// get session cookie and store in cookie jar
	session = resp.Cookies()
	jar, _ := cookiejar.New(nil)
	u, _ := url.Parse(endpoint)
	jar.SetCookies(u, session)
	client.Jar = jar

	// call timeseries API endpoint to get json payload
	req, _ = http.NewRequest("GET", endpoint, nil)
	resp, clientErr := client.Do(req)

	if clientErr != nil {
		log.Fatal(fmt.Sprintf("Error connecting to API: %v", clientErr))
	}

	defer resp.Body.Close()

	payload, ioErr := ioutil.ReadAll(resp.Body)

	if ioErr != nil {
		log.Fatal(fmt.Sprintf("Unable to read API response: %v", ioErr))
	}

	ch <- fmt.Sprintf("%s", payload)
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
	if(len(os.Args) > 1) {

		if(os.Args[1] == "test") {
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

		// set Timeseries endpoint
		var timeseriesEndpoint = apiURL + "/corps/" + config.CorpName + "/sites/" + config.SiteName + "/timeseries/requests"

		apiResponseChannel := make(chan string)

		var now = int32(time.Now().Unix())

		for {
			var fromUntil = fmt.Sprintf("&from=%d&until=%d", now-interval, now)

			// for each tag in configuration launch a goroutine
			for i := range config.Tags {
				// get tag configuration
				tag := config.Tags[i].Name
				sound := config.Tags[i].Sound
				endpoint := timeseriesEndpoint + "?tag=" + tag + fromUntil

				go APIRequest(config.Username, config.Password, endpoint, apiResponseChannel)

				var payload = <-apiResponseChannel

				// initialize Timeseries object and load json payload data
				var t Timeseries

				unmarshalErr := json.Unmarshal([]byte(payload), &t)
				if unmarshalErr != nil {
					log.Fatal(unmarshalErr)
				}

				// verify sound file exists
				_, fileErr := os.Stat(sound)

				if fileErr != nil {
					log.Fatal("Sound file is missing: ", sound)
				}

				if 0 != len(t.Data) {
					for i := range t.Data[0].Data {
						time.Sleep(time.Second)
						if t.Data[0].Data[i] > 0 {

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
