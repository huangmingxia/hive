# HIVE-2923 Test Coverage Matrix

## Scenario 1 - DNSZone Status Stability with Invalid AWS Credentials
| Platform | Cluster Type(s) | Test Type  | Reasoning                   | Priority   | Test Case ID | Status         |
|----------|-----------------|------------|-----------------------------|------------|--------------|----------------|
| AWS | Standard, Managed DNS | E2E | Reproduces controller thrashing issue; validates resourceVersion stability over time; measures quantitative impact of RequestID scrubbing | High | TC-HIVE-2923-001 | ❌ Not Started |


## Scenario 2 - Error Message Scrubbing Validation
| Platform | Cluster Type(s) | Test Type  | Reasoning                   | Priority   | Test Case ID | Status         |
|----------|-----------------|------------|-----------------------------|------------|--------------|----------------|
| AWS | Standard, Managed DNS | Manual | Validates regex pattern handles multiple RequestID formats; requires manual inspection of error messages; tests edge cases (multiple IDs, lowercase, embedded newlines) | High | TC-HIVE-2923-002 | ❌ Not Started |
