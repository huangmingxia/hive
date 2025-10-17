# Test Case: HIVE-2923
**Component:** hive
**Summary:** Scrub AWS RequestID from DNSZone "DNSError" status condition message (thrashing)

## Test Overview
- **Total Test Cases:** 1
- **Test Types:** E2E
- **Estimated Time:** 10 minutes

## Test Cases

### Test Case HIVE-2923_001
**Name:** DNSZone Status Stability with Invalid AWS Credentials
**Description:** Verify that DNSZone controller does not thrash when encountering AWS API errors containing RequestID. The test ensures that RequestID values are properly scrubbed from status condition messages, preventing unnecessary status updates that trigger immediate requeues.
**Type:** E2E
**Priority:** High

#### Prerequisites
- OpenShift cluster with Hive operator installed
- AWS credentials with permissions to create ClusterDeployment
- kubectl/oc CLI configured
- Test AWS credentials that are invalid or lack DNS permissions

#### Test Steps
1. **Action:** Configure invalid AWS credentials in a Secret
   ```bash
   # Create secret with invalid AWS credentials
   cat <<EOF | oc apply -f -
   apiVersion: v1
   kind: Secret
   metadata:
     name: invalid-aws-creds
     namespace: hive
   type: Opaque
   stringData:
     aws_access_key_id: "INVALID_ACCESS_KEY"
     aws_secret_access_key: "INVALID_SECRET_KEY"
   EOF
   ```
   **Expected:** Secret created successfully

2. **Action:** Create ClusterDeployment with managed DNS using invalid credentials
   ```bash
   cat <<EOF | oc apply -f -
   apiVersion: hive.openshift.io/v1
   kind: ClusterDeployment
   metadata:
     name: test-dnszone-thrash
     namespace: hive
   spec:
     baseDomain: test.example.com
     clusterName: test-dnszone-thrash
     platform:
       aws:
         credentialsSecretRef:
           name: invalid-aws-creds
         region: us-east-1
     managedDNS:
       route53:
         credentialsSecretRef:
           name: invalid-aws-creds
     provisioning:
       installConfigSecretRef:
         name: test-install-config
   EOF
   ```
   **Expected:** ClusterDeployment created successfully

3. **Action:** Wait for DNSZone to be created and enter error state
   ```bash
   # Wait for DNSZone creation (max 2 minutes)
   timeout 120 bash -c 'until oc get dnszone -n hive | grep test-dnszone-thrash-zone; do sleep 5; done'

   # Verify DNSZone exists
   oc get dnszone -n hive -l hive.openshift.io/cluster-deployment-name=test-dnszone-thrash
   ```
   **Expected:** DNSZone CR created with name matching pattern "test-dnszone-thrash-*-zone"

4. **Action:** Capture initial resourceVersion and wait 30 seconds
   ```bash
   # Get DNSZone name
   DNSZONE_NAME=$(oc get dnszone -n hive -l hive.openshift.io/cluster-deployment-name=test-dnszone-thrash -o jsonpath='{.items[0].metadata.name}')

   # Capture initial resourceVersion
   INITIAL_RV=$(oc get dnszone -n hive ${DNSZONE_NAME} -o jsonpath='{.metadata.resourceVersion}')
   echo "Initial resourceVersion: ${INITIAL_RV}"

   # Wait 30 seconds
   sleep 30
   ```
   **Expected:** Initial resourceVersion captured successfully

5. **Action:** Capture final resourceVersion and calculate change
   ```bash
   # Capture final resourceVersion
   FINAL_RV=$(oc get dnszone -n hive ${DNSZONE_NAME} -o jsonpath='{.metadata.resourceVersion}')
   echo "Final resourceVersion: ${FINAL_RV}"

   # Calculate change
   RV_CHANGE=$((FINAL_RV - INITIAL_RV))
   echo "ResourceVersion change: ${RV_CHANGE}"
   ```
   **Expected:** resourceVersion change should be less than 5 (indicating stable status, no thrashing)

6. **Action:** Verify DNSZone status condition message does not contain RequestID UUID
   ```bash
   # Extract DNSError condition message
   ERROR_MSG=$(oc get dnszone -n hive ${DNSZONE_NAME} -o jsonpath='{.status.conditions[?(@.type=="DNSError")].message}')
   echo "Error message: ${ERROR_MSG}"

   # Check for UUID pattern (should not find any)
   UUID_COUNT=$(echo "${ERROR_MSG}" | grep -oE '[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}' | wc -l)
   echo "UUID count in error message: ${UUID_COUNT}"
   ```
   **Expected:** UUID_COUNT should be 0 (no RequestID UUIDs present in error message)

7. **Action:** Verify controller is not thrashing by checking reconcile log frequency
   ```bash
   # Count DNSZone reconcile logs in last 30 seconds
   POD_NAME=$(oc get pods -n hive -l control-plane=controller-manager --field-selector=status.phase=Running -o jsonpath='{.items[0].metadata.name}')
   RECONCILE_COUNT=$(oc logs -n hive ${POD_NAME} --since=30s | grep -c "reconciling DNSZone.*${DNSZONE_NAME}" || echo 0)
   echo "Reconcile count in 30 seconds: ${RECONCILE_COUNT}"
   ```
   **Expected:** RECONCILE_COUNT should be less than 10 (no excessive reconciliation)

8. **Action:** Cleanup test resources
   ```bash
   # Delete ClusterDeployment (will cascade delete DNSZone)
   oc delete clusterdeployment -n hive test-dnszone-thrash

   # Delete invalid credentials secret
   oc delete secret -n hive invalid-aws-creds
   ```
   **Expected:** All test resources cleaned up successfully

#### Validation Criteria
- ✅ **PASS**: resourceVersion change < 5 AND UUID count = 0 AND reconcile count < 10
- ❌ **FAIL**: resourceVersion change >= 5 OR UUID count > 0 OR reconcile count >= 10

#### Notes
- This test validates the fix for AWS SDK v2 transition where RequestID format changed from "Request ID" (with space) to "RequestID" (without space)
- Before the fix, resourceVersion would change 100+ times in 30 seconds due to thrashing
- After the fix, resourceVersion should remain stable (0-2 changes at most)
- The error scrubber regex now handles both "Request ID" and "RequestID" formats
