# HIVE-2302 Test Coverage Matrix

## Scenario 1 - Modern Cluster Provisioning with metadata.json Secret
| Platform | Cluster Type(s) | Test Type  | Reasoning                   | Priority   | Test Case ID | Status         |
|----------|-----------------|------------|-----------------------------|------------|--------------|----------------|
| AWS | Standard | E2E | Core functionality - automatic Secret creation and generic destroyer | High | HIVE-2302_001 | ❌ Not Started |
| Azure | Standard | E2E | Core functionality - automatic Secret creation and generic destroyer | High | HIVE-2302_002 | ❌ Not Started |
| GCP | Standard | E2E | Core functionality - automatic Secret creation and generic destroyer | High | HIVE-2302_003 | ❌ Not Started |

## Scenario 2 - Legacy Cluster Retrofit from ClusterProvision
| Platform | Cluster Type(s) | Test Type  | Reasoning                   | Priority   | Test Case ID | Status         |
|----------|-----------------|------------|-----------------------------|------------|--------------|----------------|
| AWS | Standard | E2E | Validate automatic Secret generation from existing ClusterProvision metadataJSON | High | HIVE-2302_004 | ❌ Not Started |

## Scenario 3 - Legacy Cluster Retrofit from Bespoke Fields
| Platform | Cluster Type(s) | Test Type  | Reasoning                   | Priority   | Test Case ID | Status         |
|----------|-----------------|------------|-----------------------------|------------|--------------|----------------|
| AWS | Standard | E2E | Validate fallback Secret generation from legacy bespoke fields when ClusterProvision unavailable | High | HIVE-2302_005 | ❌ Not Started |

## Scenario 4 - Legacy Deprovision Escape Hatch
| Platform | Cluster Type(s) | Test Type  | Reasoning                   | Priority   | Test Case ID | Status         |
|----------|-----------------|------------|-----------------------------|------------|--------------|----------------|
| AWS | Standard | E2E | Verify annotation-based legacy destroyer path still functional | Medium | HIVE-2302_006 | ❌ Not Started |

## Scenario 5 - AWS with HostedZoneRole (Shared VPC)
| Platform | Cluster Type(s) | Test Type  | Reasoning                   | Priority   | Test Case ID | Status         |
|----------|-----------------|------------|-----------------------------|------------|--------------|----------------|
| AWS | Shared VPC | Manual | Requires pre-existing hosted zone in another AWS account - cannot automate in standard E2E environment | High | HIVE-2302_M001 | ❌ Not Started |

## Scenario 6 - Azure with ResourceGroupName
| Platform | Cluster Type(s) | Test Type  | Reasoning                   | Priority   | Test Case ID | Status         |
|----------|-----------------|------------|-----------------------------|------------|--------------|----------------|
| Azure | Custom RG | Manual | Requires pre-created resource group - manual setup needed | High | HIVE-2302_M002 | ❌ Not Started |

## Scenario 7 - GCP with NetworkProjectID (Shared VPC)
| Platform | Cluster Type(s) | Test Type  | Reasoning                   | Priority   | Test Case ID | Status         |
|----------|-----------------|------------|-----------------------------|------------|--------------|----------------|
| GCP | Shared VPC | Manual | Requires shared VPC setup with separate network project - complex manual infrastructure | High | HIVE-2302_M003 | ❌ Not Started |

## Scenario 8 - vSphere Pre-Zonal Metadata Format
| Platform | Cluster Type(s) | Test Type  | Reasoning                   | Priority   | Test Case ID | Status         |
|----------|-----------------|------------|-----------------------------|------------|--------------|----------------|
| vSphere | Standard | Manual | Requires vSphere infrastructure and specific pre-zonal release version | Medium | HIVE-2302_M004 | ❌ Not Started |

## Scenario 9 - vSphere Post-Zonal Metadata Format
| Platform | Cluster Type(s) | Test Type  | Reasoning                   | Priority   | Test Case ID | Status         |
|----------|-----------------|------------|-----------------------------|------------|--------------|----------------|
| vSphere | Standard | Manual | Requires vSphere infrastructure with zonal support | Medium | HIVE-2302_M005 | ❌ Not Started |

## Scenario 10 - IBMCloud Deprovision
| Platform | Cluster Type(s) | Test Type  | Reasoning                   | Priority   | Test Case ID | Status         |
|----------|-----------------|------------|-----------------------------|------------|--------------|----------------|
| IBMCloud | Standard | Manual | Requires IBMCloud infrastructure access | Medium | HIVE-2302_M006 | ❌ Not Started |

## Scenario 11 - Nutanix Deprovision
| Platform | Cluster Type(s) | Test Type  | Reasoning                   | Priority   | Test Case ID | Status         |
|----------|-----------------|------------|-----------------------------|------------|--------------|----------------|
| Nutanix | Standard | Manual | Requires Nutanix infrastructure access | Medium | HIVE-2302_M007 | ❌ Not Started |

## Scenario 12 - OpenStack Deprovision
| Platform | Cluster Type(s) | Test Type  | Reasoning                   | Priority   | Test Case ID | Status         |
|----------|-----------------|------------|-----------------------------|------------|--------------|----------------|
| OpenStack | Standard | Manual | Requires OpenStack infrastructure access | Medium | HIVE-2302_M008 | ❌ Not Started |
