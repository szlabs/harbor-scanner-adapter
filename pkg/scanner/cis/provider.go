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
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/szlabs/goworker/pkg/errs"
	"github.com/szlabs/goworker/pkg/job"

	"github.com/szlabs/harbor-scanner-adapter/pkg/auth"
	"github.com/szlabs/harbor-scanner-adapter/pkg/oci"
	"github.com/szlabs/harbor-scanner-adapter/pkg/runner"
	"github.com/szlabs/harbor-scanner-adapter/pkg/scan"
	"github.com/szlabs/harbor-scanner-adapter/pkg/scan/cis"
	"github.com/szlabs/harbor-scanner-adapter/pkg/store"
	"github.com/szlabs/harbor-scanner-adapter/pkg/store/data"
	"github.com/szlabs/harbor-scanner-adapter/pkg/store/rds"
	"github.com/szlabs/harbor-scanner-adapter/pkg/uuid"
	"github.com/szlabs/harbor-scanner-adapter/pkg/zlog"
	"github.com/szlabs/harbor-scanner-adapter/server/spec"
)

const (
	// ParamKeyImage is parameter key of image path.
	ParamKeyImage = "image"
	cisProvider   = "cis-provider"
)

// Use singleton provider.
var provider *Provider
var once sync.Once

// Provider to support CIS scan.
type Provider struct {
	store store.Provider
}

// New a CIS provider.
func New() *Provider {
	once.Do(func() {
		provider = &Provider{
			store: store.Default(),
		}
	})

	return provider
}

// Metadata implements scanner.Provider.
func (p *Provider) Metadata() *spec.ScannerAdapterMetadata {
	return &spec.ScannerAdapterMetadata{
		Scanner: &spec.Scanner{
			Name:    Name,
			Version: Version,
			Vendor:  Vendor,
		},
		Capabilities: []spec.ScannerCapability{
			{
				ConsumesMimeTypes: []string{
					oci.OCIImage,
					oci.DockerV2Image,
				},
				ProducesMimeTypes: []string{
					ReportMimeType,
				},
			},
		},
		Properties: ExtraMeta,
	}
}

// AcceptScanRequest implements scanner.Provider.
func (p *Provider) AcceptScanRequest(_ context.Context, req *spec.ScanRequest) (*spec.ScanResponse, error) {
	errorf := errs.WithPrefix("accept scan request error")

	// Convert job parameters.
	jp, err := toJobParams(req)
	if err != nil {
		return nil, errorf.Wrap("parse job parameters error", err)
	}

	if err := p.store.Unique(cisProvider, jp[ParamKeyImage].(string)); err != nil {
		return nil, errorf.Wrap("last scan is still not finished yet", err)
	}

	// Generate req uuid and append to job parameters.
	reqID := uuid.Random()
	jp["reqID"] = reqID

	// Enqueue scan job.
	enq, err := runner.Enqueuer()
	if err != nil {
		return nil, errorf.Wrap("get job enqueuer error", err)
	}

	j, err := enq.EnqueueUnique(cis.Name, jp)
	if err != nil {
		return nil, errorf.Wrap("enqueue cis scan job error", err)
	}

	// Log the backend job info for potential debug.
	zlog.Logger().Info("CIS backend scan job is enqueued", "job", *j)

	// Create result placeholder and set the status to pending.
	if err := p.store.SaveResult(reqID, &data.Item{
		Timestamp: time.Now().UTC().Unix(),
		Status:    data.Pending,
	}); err != nil {
		// Not a panic case, just logged it.
		// Once the scan job is started, it can be recovered again.
		zlog.Logger().Errorw("save result placeholder failed", "error", err)
	}

	return &spec.ScanResponse{
		Id: reqID,
	}, nil
}

// RetrieveScanResult implements scanner.Provider.
func (p *Provider) RetrieveScanResult(_ context.Context, reqID string) (scan.Result, error) {
	errorf := errs.WithPrefix("retrieve scan request error")

	res := &cis.Result{}

	dt, err := p.store.GetResult(reqID)
	if err != nil {
		if errors.Is(err, rds.NotFoundErr) {
			return res, nil
		}
		return nil, errorf.Wrap("store get result error", err)
	}

	switch dt.Status {
	case data.Pending, data.Ongoing:
		return res, res.Write(nil, scan.Phase(scan.ResultPhaseNotReady))
	case data.Error:
		return nil, fmt.Errorf(dt.Error)
	case data.Success:
		return res, res.Write(
			dt.JSON,
			scan.Phase(scan.ResultPhaseReady),
			scan.NextTry(0), // skip retry
		)
	default:
		return nil, errorf.Error("unknown status", "status", dt.Status)
	}
}

func toJobParams(req *spec.ScanRequest) (job.Parameters, error) {
	// Parse authorization credential.
	// NOTES: so far, only no authorization or basic authorization is supported.
	ap, err := auth.Parse(req.Registry.Authorization)
	if err != nil {
		return nil, fmt.Errorf("parse registry authorization error: %w", err)
	}

	jp := make(job.Parameters)
	if err := ap.Inject(jp); err != nil {
		return nil, fmt.Errorf("inject auth params error: %w", err)
	}

	imgPath := fmt.Sprintf("%s/%s", req.Registry.Url, req.Artifact.Repository)
	if req.Artifact.Tag != "" {
		imgPath = fmt.Sprintf("%s:%s", imgPath, req.Artifact.Tag)
	} else {
		imgPath = fmt.Sprintf("%s@%s", imgPath, req.Artifact.Digest)
	}
	jp[ParamKeyImage] = imgPath

	return jp, nil
}
