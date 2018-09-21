# sigsci-sounds

Listen to the soothing sounds of attacks and anomalies.

[![Build Status](https://travis-ci.org/foospidy/sigsci-sounds.svg?branch=master)](https://travis-ci.org/foospidy/sigsci-sounds)
[![Go Report Card](https://goreportcard.com/badge/github.com/foospidy/sigsci-sounds)](https://goreportcard.com/report/github.com/foospidy/sigsci-sounds)

## Description

The [Signal Sciences](https://signalsciences.com) web protection platform offers a rich API that enables limitless integrations and automation. Or, at least only limited by your imagination. As an example of this, sigsci-sounds is a utility to audibility monitor when Signal Sciences detects attack or anomaly events against your web site. You literally can listen to your site being attacked.

The utility is configurable so you can define which attacks or anomalies you want to hear.

Note: obviously you must be a Signal Sciences customer to make use of this utility.

## Requirements

- Golang 1.7 or higher.
- A Signal Sciences API account.

## Instructions

__Be sure to have your `GOPATH` set properly.__

If you have `make` on your system run `make all` to build the binary. Once built, you can run `./sigsci-sounds`. Or you can run `make run` to run sigsci-sounds without building the binary. If you do not have `make`, You can run `go build -i sigsci-sounds.go` to build the binary.

### Configure SigSci API Info

When accessing the [Signal Sciences API](https://docs.signalsciences.net/api/) you need to specify your credentials (email and password), corp name, and site name. You can do this via environment variables, example:

```bash
export SIGSCI_EMAIL=<your email>
export SIGSCI_PASSWORD=<your password>
export SIGSCI_CORP=<your corp name>
export SIGSCI_SITE=<site name of site to be monitored by sigsci-sounds>
```

### Configure sigsci-sounds Themes

By default sigsci-sounds will use the theme file `themes/sigsci-sounds-osx.conf`. However, you can use a different theme by setting the `SIGSCI_SOUNDS_CONFIG` environment variable, e.g. `export SIGSCI_SOUNDS_CONFIG=theme/star-trek-tos.conf`.

Note: You can edit theme files and include your API information (email, password, corp, site), rather than setting those values as environment variables.

## Themes

You can edit and create themes!

The theme file is JSON format. See one of the provided [configuration files](https://github.com/foospidy/sigsci-sounds/blob/master/themes/star-trek-tos.conf) as an example.

To customize a theme you only need to edit the `"Tags": []` section of the file. This section is a JSON array where each entry requires two fields: name and sound.

- __name__ is the actual tag "short name" you want to monitor. This can be a Signal Sciences default sytem tag or a custom tag.
- __sound__ is the path to the sound file you want to play for the specified tag name.
