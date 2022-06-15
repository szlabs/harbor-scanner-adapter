/*
 * Harbor Scanner Adapter API
 *
 * ## Overview  This API must be implemented in order to register a new artifact scanner in [Harbor](https://goharbor.io) registry.  The [/scan](#operation/AcceptScanRequest) and [/scan/{scan_request_id}/report](#operation/GetScanReport) operations are responsible for the actual scanning and return a scan report that is visible in the Harbor web console.  The [/scan](#operation/AcceptScanRequest) operation is asynchronous. It should enqueue the job for processing a scan request and return the identifier. This allows Harbor to poll a corresponding scan report with the [/scan/{scan_request_id}/report](#operation/GetScanReport) operation. Harbor will call the [/scan/{scan_request_id}/report](#operation/GetScanReport) operation periodically periodically until it returns 200 or 500 status codes.  The [/metadata](#operation/GetMetadata) operation allows a Harbor admin to configure and register a scanner and discover its capabilities.  ## Supported consumed MIME spec  - `application/vnd.oci.image.manifest.v1+json` - `application/vnd.docker.distribution.manifest.v2+json`  ## Supported produced MIME spec  - `application/vnd.scanner.adapter.vuln.report.harbor+json; version=1.0` - `application/vnd.security.vulnerability.report; version=1.1` - `application/vnd.scanner.adapter.vuln.report.raw` - `application/vnd.security.cis.report; version=1.0`
 *
 * API version: 1.2
 * Contact: cncf-harbor-maintainers@lists.cncf.io
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package spec

// Severity : A standard scale for measuring the severity of a vulnerability.  * `Unknown` - either a security problem that has not been assigned to a priority yet or a priority that the   scanner did not recognize. * `Negligible` - technically a security problem, but is only theoretical in nature, requires a very special   situation, has almost no install base, or does no real damage. * `Low` - a security problem, but is hard to exploit due to environment, requires a user-assisted attack,   a small install base, or does very little damage. * `Medium` - a real security problem, and is exploitable for many people. Includes network daemon denial of   service attacks, cross-site scripting, and gaining user privileges. * `High` - a real problem, exploitable for many people in a default installation. Includes serious remote denial   of service, local root privilege escalations, or data loss. * `Critical` - a world-burning problem, exploitable for nearly all people in a default installation. Includes   remote root privilege escalations, or massive data loss.
type Severity string

// List of Severity
const (
	UNKNOWN    Severity = "Unknown"
	NEGLIGIBLE Severity = "Negligible"
	LOW        Severity = "Low"
	MEDIUM     Severity = "Medium"
	HIGH       Severity = "High"
	CRITICAL   Severity = "Critical"
)
