package main

import (
    "fmt"
    "log"
    "os"
    "io/ioutil"
    "encoding/json"
    "os/exec"
    "runtime"
    "sync"
    "net/http"
    "net/url"
    "net/http/cookiejar"
    "strings"
    "time"
)

const (
    defaultConfigFile   = "./sigsci-sounds.conf"
    apiURL             = "https://dashboard.signalsciences.net/api/v0"
    loginEndpoint      = apiURL + "/auth/login"
    interval            = 600
)

// Config Configuration for sigsci-sounds
type Config struct {
    Username string
    Password string
    CorpName string
    SiteName string
	Tags []struct {
		Name  string
		Sound string
        Text string
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
func initConfig() Config {	
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
        log.Fatal("File error: %v\n", e)
    }

    // decode json and load config object
    var c Config

    jsonErr := json.Unmarshal(file, &c)
    if jsonErr != nil {
        log.Fatal(jsonErr)
    }

    return c
}

func main() {
    fmt.Println("Initiating SigSci Sound!")
    fmt.Println("Enjoy the soothing sounds of attacks and anomalies...")
    fmt.Println("Press Ctrl+C to terminate.")
    runtime.GOMAXPROCS(2)

    var wg sync.WaitGroup
    var config Config
    var session []*http.Cookie

    // initialize configuration
    config = initConfig()

    // add WaitGroups for the number of tags in the configuration
    // concurrency implementaiton based on https://www.goinggo.net/2014/01/concurrency-goroutines-and-gomaxprocs.html
    wg.Add(len(config.Tags))
	
    // set Timeseries endpoint
    var timeseriesEndpoint = apiURL + "/corps/" + config.CorpName + "/sites/" + config.SiteName + "/timeseries/requests"

    // get credentials from configuration and authenticate to SigSci API
    form := url.Values{
        "email":    []string{config.Username},
        "password": []string{config.Password},
    }

    req, _ := http.NewRequest("POST", loginEndpoint, strings.NewReader(form.Encode()))
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    client                          := &http.Client{}
    var transport http.RoundTripper = &http.Transport{}
    resp, transportErr              := transport.RoundTrip(req)

    if transportErr != nil {
        log.Fatal(fmt.Sprintf("Error connecting to API: %v", transportErr))
    }

    // check for invalid login
    if("/login?p=invalid" == resp.Header["Location"][0]) {
        log.Fatal("Invalid Login")
    }

    // get session cookie and store in cookie jar
    session  = resp.Cookies()
    jar, _  := cookiejar.New(nil)
    u, _    := url.Parse(timeseriesEndpoint)
    jar.SetCookies(u, session)
    client.Jar = jar
    
    // for each tag in configuration launch a goroutine
    for i := range(config.Tags) {
        // get tag configuration
        tag   := config.Tags[i].Name
        sound := config.Tags[i].Sound
        text  := config.Tags[i].Text

        // go go gadget goroutine!
        go func() {
            // defer WaitGroup magic
            defer wg.Done()

            // initialize needed variables
            var command, content string
            var now        = int32(time.Now().Unix())
            var fromUntil = fmt.Sprintf("&from=%d&until=%d", now - interval, now)

            // start infinate loop
            for {
                // call timeseries API endpoint to get json payload
                req, _          = http.NewRequest("GET", timeseriesEndpoint + "?&tag=" + tag + fromUntil, nil)
                resp, clientErr := client.Do(req)

                if clientErr != nil {
                    log.Fatal(fmt.Sprintf("Error connecting to API: %v", clientErr))
                }

                defer resp.Body.Close()

                payload, ioErr := ioutil.ReadAll(resp.Body)

                if ioErr != nil {
                    log.Fatal(fmt.Sprintf("Unable to read API response: %v", ioErr))
                }

                // initialize Timeseries object and load json payload data
                var t Timeseries

                unmarshalErr := json.Unmarshal(payload, &t)
                if unmarshalErr != nil {
                    log.Fatal(unmarshalErr)
                }
                
                // determine which binary and content to use
                if("say" == sound) {
                    command = "say"
                    content = text
                } else {
                    command = "afplay"
                    content = sound

                    // verify sound file exists
                    _, fileErr := os.Stat(content)
	
                    if fileErr != nil {
                        log.Fatal("Sound file is missing: ", content)
                    }
                }
                
                // look for binary to execute
                binary, lookErr := exec.LookPath(command)

                if lookErr != nil {
                    log.Fatal(lookErr)
                }

                // set content as an argument for binary
                args := []string{content}
                
                // for each timeseries value of data,
                // sleep for 1 second and
                // if value is greator than 0, then
                // play sound
                if(0 != len(t.Data)){
                    for i := range(t.Data[0].Data) {
                        time.Sleep(time.Second)
                        if(t.Data[0].Data[i] > 0) {
                            execErr := exec.Command(binary, args...).Run()

                            if execErr != nil {
                                fmt.Println(tag)
                                log.Fatal(execErr)
                            }
                        }
                    }

                    // set new from and until values for next API call 
                    fromUntil = fmt.Sprintf("&from=%d&until=%d", t.Data[0].Until, t.Data[0].Until + interval)
                }

                // sleep for interval before doing it all over again and playing more sounds for this tag
                time.Sleep(time.Second * interval)
            }
        } ()
    }
 
    // wiat for WaitGroups
    wg.Wait()
    fmt.Println("\nTerminating SigSci Sound!")
}
