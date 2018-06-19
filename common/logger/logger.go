// Copyright 2018 Percona LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logger

import (
	"errors"
	"io"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	lcf "github.com/Robpol86/logrus-custom-formatter"
	"github.com/alecthomas/kingpin"
	log "github.com/sirupsen/logrus"
)

// enableVerboseLogging enables verbose logging
func enableVerboseLogging(ctx *kingpin.ParseContext) error {
	log.SetLevel(log.DebugLevel)
	return nil
}

// getCallerInfo returns the file and file line-number of a caller
func getLogCallerInfo(e *log.Entry, f *lcf.CustomFormatter) (interface{}, error) {
	var skip int = 1
	var skipMax int = 10
	for skip <= skipMax {
		_, file, lineNo, _ := runtime.Caller(skip)
		if strings.Contains(file, "github.com/Robpol86/logrus-custom-formatter") || strings.Contains(file, "github.com/sirupsen/logrus") {
			skip += 1
			continue
		}
		return filepath.Base(file) + ":" + strconv.Itoa(lineNo), nil
	}
	return "", errors.New("could not find caller file")
}

// GetLogFormatter returns a configured logrus.Formatter for logging
func GetLogFormatter(progName string) log.Formatter {
	template := "%[ascTime]s %-5[process]d " + progName + "  %-16[caller]s %-6[levelName]s %[message]s %[fields]s\n"
	return lcf.NewFormatter(template, lcf.CustomHandlers{"caller": getLogCallerInfo})
}

// SetupLogger configures github.com/srupsen/logrus for logging
func SetupLogger(app *kingpin.Application, formatter log.Formatter, out io.Writer) bool {
	log.SetOutput(out)
	log.SetFormatter(formatter)
	log.SetLevel(log.InfoLevel)
	if app != nil {
		var verbose bool
		app.Flag("verbose", "enable verbose logging").Action(enableVerboseLogging).BoolVar(&verbose)
		return verbose
	}
	return false
}
