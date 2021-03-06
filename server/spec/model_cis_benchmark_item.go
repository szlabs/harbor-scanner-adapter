/*
 * Harbor Scanner Adapter API
 *
 * ## Overview  This API must be implemented in order to register a new artifact scanner in [Harbor](https://goharbor.io) registry.  The [/scan](#operation/AcceptScanRequest) and [/scan/{scan_request_id}/report](#operation/GetScanReport) operations are responsible for the actual scanning and return a scan report that is visible in the Harbor web console.  The [/scan](#operation/AcceptScanRequest) operation is asynchronous. It should enqueue the job for processing a scan request and return the identifier. This allows Harbor to poll a corresponding scan report with the [/scan/{scan_request_id}/report](#operation/GetScanReport) operation. Harbor will call the [/scan/{scan_request_id}/report](#operation/GetScanReport) operation periodically periodically until it returns 200 or 500 status codes.  The [/metadata](#operation/GetMetadata) operation allows a Harbor admin to configure and register a scanner and discover its capabilities.  ## Supported consumed MIME types  - `application/vnd.oci.image.manifest.v1+json` - `application/vnd.docker.distribution.manifest.v2+json`  ## Supported produced MIME types  - `application/vnd.scanner.adapter.vuln.report.harbor+json; version=1.0` - `application/vnd.security.vulnerability.report; version=1.1` - `application/vnd.scanner.adapter.vuln.report.raw` - `application/vnd.security.cis.report; version=1.0`
 *
 * API version: 1.2
 * Contact: cncf-harbor-maintainers@lists.cncf.io
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package spec

type ModelMap map[string]string

type CISBenchmarkItem struct {
	// The unique identifier of the CIS benchmark item.
	Code string `json:"code"`
	// The link uri for the details.
	Link string `json:"link,omitempty"`
	// The concrete description of the CIS benchmark.
	Title string `json:"title"`

	Level *CISLevel `json:"level"`
	// More details about the violation if applicable.
	Alerts []string `json:"alerts,omitempty"`

	VendorAttributes ModelMap `json:"vendor_attributes,omitempty"`
}
