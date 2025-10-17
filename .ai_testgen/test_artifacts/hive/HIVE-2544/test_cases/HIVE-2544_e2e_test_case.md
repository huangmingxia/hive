# Test Case: HIVE-2544
**Component:** hive
**Summary:** [Hive/Azure/Azure Gov] Machinepool: Unable to generate machinesets in regions without availability zone support.

## Test Overview
- **Total Test Cases:** 3
- **Test Types:** E2E
- **Estimated Time:** 60-90 minutes

## Test Cases

### Test Case HIVE-2544_001
**Name:** MachinePool Creation in Non-Zoned Azure Region
**Description:** Validate MachinePool and MachineSet generation in Azure regions without availability zone support
**Type:** E2E
**Priority:** High

#### Prerequisites
- Access to Azure cloud credentials with permission to create resources
- Hive operator deployed and running
- Azure region without availability zone support identified (e.g., specific regional locations)
- Base domain configured for cluster DNS

#### Test Steps
1. **Action:** Create ClusterDeployment in Azure region without availability zone support
   ```bash
   # Create namespace
   oc create namespace hive-test-nonzoned-azure

   # Create cloud credentials secret
   oc create secret generic azure-creds \
     --from-literal=osServicePrincipal.json='{"clientId":"xxx","clientSecret":"xxx","tenantId":"xxx","subscriptionId":"xxx"}' \
     -n hive-test-nonzoned-azure

   # Create ClusterDeployment
   cat <<EOF | oc apply -f -
   apiVersion: hive.openshift.io/v1
   kind: ClusterDeployment
   metadata:
     name: azure-nonzoned-cluster
     namespace: hive-test-nonzoned-azure
   spec:
     baseDomain: example.com
     clusterName: azure-nonzoned-cluster
     platform:
       azure:
         baseDomainResourceGroupName: example-rg
         region: francecentral
         credentialsSecretRef:
           name: azure-creds
     provisioning:
       installConfigSecretRef:
         name: install-config
       imageSetRef:
         name: openshift-v4.16.0
     pullSecretRef:
       name: pull-secret
   EOF
   ```
   **Expected:** ClusterDeployment created successfully, cluster installation progresses to Installed=True status within 60 minutes

2. **Action:** Verify cluster installation completed and retrieve cluster region information
   ```bash
   # Wait for installation to complete
   oc wait --for=condition=Installed=true clusterdeployment/azure-nonzoned-cluster \
     -n hive-test-nonzoned-azure --timeout=60m

   # Verify region configuration
   REGION=$(oc get clusterdeployment azure-nonzoned-cluster -n hive-test-nonzoned-azure \
     -o jsonpath='{.spec.platform.azure.region}')
   echo "Cluster deployed to region: $REGION"
   ```
   **Expected:** ClusterDeployment shows Installed=True condition, region matches non-zoned Azure region

3. **Action:** Create MachinePool resource without specifying zones
   ```bash
   cat <<EOF | oc apply -f -
   apiVersion: hive.openshift.io/v1
   kind: MachinePool
   metadata:
     name: worker-nonzoned
     namespace: hive-test-nonzoned-azure
   spec:
     clusterDeploymentRef:
       name: azure-nonzoned-cluster
     name: worker
     platform:
       azure:
         osDisk:
           diskSizeGB: 128
         type: Standard_D4s_v3
     replicas: 3
   EOF
   ```
   **Expected:** MachinePool resource created successfully

4. **Action:** Monitor MachinePool reconciliation and verify Ready status
   ```bash
   # Wait for MachinePool to be ready
   oc wait --for=condition=Ready=true machinepool/worker-nonzoned \
     -n hive-test-nonzoned-azure --timeout=5m

   # Check MachinePool status
   oc get machinepool worker-nonzoned -n hive-test-nonzoned-azure \
     -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}'
   ```
   **Expected:** MachinePool status shows Ready=True condition within 5 minutes, no error conditions present

5. **Action:** Verify MachineSet generation for non-zoned deployment
   ```bash
   # Get kubeconfig for target cluster
   oc get secret azure-nonzoned-cluster-admin-kubeconfig \
     -n hive-test-nonzoned-azure \
     -o jsonpath='{.data.kubeconfig}' | base64 -d > /tmp/azure-nonzoned-kubeconfig

   # Count MachineSets created
   MACHINESET_COUNT=$(oc --kubeconfig=/tmp/azure-nonzoned-kubeconfig \
     get machinesets -n openshift-machine-api \
     --selector hive.openshift.io/machine-pool=worker -o json | \
     oc --kubeconfig=/tmp/azure-nonzoned-kubeconfig get machinesets -n openshift-machine-api \
     --selector hive.openshift.io/machine-pool=worker --no-headers | wc -l)

   echo "MachineSet count: $MACHINESET_COUNT"
   ```
   **Expected:** Exactly 1 MachineSet created (non-zoned deployment)

6. **Action:** Verify MachineSet has no zone configuration
   ```bash
   # Get MachineSet zone configuration
   MACHINESET_NAME=$(oc --kubeconfig=/tmp/azure-nonzoned-kubeconfig \
     get machinesets -n openshift-machine-api \
     --selector hive.openshift.io/machine-pool=worker \
     -o jsonpath='{.items[0].metadata.name}')

   ZONE_CONFIG=$(oc --kubeconfig=/tmp/azure-nonzoned-kubeconfig \
     get machineset $MACHINESET_NAME -n openshift-machine-api \
     -o jsonpath='{.spec.template.spec.providerSpec.value.zone}')

   echo "Zone configuration: '$ZONE_CONFIG'"
   ```
   **Expected:** Zone field is empty or absent (non-zoned configuration)

7. **Action:** Verify MachineSet replica count matches MachinePool specification
   ```bash
   # Get MachineSet replicas
   MACHINESET_REPLICAS=$(oc --kubeconfig=/tmp/azure-nonzoned-kubeconfig \
     get machineset $MACHINESET_NAME -n openshift-machine-api \
     -o jsonpath='{.spec.replicas}')

   echo "MachineSet replicas: $MACHINESET_REPLICAS"
   ```
   **Expected:** MachineSet replicas equals 3 (matches MachinePool spec)

8. **Action:** Verify controller logs contain non-zoned deployment message
   ```bash
   # Get hive-controller pod logs
   HIVE_POD=$(oc get pods -n hive -l control-plane=controller-manager \
     -o jsonpath='{.items[0].metadata.name}')

   # Search for non-zoned deployment log message
   LOG_COUNT=$(oc logs $HIVE_POD -n hive | \
     grep -c "No availability zones detected for region. Using non-zoned deployment.")

   echo "Non-zoned deployment log entries: $LOG_COUNT"
   ```
   **Expected:** At least 1 log entry indicating non-zoned deployment detected

9. **Action:** Verify Machine resources are created and become ready
   ```bash
   # Wait for Machines to be created
   sleep 120

   # Count Machine resources
   MACHINE_COUNT=$(oc --kubeconfig=/tmp/azure-nonzoned-kubeconfig \
     get machines -n openshift-machine-api \
     --selector machine.openshift.io/cluster-api-machineset=$MACHINESET_NAME \
     --no-headers | wc -l)

   echo "Machine count: $MACHINE_COUNT"

   # Check Machine phase
   oc --kubeconfig=/tmp/azure-nonzoned-kubeconfig \
     get machines -n openshift-machine-api \
     --selector machine.openshift.io/cluster-api-machineset=$MACHINESET_NAME \
     -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.status.phase}{"\n"}{end}'
   ```
   **Expected:** 3 Machine resources created, eventually reaching Running phase within 10 minutes

10. **Action:** Verify no error messages related to zone detection
    ```bash
    # Search for error messages in controller logs
    ERROR_COUNT=$(oc logs $HIVE_POD -n hive | \
      grep -c "zero zones returned for region")

    echo "Error message count: $ERROR_COUNT"
    ```
    **Expected:** Zero occurrences of "zero zones returned" error message

### Test Case HIVE-2544_002
**Name:** User-Specified Zones in Non-Zoned Azure Region
**Description:** Validate graceful degradation when user specifies zones but region doesn't support them
**Type:** E2E
**Priority:** Medium

#### Prerequisites
- Access to Azure cloud credentials
- Hive operator deployed and running
- ClusterDeployment already installed in non-zoned Azure region (reuse from HIVE-2544_001)

#### Test Steps
1. **Action:** Create MachinePool with explicit zone specification in non-zoned region
   ```bash
   cat <<EOF | oc apply -f -
   apiVersion: hive.openshift.io/v1
   kind: MachinePool
   metadata:
     name: worker-explicit-zones
     namespace: hive-test-nonzoned-azure
   spec:
     clusterDeploymentRef:
       name: azure-nonzoned-cluster
     name: worker-zones
     platform:
       azure:
         osDisk:
           diskSizeGB: 128
         type: Standard_D4s_v3
         zones:
         - "1"
         - "2"
         - "3"
     replicas: 3
   EOF
   ```
   **Expected:** MachinePool resource created successfully

2. **Action:** Verify MachinePool reconciliation succeeds despite zone mismatch
   ```bash
   # Wait for MachinePool ready status
   oc wait --for=condition=Ready=true machinepool/worker-explicit-zones \
     -n hive-test-nonzoned-azure --timeout=5m

   # Check for error conditions
   oc get machinepool worker-explicit-zones -n hive-test-nonzoned-azure \
     -o jsonpath='{.status.conditions[?(@.type=="Ready")]}'
   ```
   **Expected:** MachinePool reaches Ready=True status, no error conditions blocking deployment

3. **Action:** Verify warning log generated for zone override
   ```bash
   # Get controller logs and check for warning
   HIVE_POD=$(oc get pods -n hive -l control-plane=controller-manager \
     -o jsonpath='{.items[0].metadata.name}')

   # Look for zone-related warning messages
   oc logs $HIVE_POD -n hive --tail=200 | \
     grep -i "zone" | grep -i "worker-explicit-zones"
   ```
   **Expected:** Warning message logged indicating zones specified but region doesn't support them

4. **Action:** Verify MachineSet created without zone configuration (graceful degradation)
   ```bash
   # Get MachineSet details
   MACHINESET_NAME=$(oc --kubeconfig=/tmp/azure-nonzoned-kubeconfig \
     get machinesets -n openshift-machine-api \
     --selector hive.openshift.io/machine-pool=worker-zones \
     -o jsonpath='{.items[0].metadata.name}')

   # Check zone configuration
   ZONE_CONFIG=$(oc --kubeconfig=/tmp/azure-nonzoned-kubeconfig \
     get machineset $MACHINESET_NAME -n openshift-machine-api \
     -o jsonpath='{.spec.template.spec.providerSpec.value.zone}')

   echo "Zone configuration after graceful degradation: '$ZONE_CONFIG'"

   # Count MachineSets
   MACHINESET_COUNT=$(oc --kubeconfig=/tmp/azure-nonzoned-kubeconfig \
     get machinesets -n openshift-machine-api \
     --selector hive.openshift.io/machine-pool=worker-zones --no-headers | wc -l)

   echo "MachineSet count: $MACHINESET_COUNT"
   ```
   **Expected:** MachineSet has no zone configuration, only 1 MachineSet created (not 3), graceful degradation to non-zoned deployment successful

5. **Action:** Verify deployment completes successfully
   ```bash
   # Verify Machine creation
   sleep 120

   MACHINE_COUNT=$(oc --kubeconfig=/tmp/azure-nonzoned-kubeconfig \
     get machines -n openshift-machine-api \
     --selector machine.openshift.io/cluster-api-machineset=$MACHINESET_NAME \
     --no-headers | wc -l)

   echo "Machine count: $MACHINE_COUNT"
   ```
   **Expected:** 3 Machine resources created successfully despite zone specification mismatch

### Test Case HIVE-2544_003
**Name:** Backward Compatibility - Zoned Region Behavior
**Description:** Ensure fix doesn't break existing functionality for Azure regions with availability zone support
**Type:** E2E
**Priority:** High

#### Prerequisites
- Access to Azure cloud credentials
- Hive operator deployed and running
- Azure region with availability zone support (e.g., eastus, westus2)

#### Test Steps
1. **Action:** Create ClusterDeployment in Azure region with availability zone support
   ```bash
   # Create namespace
   oc create namespace hive-test-zoned-azure

   # Create cloud credentials secret
   oc create secret generic azure-creds-zoned \
     --from-literal=osServicePrincipal.json='{"clientId":"xxx","clientSecret":"xxx","tenantId":"xxx","subscriptionId":"xxx"}' \
     -n hive-test-zoned-azure

   # Create ClusterDeployment
   cat <<EOF | oc apply -f -
   apiVersion: hive.openshift.io/v1
   kind: ClusterDeployment
   metadata:
     name: azure-zoned-cluster
     namespace: hive-test-zoned-azure
   spec:
     baseDomain: example.com
     clusterName: azure-zoned-cluster
     platform:
       azure:
         baseDomainResourceGroupName: example-rg
         region: eastus
         credentialsSecretRef:
           name: azure-creds-zoned
     provisioning:
       installConfigSecretRef:
         name: install-config
       imageSetRef:
         name: openshift-v4.16.0
     pullSecretRef:
       name: pull-secret
   EOF
   ```
   **Expected:** ClusterDeployment created successfully in zoned region

2. **Action:** Wait for cluster installation to complete
   ```bash
   # Wait for installation
   oc wait --for=condition=Installed=true clusterdeployment/azure-zoned-cluster \
     -n hive-test-zoned-azure --timeout=60m

   # Verify region
   REGION=$(oc get clusterdeployment azure-zoned-cluster -n hive-test-zoned-azure \
     -o jsonpath='{.spec.platform.azure.region}')
   echo "Cluster deployed to region: $REGION"
   ```
   **Expected:** Cluster installs successfully in zoned Azure region (eastus)

3. **Action:** Create MachinePool without specifying zones (auto-detection test)
   ```bash
   cat <<EOF | oc apply -f -
   apiVersion: hive.openshift.io/v1
   kind: MachinePool
   metadata:
     name: worker-autozone
     namespace: hive-test-zoned-azure
   spec:
     clusterDeploymentRef:
       name: azure-zoned-cluster
     name: worker
     platform:
       azure:
         osDisk:
           diskSizeGB: 128
         type: Standard_D4s_v3
     replicas: 3
   EOF
   ```
   **Expected:** MachinePool created successfully

4. **Action:** Verify automatic zone detection and population
   ```bash
   # Wait for MachinePool ready
   oc wait --for=condition=Ready=true machinepool/worker-autozone \
     -n hive-test-zoned-azure --timeout=5m

   # Get kubeconfig
   oc get secret azure-zoned-cluster-admin-kubeconfig \
     -n hive-test-zoned-azure \
     -o jsonpath='{.data.kubeconfig}' | base64 -d > /tmp/azure-zoned-kubeconfig
   ```
   **Expected:** MachinePool becomes Ready within 5 minutes

5. **Action:** Verify multiple MachineSets created (one per zone)
   ```bash
   # Count MachineSets
   MACHINESET_COUNT=$(oc --kubeconfig=/tmp/azure-zoned-kubeconfig \
     get machinesets -n openshift-machine-api \
     --selector hive.openshift.io/machine-pool=worker --no-headers | wc -l)

   echo "MachineSet count in zoned region: $MACHINESET_COUNT"
   ```
   **Expected:** 3 MachineSets created (one per availability zone)

6. **Action:** Verify each MachineSet has zone configuration
   ```bash
   # Get zone configuration for each MachineSet
   oc --kubeconfig=/tmp/azure-zoned-kubeconfig \
     get machinesets -n openshift-machine-api \
     --selector hive.openshift.io/machine-pool=worker \
     -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.spec.template.spec.providerSpec.value.zone}{"\n"}{end}'
   ```
   **Expected:** Each MachineSet has exactly one zone assigned (e.g., 1, 2, 3)

7. **Action:** Verify replica distribution across zones
   ```bash
   # Check replica count per MachineSet
   oc --kubeconfig=/tmp/azure-zoned-kubeconfig \
     get machinesets -n openshift-machine-api \
     --selector hive.openshift.io/machine-pool=worker \
     -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.spec.replicas}{"\n"}{end}'

   # Calculate total replicas
   TOTAL_REPLICAS=$(oc --kubeconfig=/tmp/azure-zoned-kubeconfig \
     get machinesets -n openshift-machine-api \
     --selector hive.openshift.io/machine-pool=worker \
     -o jsonpath='{.items[*].spec.replicas}' | awk '{for(i=1;i<=NF;i++) sum+=$i; print sum}')

   echo "Total replicas across MachineSets: $TOTAL_REPLICAS"
   ```
   **Expected:** Total replicas equals 3 (MachinePool spec), balanced distribution (each MachineSet has 1 replica, or difference ≤1)

8. **Action:** Verify Machine resources provisioned across zones
   ```bash
   # Wait for Machines
   sleep 120

   # Check Machine distribution
   oc --kubeconfig=/tmp/azure-zoned-kubeconfig \
     get machines -n openshift-machine-api \
     -l machine.openshift.io/cluster-api-machine-type=worker \
     -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.spec.providerSpec.value.zone}{"\n"}{end}'

   # Count total Machines
   MACHINE_COUNT=$(oc --kubeconfig=/tmp/azure-zoned-kubeconfig \
     get machines -n openshift-machine-api \
     -l machine.openshift.io/cluster-api-machine-type=worker --no-headers | wc -l)

   echo "Total Machine count: $MACHINE_COUNT"
   ```
   **Expected:** 3 Machine resources created, distributed across zones (each in different zone)

9. **Action:** Verify no regression - check for non-zoned deployment log message
   ```bash
   # Verify non-zoned message NOT present for zoned region
   HIVE_POD=$(oc get pods -n hive -l control-plane=controller-manager \
     -o jsonpath='{.items[0].metadata.name}')

   # Search recent logs for worker-autozone MachinePool
   oc logs $HIVE_POD -n hive --tail=500 | \
     grep "worker-autozone" | grep -c "Using non-zoned deployment" || echo "0"
   ```
   **Expected:** No "Using non-zoned deployment" message for zoned region MachinePool

10. **Action:** Verify zone detection worked correctly
    ```bash
    # Check controller logs for zone detection
    oc logs $HIVE_POD -n hive --tail=500 | \
      grep "worker-autozone" | grep -i "zone"
    ```
    **Expected:** Logs indicate zones were detected and applied, no errors related to zone handling
