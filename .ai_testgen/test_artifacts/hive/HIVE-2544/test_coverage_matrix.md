# HIVE-2544 Test Coverage Matrix

## Scenario 1 - MachinePool in Non-Zoned Azure Region
| Platform | Cluster Type(s) | Test Type  | Reasoning                   | Priority   | Test Case ID | Status         |
|----------|-----------------|------------|-----------------------------|------------|--------------|----------------|
| Azure | Standard | E2E | Core fix validation - verify MachineSet generation succeeds in regions without AZ support | High | HIVE-2544_001 | ❌ Not Started |

## Scenario 2 - MachinePool in Non-Zoned Azure Government Region
| Platform | Cluster Type(s) | Test Type  | Reasoning                   | Priority   | Test Case ID | Status         |
|----------|-----------------|------------|-----------------------------|------------|--------------|----------------|
| Azure Government | Standard | Manual | Azure Gov requires special infrastructure access not available in standard E2E environment | High | HIVE-2544_M001 | ❌ Not Started |

## Scenario 3 - User-Specified Zones in Non-Zoned Region
| Platform | Cluster Type(s) | Test Type  | Reasoning                   | Priority   | Test Case ID | Status         |
|----------|-----------------|------------|-----------------------------|------------|--------------|----------------|
| Azure | Standard | E2E | Edge case validation - graceful degradation when zones specified but unavailable | Medium | HIVE-2544_002 | ❌ Not Started |

## Scenario 4 - Backward Compatibility - Zoned Region Behavior
| Platform | Cluster Type(s) | Test Type  | Reasoning                   | Priority   | Test Case ID | Status         |
|----------|-----------------|------------|-----------------------------|------------|--------------|----------------|
| Azure | Standard | E2E | Regression prevention - ensure zoned regions still work correctly after fix | High | HIVE-2544_003 | ❌ Not Started |
