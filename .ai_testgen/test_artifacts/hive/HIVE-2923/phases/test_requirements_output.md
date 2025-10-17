# Test Requirements Output - HIVE-2923

## Component Name
hive

## Card Summary
DNSZone and ClusterDeployment controllers experiencing thrashing due to AWS RequestID appearing in status condition messages. The AWS SDK v2 transition or AWS API change resulted in RequestID format change from "Request ID" (with space) to "RequestID" (no space), causing the error scrubber regex to fail. This leads to status updates on every reconcile loop, triggering immediate requeues without backoff.

## Test Requirements

### Functional Requirements
1. **Error Scrubber Fix**: Verify that the error scrubber regex correctly handles both "Request ID" (with space) and "RequestID" (without space) formats
2. **Status Condition Stability**: Ensure DNSZone status conditions are not updated when only the AWS RequestID changes
3. **Controller Thrashing Prevention**: Verify that DNSZone and ClusterDeployment controllers do not experience high reconcile rates due to RequestID changes
4. **DRY Implementation**: Confirm that multiple locations using similar regex-based scrubbing now use the centralized `ErrorScrub` function

### Non-Functional Requirements
1. **Performance**: Controller reconcile rate should remain stable even with AWS API errors
2. **Consistency**: All AWS error messages should have RequestIDs scrubbed consistently
3. **Backward Compatibility**: Error scrubbing should handle both old and new RequestID formats

## Affected Platforms
- AWS (All regions)
- DNSZone controller
- ClusterDeployment controller
- AWS PrivateLink controller

## Test Scenarios

### Scenario 1: DNSZone with Invalid AWS Credentials
**Objective**: Verify that DNSZone status condition message does not contain AWS RequestID and resourceVersion remains stable

**Pre-conditions**:
- Cluster with Hive installed
- Invalid AWS credentials configured for DNSZone

**Test Steps**:
1. Create a ClusterDeployment with managed DNS using invalid AWS credentials
2. Monitor the DNSZone CR's status condition messages
3. Track the resourceVersion of the DNSZone CR over 30 seconds

**Expected Results**:
- Status condition message does not contain AWS RequestID (should be scrubbed to "XXXX")
- resourceVersion change should be 0 or minimal (not hundreds)
- No controller thrashing in logs

### Scenario 2: Error Scrubber Regex Pattern Validation
**Objective**: Verify error scrubber handles multiple RequestID formats

**Test Cases**:
1. "Request ID" with space (old format)
2. "RequestID" without space (new format)
3. "request id" lowercase with space
4. Multiple RequestIDs in same error message

**Expected Results**:
- All RequestID variants are correctly scrubbed to "XXXX" or removed
- Error message remains readable and informative

### Scenario 3: AWS PrivateLink Controller Error Handling
**Objective**: Verify AWS PrivateLink controller uses centralized ErrorScrub function

**Pre-conditions**:
- Cluster with PrivateLink configuration
- Invalid AWS credentials

**Test Steps**:
1. Configure AWS PrivateLink with invalid credentials
2. Trigger PrivateLink reconciliation
3. Check status condition messages

**Expected Results**:
- RequestIDs are scrubbed from status condition messages
- No duplicate error scrubbing logic in codebase

## Edge Cases

### Edge Case 1: Mixed RequestID Formats in Single Error Message
**Scenario**: AWS API error contains both "Request ID" and "request id" in different parts
**Expected Behavior**: Both variants should be scrubbed correctly

### Edge Case 2: Rapid AWS API Error Changes
**Scenario**: Multiple AWS API calls failing rapidly with different RequestIDs
**Expected Behavior**: Status condition should not update if only RequestID differs, preventing thrashing

### Edge Case 3: Embedded Newlines in Error Messages
**Scenario**: Error message contains RequestID after newline character
**Expected Behavior**: RequestID should be scrubbed regardless of position or newlines

### Edge Case 4: RequestID at End of Message
**Scenario**: RequestID appears at the very end of error message with no trailing content
**Expected Behavior**: RequestID should be scrubbed without leaving trailing commas or spaces
