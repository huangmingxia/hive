# Test Cases: AWS RequestID Scrubbing in DNSZone Error Messages

## Test Information
- **JIRA Issue**: [HIVE-2923](https://issues.redhat.com/browse/HIVE-2923)
- **Title**: Scrub AWS RequestID from DNSZone "DNSError" status condition message (thrashing)
- **Component**: Hive DNSZone Controller
- **Priority**: High
- **Status**: Draft
- **Type**: E2E

## Test Case 1: DNSZone Error Message Stability with Invalid AWS Credentials

**Test ID**: HIVE-2923_01
**Objective**: Verify AWS RequestID is properly scrubbed from DNSZone error messages and status conditions remain stable across multiple reconciles
**Validation Pattern**: Stability Validation (Pattern A)

### Prerequisites
- Hive operator installed and running on AWS cluster
- Access to create ClusterDeployment resources with managedDNS enabled
- Ability to configure AWS credentials for testing

### Test Steps

#### Step 1: Configure HiveConfig with managedDomains

**Action:**
Use hiveutil to configure managedDomains in HiveConfig to enable automatic DNSZone creation.

```bash
# Configure managed domains in HiveConfig
cat <<EOF | oc apply -f -
apiVersion: hive.openshift.io/v1
kind: HiveConfig
metadata:
  name: hive
spec:
  managedDomains:
  - domains:
    - qe.devcluster.openshift.com
    aws:
      credentialsSecretRef:
        name: invalid-aws-credentials
EOF
```

**Expected Result:**
HiveConfig is successfully updated with managedDomains configuration. Verify using:
```bash
oc get hiveconfig hive -o jsonpath='{.spec.managedDomains[0].domains[0]}'
# Expected output: qe.devcluster.openshift.com
```

#### Step 2: Create AWS credentials secret with invalid credentials

**Action:**
Create a secret containing invalid AWS credentials to trigger authentication errors.

```bash
# Create secret with invalid AWS credentials
cat <<EOF | oc apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: invalid-aws-credentials
  namespace: hive
type: Opaque
stringData:
  aws_access_key_id: INVALID_ACCESS_KEY
  aws_secret_access_key: INVALID_SECRET_KEY
EOF
```

**Expected Result:**
Secret is created successfully in hive namespace. Verify using:
```bash
oc get secret invalid-aws-credentials -n hive -o jsonpath='{.data.aws_access_key_id}' | base64 -d
# Expected output: INVALID_ACCESS_KEY
```

#### Step 3: Create ClusterDeployment with managedDNS enabled

**Action:**
Create a ClusterDeployment with managedDNS=true to trigger automatic DNSZone creation.

```bash
# Create ClusterDeployment with managedDNS
cat <<EOF | oc apply -f -
apiVersion: hive.openshift.io/v1
kind: ClusterDeployment
metadata:
  name: test-dnszone-2923
  namespace: hive
spec:
  baseDomain: qe.devcluster.openshift.com
  clusterName: test-dnszone-2923
  platform:
    aws:
      credentialsSecretRef:
        name: invalid-aws-credentials
      region: us-east-1
  provisioning:
    installConfigSecretRef:
      name: test-install-config
    imageSetRef:
      name: test-imageset
  managedDNS: true
EOF
```

**Expected Result:**
ClusterDeployment is created and DNSZone is automatically created. Verify DNSZone creation:
```bash
oc get dnszone -n hive | grep test-dnszone-2923
# Expected: DNSZone CR exists with matching name
```

#### Step 4: Wait for DNSZone controller to encounter AWS error

**Action:**
Wait for DNSZone controller to attempt AWS Route53 operations and encounter authentication errors.

```bash
# Wait for DNSZone to have error condition
oc wait --for=condition=DNSError dnszone/test-dnszone-2923-qe-devcluster-openshift-com -n hive --timeout=120s

# Verify DNSError condition exists
oc get dnszone test-dnszone-2923-qe-devcluster-openshift-com -n hive -o jsonpath='{.status.conditions[?(@.type=="DNSError")].status}'
```

**Expected Result:**
DNSZone status contains DNSError condition with status=True, indicating AWS authentication failure has been encountered and recorded.

#### Step 5: Capture initial DNSError message (T1)

**Action:**
Capture the DNSError status condition message at time T1.

```bash
# Capture initial error message
ERROR_MSG_T1=$(oc get dnszone test-dnszone-2923-qe-devcluster-openshift-com -n hive -o jsonpath='{.status.conditions[?(@.type=="DNSError")].message}')
echo "T1 Message: $ERROR_MSG_T1"

# Capture initial resourceVersion
RV_T1=$(oc get dnszone test-dnszone-2923-qe-devcluster-openshift-com -n hive -o jsonpath='{.metadata.resourceVersion}')
echo "T1 ResourceVersion: $RV_T1"
```

**Expected Result:**
Error message is captured successfully. Message should contain AWS authentication error but NO UUID patterns (RequestID should be scrubbed). Example expected format:
```
operation error Resource Groups Tagging API: GetResources, https response error StatusCode: 400, api error UnrecognizedClientException: The security token included in the request is invalid.
```

#### Step 6: Wait for multiple reconcile cycles

**Action:**
Wait 60 seconds to allow DNSZone controller to perform multiple reconcile attempts.

```bash
# Wait for multiple reconcile cycles
sleep 60
```

**Expected Result:**
DNSZone controller continues to reconcile with invalid credentials, triggering multiple AWS API calls with different RequestIDs.

#### Step 7: Capture subsequent DNSError message (T2)

**Action:**
Capture the DNSError status condition message at time T2 after multiple reconciles.

```bash
# Capture subsequent error message
ERROR_MSG_T2=$(oc get dnszone test-dnszone-2923-qe-devcluster-openshift-com -n hive -o jsonpath='{.status.conditions[?(@.type=="DNSError")].message}')
echo "T2 Message: $ERROR_MSG_T2"

# Capture subsequent resourceVersion
RV_T2=$(oc get dnszone test-dnszone-2923-qe-devcluster-openshift-com -n hive -o jsonpath='{.metadata.resourceVersion}')
echo "T2 ResourceVersion: $RV_T2"
```

**Expected Result:**
Error message at T2 is captured successfully. ResourceVersion should be stable (not changing every second, indicating proper backoff behavior).

#### Step 8: Verify message stability (Quantitative Validation)

**Action:**
Compare error messages from T1 and T2 to verify they are identical.

```bash
# Compare messages
if [ "$ERROR_MSG_T1" = "$ERROR_MSG_T2" ]; then
    echo "✅ PASS: Error messages are identical (stable across reconciles)"
else
    echo "❌ FAIL: Error messages differ between T1 and T2"
    echo "Difference indicates RequestID was not scrubbed properly"
    diff <(echo "$ERROR_MSG_T1") <(echo "$ERROR_MSG_T2")
fi
```

**Expected Result:**
Messages must be byte-identical. SUCCESS CRITERIA:
- ERROR_MSG_T1 == ERROR_MSG_T2 (zero byte difference)
- This proves RequestID is properly scrubbed and not causing status condition changes

#### Step 9: Verify RequestID absence (Absence Validation)

**Action:**
Search error message for UUID patterns to confirm RequestID is not present.

```bash
# Search for UUID patterns (RequestID format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx)
UUID_COUNT=$(echo "$ERROR_MSG_T2" | grep -oE '[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}' | wc -l)

if [ "$UUID_COUNT" -eq 0 ]; then
    echo "✅ PASS: No UUID patterns found (RequestID properly scrubbed)"
else
    echo "❌ FAIL: Found $UUID_COUNT UUID patterns in error message"
    echo "$ERROR_MSG_T2" | grep -oE '[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}'
fi
```

**Expected Result:**
UUID_COUNT must equal 0. SUCCESS CRITERIA:
- Zero UUID patterns found in error message
- Confirms RequestID has been replaced with placeholder or removed

#### Step 10: Verify reconcile frequency is stable (Frequency Validation)

**Action:**
Monitor DNSZone resourceVersion changes over 5-minute window to verify proper backoff behavior.

```bash
# Monitor resourceVersion changes for 5 minutes
RV_START=$(oc get dnszone test-dnszone-2923-qe-devcluster-openshift-com -n hive -o jsonpath='{.metadata.resourceVersion}')
START_TIME=$(date +%s)

sleep 300  # Wait 5 minutes

RV_END=$(oc get dnszone test-dnszone-2923-qe-devcluster-openshift-com -n hive -o jsonpath='{.metadata.resourceVersion}')
END_TIME=$(date +%s)

# Calculate change rate
RV_CHANGES=$((RV_END - RV_START))
TIME_ELAPSED=$((END_TIME - START_TIME))
CHANGES_PER_MINUTE=$(echo "scale=2; $RV_CHANGES * 60 / $TIME_ELAPSED" | bc)

echo "ResourceVersion changes: $RV_CHANGES over $TIME_ELAPSED seconds"
echo "Change rate: $CHANGES_PER_MINUTE changes per minute"

# Verify reconcile rate is reasonable (not thrashing)
if (( $(echo "$CHANGES_PER_MINUTE < 5" | bc -l) )); then
    echo "✅ PASS: Reconcile frequency is stable (< 5 changes/min, proper backoff)"
else
    echo "❌ FAIL: Excessive reconcile frequency ($CHANGES_PER_MINUTE changes/min indicates thrashing)"
fi
```

**Expected Result:**
Reconcile rate should be less than 5 changes per minute. SUCCESS CRITERIA:
- CHANGES_PER_MINUTE < 5 (indicates proper exponential backoff, not immediate requeues)
- Confirms DNSZone CR is not being updated every reconcile due to message changes

### Cleanup

**Action:**
Delete test resources after test completion.

```bash
# Delete ClusterDeployment (cascades to DNSZone)
oc delete clusterdeployment test-dnszone-2923 -n hive

# Delete invalid credentials secret
oc delete secret invalid-aws-credentials -n hive

# Restore HiveConfig if needed
oc patch hiveconfig hive --type='json' -p='[{"op": "remove", "path": "/spec/managedDomains"}]'
```

**Expected Result:**
All test resources are deleted successfully. Verify cleanup:
```bash
oc get clusterdeployment test-dnszone-2923 -n hive 2>&1 | grep "NotFound"
oc get dnszone -n hive | grep test-dnszone-2923 | wc -l  # Should return 0
```

---
