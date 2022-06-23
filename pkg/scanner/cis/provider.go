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
	"github.com/szlabs/harbor-scanner-adapter/pkg/client"
	"github.com/szlabs/harbor-scanner-adapter/pkg/oci"
	"github.com/szlabs/harbor-scanner-adapter/pkg/scan"
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
	// ParamReqID is parameter key of request ID.
	ParamReqID = "reqID"
	// ParamArtifact is the artifact reference.
	ParamArtifact = "artifact"

	dataPrefix = "{cis-result-store}"
)

// Use singleton provider.
var provider *Provider
var once sync.Once

// Provider to support CIS scan.
type Provider struct {
	store  store.Provider
	name   string
	dataNS string
}

// New a CIS provider.
func New() *Provider {
	once.Do(func() {
		provider = &Provider{
			store:  store.Default(),
			name:   Name,
			dataNS: dataPrefix,
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
func (p *Provider) AcceptScanRequest(ctx context.Context, req *spec.ScanRequest) (*spec.ScanResponse, error) {
	errorf := errs.WithPrefix("accept scan request error")

	// Extract request ID first.
	reqID := uuid.FromContext(ctx)
	if reqID == "" {
		return nil, errorf.Error("missing request ID in the context")
	}

	// Validate image mimetype.
	if req.Artifact.MimeType != oci.OCIImage && req.Artifact.MimeType != oci.DockerV2Image {
		return nil, errorf.Error("only support mimetypes: %s,%s", oci.OCIImage, oci.DockerV2Image)
	}

	// Convert job parameters.
	jp, err := toJobParams(req)
	if err != nil {
		return nil, errorf.Wrap("parse job parameters error", err)
	}

	// Append request ID to job parameters.
	jp[ParamReqID] = reqID

	// Enqueue scan job.
	enq, err := client.Enqueuer()
	if err != nil {
		return nil, errorf.Wrap("get job enqueuer error", err)
	}

	j, err := enq.EnqueueUnique(JobName, jp)
	if err != nil {
		return nil, errorf.Wrap("enqueue cis scan job error", err)
	}

	// Log the backend job info for potential debug.
	zlog.Logger().Infow("CIS backend scan job is enqueued", "job", j.Name, "id", j.ID)

	dk := &data.Key{
		ReqID:    reqID,
		Provider: p.name,
		Mimetype: ReportMimeType,
	}
	dk.AppendPrefix(p.dataNS)

	// Create result placeholder and set the status to pending.
	if err := p.store.SaveResult(dk, &data.Item{
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
func (p *Provider) RetrieveScanResult(_ context.Context, reqID string, mimetype string) (scan.Result, error) {
	errorf := errs.WithPrefix("retrieve scan request error")

	dk := &data.Key{
		Provider: p.name,
		ReqID:    reqID,
		Mimetype: mimetype,
	}
	dk.AppendPrefix(p.dataNS)

	res := &Result{}
	dt, err := p.store.GetResult(dk)
	if err != nil {
		if errors.Is(err, rds.NotFoundErr) {
			// Use res default phase which is not found.
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
		return nil, errorf.Error("unknown status: %s=%v", "status", dt.Status)
	}
}

func toJobParams(req *spec.ScanRequest) (job.Parameters, error) {
	errorf := errs.WithPrefix("")

	// Parse authorization credential.
	// NOTES: so far, only no authorization or basic authorization is supported.
	ap, err := auth.Parse(req.Registry.Authorization)
	if err != nil {
		return nil, errorf.Wrap("parse registry authorization error", err)
	}

	jp := make(job.Parameters)
	if err := ap.Inject(jp); err != nil {
		return nil, errorf.Wrap("inject auth params error", err)
	}

	imgPath := fmt.Sprintf("%s/%s", req.Registry.Url, req.Artifact.Repository)
	if req.Artifact.Tag != "" {
		imgPath = fmt.Sprintf("%s:%s", imgPath, req.Artifact.Tag)
	} else {
		imgPath = fmt.Sprintf("%s@%s", imgPath, req.Artifact.Digest)
	}
	jp[ParamKeyImage] = imgPath

	/*bytes, err := json.Marshal(req.Artifact)
	if err != nil {
		return nil, errorf.Wrap("marshal artifact error", err)
	}*/

	jp[ParamArtifact] = req.Artifact

	return jp, nil
}
