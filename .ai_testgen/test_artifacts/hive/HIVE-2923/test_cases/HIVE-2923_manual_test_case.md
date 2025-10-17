# Test Case: HIVE-2923
**Component:** hive
**Summary:** Scrub AWS RequestID from DNSZone "DNSError" status condition message (thrashing)

## Test Overview
- **Total Test Cases:** 1
- **Test Types:** Manual
- **Estimated Time:** 15 minutes

## Test Cases

### Test Case HIVE-2923_M001
**Name:** Error Message Scrubbing Validation for Multiple RequestID Formats
**Description:** Manually verify that the error scrubber correctly handles all variations of AWS RequestID format (with space, without space, lowercase, multiple occurrences, embedded newlines). This requires code-level inspection and unit test verification.
**Type:** Manual
**Priority:** High

#### Prerequisites
- Access to Hive source code repository
- Go 1.23.2+ development environment
- Ability to run unit tests locally

#### Test Steps
1. **Action:** Review the error scrubber regex pattern implementation
   ```bash
   # Navigate to error scrubber implementation
   cat pkg/controller/utils/errorscrub.go

   # Verify regex pattern for AWS RequestID
   # Should match: `(, )*(?i)(request ?id: )(?:[-[:xdigit:]]+)`
   # Note the optional space: "request ?id:"
   ```
   **Expected:** Regex pattern includes optional space between "request" and "id", making it case-insensitive and able to handle both old and new AWS formats

2. **Action:** Run error scrubber unit tests
   ```bash
   # Navigate to Hive repository root
   cd /path/to/hive

   # Run error scrubber unit tests
   go test -v ./pkg/controller/utils/... -run TestErrorScrubber
   ```
   **Expected:** All unit tests pass, including new test cases for:
   - "RequestID" format (no space)
   - "Request ID" format (with space)
   - Multiple RequestIDs in same message
   - RequestID with embedded newlines
   - RequestID at end of message

3. **Action:** Verify test case coverage for new RequestID format
   ```bash
   # Review test cases in errorscrub_test.go
   grep -A 3 "suddenly it is spelled RequestID" pkg/controller/utils/errorscrub_test.go
   ```
   **Expected:** Test case exists for new "RequestID" format:
   ```go
   {
     name:     "suddenly it is spelled RequestID",
     input:    errors.New("operation error Resource Groups Tagging API: GetResources, https response error StatusCode: 400, RequestID: 032cc7f0-b1a6-4183-bdbb-a15a23a9e029, api error UnrecognizedClientException: The security token included in the request is invalid."),
     expected: `operation error Resource Groups Tagging API: GetResources, https response error StatusCode: 400, api error UnrecognizedClientException: The security token included in the request is invalid.`,
   }
   ```

4. **Action:** Verify centralized ErrorScrub usage in AWS PrivateLink controller
   ```bash
   # Check that awsprivatelink_controller.go uses ErrorScrub
   grep -n "controllerutils.ErrorScrub" pkg/controller/awsprivatelink/awsprivatelink_controller.go

   # Verify local filterErrorMessage function was removed
   grep -n "filterErrorMessage" pkg/controller/awsprivatelink/awsprivatelink_controller.go
   ```
   **Expected:**
   - Line ~397 shows: `message := controllerutils.ErrorScrub(err)`
   - No local filterErrorMessage function exists (removed for DRY)

5. **Action:** Verify centralized ErrorScrub usage in PrivateLink conditions
   ```bash
   # Check that conditions.go uses ErrorScrub
   grep -n "controllerutils.ErrorScrub" pkg/controller/privatelink/conditions/conditions.go

   # Verify local filterErrorMessage function was removed
   grep -n "filterErrorMessage" pkg/controller/privatelink/conditions/conditions.go
   ```
   **Expected:**
   - Line ~37 shows: `message := controllerutils.ErrorScrub(err)`
   - No local filterErrorMessage function exists (removed for DRY)

6. **Action:** Test regex pattern manually with different RequestID formats
   ```bash
   # Create a test script to verify regex behavior
   cat > /tmp/test_requestid_scrub.sh <<'EOF'
   #!/bin/bash

   TEST_CASES=(
     "RequestID: 032cc7f0-b1a6-4183-bdbb-a15a23a9e029"
     "Request ID: 42a5a4ce-9c1a-4916-a62a-72a2e6d9ae59"
     "request id: 22bc2e2e-9381-485f-8a46-c7ce8aad2a4d"
     "request  id: 9cc3b1f9-e161-402c-a942-d0ed7c7e5fd4"
   )

   echo "Testing RequestID scrubbing patterns:"
   for TEST in "${TEST_CASES[@]}"; do
     # Simulate regex replacement (remove RequestID UUID)
     SCRUBBED=$(echo "$TEST" | sed -E 's/(, )*[Rr]equest ?[Ii][Dd]: [-0-9a-f]+//')
     echo "Original: $TEST"
     echo "Scrubbed: $SCRUBBED"
     echo "---"
   done
   EOF

   chmod +x /tmp/test_requestid_scrub.sh
   /tmp/test_requestid_scrub.sh
   ```
   **Expected:** All test cases show RequestID UUID removed from output

7. **Action:** Verify fix addresses multiple RequestID occurrences
   ```bash
   # Check test case for multiple RequestIDs with different formats
   grep -A 5 "two request IDs, spelled differently" pkg/controller/utils/errorscrub_test.go
   ```
   **Expected:** Test case exists showing both "Request ID" and "request id" are scrubbed:
   ```go
   {
     name:     "two request IDs, spelled differently",
     input:    errors.New(`AccessDenied: Failed to verify the given VPC by calling ec2:DescribeVpcs: You are not authorized to perform this operation. (Service: AmazonEC2; Status Code: 403; Error Code: UnauthorizedOperation; Request ID: 42a5a4ce-9c1a-4916-a62a-72a2e6d9ae59; Proxy: null)\n\tstatus code: 403, request id: 9cc3b1f9-e161-402c-a942-d0ed7c7e5fd4`),
     expected: `AccessDenied: Failed to verify the given VPC by calling ec2:DescribeVpcs: You are not authorized to perform this operation. (Service: AmazonEC2; Status Code: 403; Error Code: UnauthorizedOperation; ; Proxy: null)\n\tstatus code: 403`,
   }
   ```

8. **Action:** Review PR changes to confirm DRY implementation
   ```bash
   # View PR #2744 diff summary
   # Verify these files were modified:
   # - pkg/controller/utils/errorscrub.go (regex updated)
   # - pkg/controller/utils/errorscrub_test.go (new test cases added)
   # - pkg/controller/awsprivatelink/awsprivatelink_controller.go (uses ErrorScrub)
   # - pkg/controller/privatelink/conditions/conditions.go (uses ErrorScrub)
   # - pkg/controller/awsprivatelink/awsprivatelink_controller_test.go (tests removed)
   ```
   **Expected:**
   - Central regex pattern updated in errorscrub.go
   - Local filterErrorMessage functions removed from awsprivatelink and privatelink
   - Unit tests consolidated in errorscrub_test.go
   - No duplicate regex patterns in codebase

#### Validation Criteria
- ✅ **PASS**: All unit tests pass AND regex handles all RequestID formats AND centralized ErrorScrub used in all locations AND no duplicate scrubbing logic exists
- ❌ **FAIL**: Any unit test fails OR regex doesn't handle a format OR duplicate scrubbing logic found

#### Notes
- This manual test verifies code-level changes that cannot be fully validated through E2E tests
- The error scrubber regex pattern is critical for preventing controller thrashing
- The DRY principle ensures consistent error scrubbing across all AWS-related controllers
- Test cases in errorscrub_test.go provide regression protection for future AWS API changes
- The regex pattern `(?i)(request ?id: )` uses:
  - `(?i)` for case-insensitive matching
  - `?` to make the space optional (handles both "Request ID" and "RequestID")
