[![Go Report Card](https://goreportcard.com/badge/github.com/foospidy/sigsci-sounds)](https://goreportcard.com/report/github.com/foospidy/sigsci-sounds)

# sigsci-sounds
Listen to the soothing sounds of attacks and anomalies.

### Description

The [Signal Sciences](https://signalsciences.com) web protection platform offers a rich API that enables limitless integrations and automation. Or, at least only limited by your imagination. As an example of this, sigsci-sounds is a utility to audibility monitor when Signal Sciences detects attack or anomaly events against your web site. You literally can listen to your site being attacked.

The utility is configurable so you can define which attacks or anomalies you want to hear.

Note: obviously you must be a Signal Sciences customer to make use of this utility.

### Requirements

- Golang 1.7 or higher.
- A Signal Sciences API account.

### Instructions

If you have `make` on your system run `make install`, be sure your `$GOPATH` is set. Or you can run `make run` to build and run sigsci-sounds.

By default sigsci-sounds will look for the configuration file in its current directory, e.g. `./sigsci-sounds.conf`. However, you can export a different location using the variable `SIGSCI_SOUNDS_CONFIG`, e.g. `export SIGSCI_SOUNDS_CONFIG=/etc/sigsci/sigsci-sounds.conf`.

The configuration file is JSON format. It requires your API account username and password, corp name and site name you want to monitor, and at least one tag entry. See the provided [configuration file] (https://github.com/foospidy/sigsci-sounds/blob/master/themes/star-trek-tos.conf) as an example.

#### Defining a tag entry

Each entry requires two fields: name and sound.

- __name__ is the actual tag "short name" you want to monitor. This can be a Signal Sciences default sytem tag, or a custom tag.
- __sound__ is the path to the sound file you want to play for the specified tag name.
