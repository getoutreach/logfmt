// Copyright 2022 Outreach Corporation. All Rights Reserved.

package main_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"gotest.tools/v3/icmd"
)

type line struct {
	Level     string `json:"level"`
	Message   string `json:"message"`
	Timestamp string `json:"@timestamp"`
}

func TestLogfmtNoJSONText(t *testing.T) {
	txt := "Line 1.\nLine 2.\n"

	runLogfmt(t, txt, txt)
}

func TestLogfmtStructured(t *testing.T) {
	var buff bytes.Buffer
	enc := json.NewEncoder(&buff)

	enc.Encode(line{
		Level:     "info",
		Message:   "Hello, World!",
		Timestamp: (time.Time{}.Add(1 * time.Second)).Format(time.RFC3339Nano),
	})

	runLogfmt(t,
		buff.String(),
		`time="0001-01-01T00:00:01Z" level=info msg="Hello, World!"`)
}

func TestLogfmtStructuredFormat(t *testing.T) {
	var buff bytes.Buffer
	enc := json.NewEncoder(&buff)

	enc.Encode(line{Level: "info", Message: "msg1"})
	enc.Encode(line{Level: "info", Message: "msg2"})

	runLogfmt(t,
		buff.String(),
		"msg1\nmsg2\n",
		"--format", "{{ .message }}",
	)
}

func TestLogfmtStructuredFilter(t *testing.T) {
	var buff bytes.Buffer
	enc := json.NewEncoder(&buff)

	enc.Encode(line{Level: "info", Message: "info"})
	enc.Encode(line{Level: "warn", Message: "warn"})
	enc.Encode(line{Level: "error", Message: "error"})

	runLogfmt(t,
		buff.String(),
		"info\nerror\n",
		"--format", "{{ .message }}",
		"--filter", `select(.level != "warn")`,
	)
}

func runLogfmt(t *testing.T, input, output string, args ...string) {
	logs := strings.NewReader(input)

	args = append([]string{"run", `logfmt.go`}, args...)

	result := icmd.RunCmd(icmd.Command("go", args...), icmd.WithStdin(logs))
	result.Assert(t, icmd.Expected{
		ExitCode: 0,
		Err:      output,
	})
}
