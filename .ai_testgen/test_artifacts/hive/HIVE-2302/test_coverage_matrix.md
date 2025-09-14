# HIVE-2302 Test Coverage Matrix

## Scenario 1 - Modern Cluster Provision and Deprovision with metadata.json Secret
| Platform | Cluster Type(s) | Test Type  | Reasoning                   | Priority   | Test Case ID | Status         |
|----------|-----------------|------------|-----------------------------|------------|--------------|----------------|
| AWS | Standard | E2E | Modern clusters get metadata.json Secret automatically - primary scenario | Critical | HIVE-2302_E001 | ❌ Not Started |
| Azure | Standard | E2E | Verify metadata.json Secret creation on Azure platform | High | HIVE-2302_E001 | ❌ Not Started |
| GCP | Standard | E2E | Verify metadata.json Secret creation on GCP platform | High | HIVE-2302_E001 | ❌ Not Started |
| vSphere | Standard | E2E | Verify metadata.json Secret creation on vSphere platform | High | HIVE-2302_E001 | ❌ Not Started |

## Scenario 2 - Deprovision Resource Leak Detection (Regression)
| Platform | Cluster Type(s) | Test Type  | Reasoning                   | Priority   | Test Case ID | Status         |
|----------|-----------------|------------|-----------------------------|------------|--------------|----------------|
| AWS | Standard | E2E | Verify no resources leaked after deprovision using metadata.json | Critical | HIVE-2302_E002 | ❌ Not Started |
| Azure | Standard | E2E | Verify no resources leaked after deprovision using metadata.json | High | HIVE-2302_E002 | ❌ Not Started |
| GCP | Standard | E2E | Verify no resources leaked after deprovision using metadata.json | High | HIVE-2302_E002 | ❌ Not Started |
| vSphere | Standard | E2E | Verify no resources leaked after deprovision using metadata.json | High | HIVE-2302_E002 | ❌ Not Started |

## Scenario 3 - Legacy Cluster Retrofit
| Platform | Cluster Type(s) | Test Type  | Reasoning                   | Priority   | Test Case ID | Status         |
|----------|-----------------|------------|-----------------------------|------------|--------------|----------------|
| AWS | Legacy | Manual | Requires clusters created before HIVE-2302 or upgrade scenario | High | HIVE-2302_M001 | ❌ Not Started |
| Azure | Legacy | Manual | Requires clusters created before HIVE-2302 or upgrade scenario | Medium | HIVE-2302_M001 | ❌ Not Started |

## Scenario 4 - AWS Shared VPC with HostedZoneRole
| Platform | Cluster Type(s) | Test Type  | Reasoning                   | Priority   | Test Case ID | Status         |
|----------|-----------------|------------|-----------------------------|------------|--------------|----------------|
| AWS | Shared VPC | Manual | Requires pre-configured AWS shared VPC infrastructure | High | HIVE-2302_M002 | ❌ Not Started |

## Scenario 5 - GCP Shared VPC with NetworkProjectID
| Platform | Cluster Type(s) | Test Type  | Reasoning                   | Priority   | Test Case ID | Status         |
|----------|-----------------|------------|-----------------------------|------------|--------------|----------------|
| GCP | Shared VPC | Manual | Requires pre-configured GCP shared VPC infrastructure | High | HIVE-2302_M003 | ❌ Not Started |

## Scenario 6 - Azure Existing Resource Group
| Platform | Cluster Type(s) | Test Type  | Reasoning                   | Priority   | Test Case ID | Status         |
|----------|-----------------|------------|-----------------------------|------------|--------------|----------------|
| Azure | Existing RG | Manual | Requires pre-created Azure resource group | High | HIVE-2302_M004 | ❌ Not Started |

## Scenario 7 - vSphere Pre-Zonal and Post-Zonal Metadata Format
| Platform | Cluster Type(s) | Test Type  | Reasoning                   | Priority   | Test Case ID | Status         |
|----------|-----------------|------------|-----------------------------|------------|--------------|----------------|
| vSphere | Pre-zonal | Manual | Requires specific vSphere release before zonal support (version control) | Medium | HIVE-2302_M005 | ❌ Not Started |
| vSphere | Post-zonal | Manual | Requires specific vSphere release after zonal support (version control) | Medium | HIVE-2302_M005 | ❌ Not Started |
