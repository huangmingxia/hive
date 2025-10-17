# Test Strategy - HIVE-2923

## Test Coverage Matrix

### Scenario Coverage
- **Scenario 1**: DNSZone Status Stability with Invalid Credentials
- **Scenario 2**: Error Message Scrubbing Validation

### Test Type Distribution
- E2E Tests: 1 scenario (DNSZone status stability validation)
- Manual Tests: 1 scenario (Error scrubbing regex verification)

## Test Scenarios

### Scenario 1: DNSZone Status Stability with Invalid AWS Credentials
**Type**: E2E Test
**Platform**: AWS
**Priority**: High

**Objective**: Verify that DNSZone controller does not thrash when encountering AWS errors containing RequestID

**User Workflow**:
1. User creates ClusterDeployment with managed DNS configuration
2. AWS credentials are invalid or have insufficient permissions
3. DNSZone controller encounters AWS API errors with RequestID

**Validation Approach**: Type A - Stability
**Key Measurement**: resourceVersion stability over time window (30 seconds)

**Test Steps**:
1. Create ClusterDeployment with managed DNS using invalid AWS credentials
2. Wait for DNSZone CR to be created
3. Capture initial resourceVersion of DNSZone CR
4. Wait 30 seconds
5. Capture final resourceVersion of DNSZone CR
6. Calculate resourceVersion change

**Expected Results**:
- resourceVersion change should be 0 or minimal (< 5)
- BEFORE fix: resourceVersion change would be 100+ (thrashing)
- AFTER fix: resourceVersion change should be 0 (stable)

**Quantitative Threshold**: resourceVersion change < 5 = PASS

### Scenario 2: Error Message Scrubbing Validation
**Type**: Manual Test
**Platform**: AWS
**Priority**: High

**Objective**: Verify error scrubber correctly handles both "Request ID" and "RequestID" formats

**Test Approach**: Type D - Absence
**Key Measurement**: Absence of RequestID UUID in status condition messages

**Test Cases**:

#### Test Case 2.1: New RequestID Format (No Space)
**Input**: Error containing "RequestID: 032cc7f0-b1a6-4183-bdbb-a15a23a9e029"
**Expected**: RequestID UUID should be scrubbed (not present in status message)

#### Test Case 2.2: Old Request ID Format (With Space)
**Input**: Error containing "Request ID: 42a5a4ce-9c1a-4916-a62a-72a2e6d9ae59"
**Expected**: Request ID UUID should be scrubbed (not present in status message)

#### Test Case 2.3: Multiple RequestIDs in Same Error
**Input**: Error containing both "Request ID: xxx" and "request id: yyy"
**Expected**: Both RequestID UUIDs should be scrubbed

#### Test Case 2.4: Lowercase Variant
**Input**: Error containing "request id: 22bc2e2e-9381-485f-8a46-c7ce8aad2a4d"
**Expected**: request id UUID should be scrubbed

**Validation Method**: Inspect DNSZone status condition message field and verify no UUID patterns remain

## Validation Methods

### Method 1: ResourceVersion Tracking
**Purpose**: Measure controller thrashing behavior
**Tool**: kubectl/oc CLI
**Commands**:
```bash
# Capture initial resourceVersion
INITIAL_RV=$(oc get dnszone <dnszone-name> -o jsonpath='{.metadata.resourceVersion}')

# Wait measurement window
sleep 30

# Capture final resourceVersion
FINAL_RV=$(oc get dnszone <dnszone-name> -o jsonpath='{.metadata.resourceVersion}')

# Calculate change
echo "ResourceVersion change: $((FINAL_RV - INITIAL_RV))"
```

**Success Criteria**: Change < 5

### Method 2: Status Condition Message Inspection
**Purpose**: Verify RequestID scrubbing effectiveness
**Tool**: kubectl/oc CLI
**Commands**:
```bash
# Extract status condition message
oc get dnszone <dnszone-name> -o jsonpath='{.status.conditions[?(@.type=="DNSError")].message}'

# Verify no UUID pattern present
oc get dnszone <dnszone-name> -o jsonpath='{.status.conditions[?(@.type=="DNSError")].message}' | grep -E '[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}'
```

**Success Criteria**: No UUID pattern found (grep returns empty)

### Method 3: Controller Log Analysis
**Purpose**: Verify no excessive reconcile loop logging
**Tool**: kubectl/oc logs
**Commands**:
```bash
# Count DNSZone reconcile log entries in 30 second window
oc logs -n hive deployment/hive-controllers --since=30s | grep "reconciling DNSZone" | wc -l
```

**Success Criteria**: Reconcile count < 10 in 30 seconds
