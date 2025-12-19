# CloudSlash User Guide

This guide details the operation of CloudSlash for identifying AWS infrastructure waste and performing safe remediation.

---

## 1. Quick Start

### Installation

Refer to the [README](README.md) for installation instructions specific to the operating system.

### Authentication

CloudSlash utilizes standard AWS credentials located in `~/.aws/credentials` or `~/.aws/config`.

1.  **Configure AWS CLI**: `aws configure`
2.  **SSO Login** (if applicable): `aws sso login`
3.  **Permissions**: Ensure the user or role has `ReadOnlyAccess` and `ViewOnlyAccess`.

---

## 2. Workflows

### Standard Scan (Single Account)

Run CloudSlash to scan the default profile configured in `~/.aws/credentials`.

```bash
cloudslash
```

### Multi-Account Scan

To scan **every AWS account** defined in the local configuration:

```bash
cloudslash --license PRO-YOUR-KEY --all-profiles
```

- **Behavior**: Iterates through all profiles in `~/.aws/config`, checking for waste in each, and making a consolidated report.

### Cost Analysis

CloudSlash queries the AWS Price List API to calculate the monthly cost of identified waste.

- **Example**: A 50GB `gp2` volume unused for 30 days is estimated at ~$5.00/mo waste.

---

## 3. Artifacts and Remediation

After a scan completes, the `cloudslash-out/` directory contains the following artifacts:

### 1. Dashboard (`dashboard.html`)

Open in any web browser. Contains charts (Cost by Resource Type) and a sortable risk table.

### 2. Safeguard (`safe_cleanup.sh`)

Helper script for remediation.

- **Logic**: For every resource ID listed in the waste report, it attempts to create a snapshot or backup before deletion.

### 3. Terraform (`waste.tf`)

(Pro Feature) Generates Terraform code representing the waste resources, allowing them to be imported via `terraform import` managed destruction via IaC.

### 4. Reverse Terraform (`fix_terraform.sh`)

(Pro Feature) Generates a script to remove the waste resources from the Terraform State files (`.tfstate`), preventing them from being recreated by Terraform during the next apply.

### 5. Suppression (`ignore_resources.sh`)

If specific resources should be retained but excluded from reports:

1.  Review the `cloudslash-out/ignore_resources.sh` script.
2.  Run it: `bash cloudslash-out/ignore_resources.sh`
3.  It applies a `cloudslash:ignore=true` tag.
4.  Subsequent scans will respect this tag and exclude the resource from waste reports.

---

## 4. Heuristic Logic

| Resource          | Logic                                                           |
| :---------------- | :-------------------------------------------------------------- |
| **EBS Volumes**   | State is `available` OR attached to instance stopped > 30 days. |
| **EC2 Instances** | Max CPU < 5% over 7 days (Right-Sizing).                        |
| **NAT Gateways**  | Traffic < 1GB/mo AND Active Connections < 5.                    |
| **S3 Buckets**    | Incomplete Multipart Uploads > 7 days old.                      |
| **RDS**           | DB connections == 0 for 7 days OR status `stopped`.             |
| **Elastic IPs**   | Not attached to any network interface.                          |

---

## 5. Command Reference

| Flag                     | Description                                                            |
| :----------------------- | :--------------------------------------------------------------------- |
| `--license [KEY]`        | Activate Pro features.                                                 |
| `--region [REGION]`      | Override default region (e.g., `us-west-2`).                           |
| `--all-profiles`         | Scan all configured AWS profiles.                                      |
| `--tfstate [PATH]`       | Path to local `.tfstate` for drift detection.                          |
| `--required-tags [LIST]` | Comma-separated list of tags that must be present (e.g., `Env,Owner`). |
| `--slack-webhook [URL]`  | Send summary report to Slack.                                          |

### Commands

- **scan**: Run a headless scan without the TUI (useful for CI/CD).
- **update**: Update CloudSlash to the latest version.
- **completion**: Generate autocompletion script for shell.
