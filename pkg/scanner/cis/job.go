// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cis

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"time"

	"github.com/spf13/viper"

	"github.com/szlabs/goworker/pkg/errs"
	"github.com/szlabs/goworker/pkg/job"
	"github.com/szlabs/harbor-scanner-adapter/pkg/auth"
	"github.com/szlabs/harbor-scanner-adapter/pkg/store"
	"github.com/szlabs/harbor-scanner-adapter/pkg/store/data"
	"github.com/szlabs/harbor-scanner-adapter/pkg/zlog"
)

const (
	// JobName is the name of the CIS scan job.
	JobName     = "CIS_SCAN"
	concurrency = 10
)

// AddToKnownList adds cis.Job to the known list.
func AddToKnownList(l *job.KnownList) error {
	return l.AddKnownJob(JobName, job.RunnableFunc(Job), job.Concurrency(concurrency))
}

// Job for CIS scan.
func Job(_ context.Context, parameters job.Parameters) (err error) {
	// Skip parameter validation as the parameter should be validated in the API layer.

	errorf := errs.WithPrefix("cis scan job error")
	resStore := store.Default()
	lg := zlog.Logger()

	// Extract key parameters.
	reqID := parameters[ParamReqID].(string)
	imagePath := parameters[ParamKeyImage].(string)
	dataKey := &data.Key{
		Provider: Name,
		ReqID:    reqID,
		Mimetype: ReportMimeType,
	}
	dataKey.AppendPrefix(dataPrefix)

	uk := fmt.Sprintf("%s:%s:%s", dataPrefix, Name, imagePath)
	if err := resStore.Unique(uk); err != nil {
		return errorf.Wrap("last scan is still not finished yet, skip job run", err)
	}

	defer func() {
		// Defer to de-unique.
		if e := resStore.DeUnique(uk); e != nil {
			// Just log.
			lg.Error(e)
		}

		// Check if error is occurred.
		if err != nil {
			if e := resStore.SaveResult(dataKey, &data.Item{
				Timestamp: time.Now().UTC().Unix(),
				Status:    data.Error,
				Error:     err.Error(),
			}); e != nil {
				lg.Error(e)
			}
		}
	}()

	// Mark status to start.
	if err := resStore.SaveResult(dataKey, &data.Item{
		Timestamp: time.Now().UTC().Unix(),
		Status:    data.Ongoing,
	}); err != nil {
		// Just need to log
		lg.Error(err)
	}

	args := buildOptions(parameters)

	// Specify a temp file for keeping scan output.
	f, err := ioutil.TempFile("", reqID)
	if err != nil {
		return errorf.Wrap("create temp scan result file error", err)
	}
	defer func() {
		if err := os.Remove(f.Name()); err != nil {
			lg.Error(err)
		}
	}()

	// Append image path for scan.
	args = append(args, "--output", f.Name(), imagePath)

	lg.Infow("cis scan configurations", "image", imagePath)

	dt, err := exec.Command("dockle", args...).CombinedOutput()
	if err != nil {
		return errorf.Wrap("run backend engine command error: %s=%s", err, "details", dt)
	}

	result, err := readScanResult(f.Name())
	if err != nil {
		return errorf.Wrap("parse scan result error", err)
	}

	// Save data.
	if err := resStore.SaveResult(dataKey, &data.Item{
		Status:    data.Success,
		JSON:      string(result),
		Timestamp: time.Now().UTC().Unix(),
	}); err != nil {
		lg.Error(err)
	} else {
		lg.Infow("cis.job is completed")
	}

	return nil
}
func buildOptions(params job.Parameters) []string {
	// Build arguments.
	args := []string{
		"-f", "json",
		"--username", params[auth.ParamKeyUsername].(string),
		"--password", params[auth.ParamKeyPassword].(string),
	}

	// Get configurations.
	timeout := viper.GetString("scanner.backends.cis.timeout")
	insecure := viper.GetBool("scanner.backends.cis.insecure")
	ignores := viper.GetString("scanner.backends.cis.ignore")
	certPath := viper.GetString("scanner.backends.cis.certPath")

	// Append corresponding options if related configurations are set.
	if len(timeout) > 0 {
		args = append(args, "-t", timeout)
	}
	if insecure {
		args = append(args, "--insecure")
	}
	if len(ignores) > 0 {
		args = append(args, "--ignore", ignores)
	}
	if len(certPath) > 0 {
		args = append(args, "--cert-path", certPath)
	}

	return args
}

func readScanResult(tmpFile string) ([]byte, error) {
	errorf := errs.WithPrefix("")

	bytes, err := ioutil.ReadFile(tmpFile)
	if err != nil {
		return nil, errorf.Wrap("read result temp file error", err)
	}

	return bytes, nil
}
