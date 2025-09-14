# HIVE-2544 Test Coverage Matrix 

## Scenario 1 - Register and use a valid UDF
| Platform | Cluster Type(s) | Test Type  | Reasoning                   | Priority   | Test Case ID | Status         |
|----------|-----------------|------------|-----------------------------|------------|--------------|----------------|
| All | All | E2E | This is the primary success scenario. | High | TC-HIVE-2544-E2E-1 | ❌ Not Started |

## Scenario 2 - Attempt to register an invalid UDF
| Platform | Cluster Type(s) | Test Type  | Reasoning                   | Priority   | Test Case ID | Status         |
|----------|-----------------|------------|-----------------------------|------------|--------------|----------------|
| All | All | E2E | This is the primary negative scenario. | High | TC-HIVE-2544-E2E-2 | ❌ Not Started |

## Scenario 3 - Attempt to register a UDF with a name that is already in use
| Platform | Cluster Type(s) | Test Type  | Reasoning                   | Priority   | Test Case ID | Status         |
|----------|-----------------|------------|-----------------------------|------------|--------------|----------------|
| All | All | E2E | This is an edge case scenario. | Medium | TC-HIVE-2544-E2E-3 | ❌ Not Started |
