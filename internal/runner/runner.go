// Copyright 2022 Outreach Corporation. All Rights Reserved.

// Description: Package runner implements the bulk of logfmt functionality.

// Package runner implements the bulk of logfmt functionality.
package runner

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/Masterminds/sprig/v3"
	"github.com/itchyny/gojq"
	"github.com/sirupsen/logrus"
)

// traceEvent is trace
const traceEvent = "trace"

// New returns a new runner
func New(log *logrus.Logger, filter, format string) *Runner {
	r := &Runner{log: log}
	r.filter = r.filterf(filter)
	r.format = r.formatf(format)
	r.formatStr = format
	return r
}

// Runner is a logger that emits to datadog and formats
// to logrus
type Runner struct {
	log    *logrus.Logger
	filter func(v interface{}) bool
	format func(v interface{}) string

	formatStr string
}

// logrus converts data set to the runner to the logrus
// format.
func (r *Runner) logrus(data map[string]interface{}) {
	logrusLog := logrus.NewEntry(r.log)

	level := ""
	message := ""
	logType := ""

	if v, ok := data["event_name"]; ok {
		if v.(string) == traceEvent {
			logType = traceEvent
		}
	}

	for k, v := range data {
		// skip empty keys
		if sv, ok := v.(string); ok && sv == "" {
			continue
		}

		switch k {
		case "level":
			level = v.(string)
			continue
		case "message":
			if logType == traceEvent {
				message = "(trace) "
			}
			message += v.(string)
			continue
		case "@timestamp":
			if t, err := time.Parse(time.RFC3339Nano, v.(string)); err == nil {
				logrusLog = logrusLog.WithTime(t)
			}
			continue
		case "app.version", "deployment.namespace", "app.name", "timing.service_time", "timing.dequeued_at",
			"timing.finished_at", "timing.scheduled_at", "timing.total_time", "timing.wait_time", "honeycomb.trace_id",
			"event_name":
			continue
		}
		logrusLog = logrusLog.WithField(k, v)
	}

	switch strings.ToLower(level) {
	case "info":
		logrusLog.Info(message)
	case "warn":
		logrusLog.Warn(message)
	case "error", "fatal":
		logrusLog.Error(message)
	}
}

// Run starts the runner
func (r *Runner) Run() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := scanner.Text()
		var data map[string]interface{}
		err := json.Unmarshal([]byte(text), &data)
		if err != nil {
			fmt.Fprintln(r.log.Out, text)
			continue
		}

		if r.filter(data) {
			if r.formatStr != "" {
				fmt.Fprintln(r.log.Out, r.format(data))
			} else {
				r.logrus(data)
			}
		}
	}
}

// filterf filters out logs
func (r *Runner) filterf(s string) func(interface{}) bool {
	if s == "" {
		return func(_ interface{}) bool {
			return true
		}
	}

	query, err := gojq.Parse(s)
	r.must(err)

	code, err := gojq.Compile(query)
	r.must(err)

	return func(data interface{}) bool {
		result, ok := code.Run(data).Next()
		if !ok {
			return false
		}
		if err, ok := result.(error); ok {
			r.must(err)
		}
		return true
	}
}

// formatf formats the logs with the given format string
func (r *Runner) formatf(s string) func(interface{}) string {
	jsonFormat := func(v interface{}) string {
		if text, ok := v.(string); ok {
			return text
		}
		data, err := json.Marshal(v)
		r.must(err)
		return string(data)
	}

	if s == "" {
		return jsonFormat
	}

	t := template.Must(template.New("format").Funcs(sprig.TxtFuncMap()).Parse(s))
	return func(v interface{}) string {
		var buf bytes.Buffer
		if err := t.Execute(&buf, v); err != nil {
			r.log.Errorf("format template error %v", err)
		}
		return buf.String()
	}
}

// must is a handler for panics
func (r *Runner) must(err error) {
	if err != nil {
		r.log.Fatal("error ", err)
	}
}
