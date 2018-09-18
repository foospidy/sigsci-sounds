package main

import (
    "fmt"
    "log"
    "os"
    "io/ioutil"
    "encoding/json"
    "time"
    "strings"
    "github.com/faiface/beep"
    "github.com/faiface/beep/wav"
    "github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

const (
    defaultConfigFile = "themes/sigsci-sounds-osx.conf"
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
	}
}

func playMp3(sound string) {
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

func playWav(sound string) {
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
    fmt.Println("Initiating SigSci Sounds!")
    fmt.Println("Testing sound files for " + os.Getenv("SIGSCI_SOUNDS_CONFIG"))

    var config Config

    // initialize configuration
    config = initConfig()

    // for each tag in configuration launch a goroutine
    for i := range(config.Tags) {
        // get tag configuration
        tag   := config.Tags[i].Name
        sound := config.Tags[i].Sound

        // verify sound file exists
        _, fileErr := os.Stat(sound)

        if fileErr != nil {
            log.Fatal("Sound file is missing: ", sound)
        }

        fmt.Println("Playing sound for " + tag + " (" + sound + ")")

        if(strings.HasSuffix(sound, ".mp3")) {
            playMp3(sound)
        } else {
            playWav(sound)
        }
    }

    fmt.Println("Done!")
}
