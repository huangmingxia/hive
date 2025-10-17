# Test Case: HIVE-2544
**Component:** hive
**Summary:** [Hive/Azure/Azure Gov] Machinepool: Unable to generate machinesets in regions without availability zone support.

## Test Overview
- **Total Test Cases:** 1
- **Test Types:** Manual
- **Estimated Time:** 90-120 minutes

## Test Cases

### Test Case HIVE-2544_M001
**Name:** MachinePool Creation in Non-Zoned Azure Government Region
**Description:** Verify MachinePool functionality in Azure Government cloud regions without availability zone support (requires special infrastructure access)
**Type:** Manual
**Priority:** High

#### Prerequisites
- Access to Azure Government cloud subscription with active credentials
- Hive operator deployed on management cluster with Azure Gov connectivity
- Azure Government region without availability zone support identified (e.g., usgovtexas, usgovarizona, usdodeast, usdodcentral)
- Base domain configured for Azure Gov cluster DNS
- Pull secret for OpenShift installation
- SSH key for cluster access

#### Test Steps
1. **Action:** Prepare Azure Government cloud credentials
   ```bash
   # Create namespace for test
   oc create namespace hive-test-azuregov

   # Create Azure Gov credentials secret
   # Note: Azure Gov uses different endpoints (usgovcloudapi.net)
   cat <<EOF | oc apply -f -
   apiVersion: v1
   kind: Secret
   metadata:
     name: azure-gov-creds
     namespace: hive-test-azuregov
   type: Opaque
   stringData:
     osServicePrincipal.json: |
       {
         "clientId": "<azure-gov-client-id>",
         "clientSecret": "<azure-gov-client-secret>",
         "tenantId": "<azure-gov-tenant-id>",
         "subscriptionId": "<azure-gov-subscription-id>"
       }
   EOF

   # Create pull secret
   oc create secret generic pull-secret \
     --from-file=.dockerconfigjson=/path/to/pull-secret \
     --type=kubernetes.io/dockerconfigjson \
     -n hive-test-azuregov

   # Create SSH key secret
   oc create secret generic ssh-key \
     --from-file=ssh-publickey=/path/to/ssh-public-key \
     -n hive-test-azuregov
   ```
   **Expected:** Credentials and secrets created successfully in namespace

2. **Action:** Create install-config secret for Azure Government region without zones
   ```bash
   # Create install-config for Azure Gov (usgovtexas - no zones)
   cat <<EOF | oc apply -f -
   apiVersion: v1
   kind: Secret
   metadata:
     name: install-config
     namespace: hive-test-azuregov
   type: Opaque
   stringData:
     install-config.yaml: |
       apiVersion: v1
       baseDomain: example.gov
       metadata:
         name: azgov-nonzoned
       platform:
         azure:
           baseDomainResourceGroupName: azgov-dns-rg
           region: usgovtexas
           cloudName: AzureUSGovernmentCloud
       pullSecret: ""
       sshKey: ""
   EOF
   ```
   **Expected:** Install-config secret created with Azure Gov cloud configuration

3. **Action:** Create ClusterImageSet for OpenShift version
   ```bash
   cat <<EOF | oc apply -f -
   apiVersion: hive.openshift.io/v1
   kind: ClusterImageSet
   metadata:
     name: openshift-v4.16.0
   spec:
     releaseImage: quay.io/openshift-release-dev/ocp-release:4.16.0-x86_64
   EOF
   ```
   **Expected:** ClusterImageSet created successfully

4. **Action:** Create ClusterDeployment for Azure Government region (usgovtexas)
   ```bash
   cat <<EOF | oc apply -f -
   apiVersion: hive.openshift.io/v1
   kind: ClusterDeployment
   metadata:
     name: azgov-nonzoned-cluster
     namespace: hive-test-azuregov
   spec:
     baseDomain: example.gov
     clusterName: azgov-nonzoned
     platform:
       azure:
         baseDomainResourceGroupName: azgov-dns-rg
         region: usgovtexas
         cloudName: AzureUSGovernmentCloud
         credentialsSecretRef:
           name: azure-gov-creds
     provisioning:
       installConfigSecretRef:
         name: install-config
       imageSetRef:
         name: openshift-v4.16.0
       sshPrivateKeySecretRef:
         name: ssh-key
     pullSecretRef:
       name: pull-secret
   EOF
   ```
   **Expected:** ClusterDeployment resource created successfully

5. **Action:** Monitor ClusterDeployment provisioning progress
   ```bash
   # Watch ClusterDeployment status
   oc get clusterdeployment azgov-nonzoned-cluster -n hive-test-azuregov -w

   # Check provision pod logs
   PROVISION_POD=$(oc get pods -n hive-test-azuregov \
     -l hive.openshift.io/cluster-deployment-name=azgov-nonzoned-cluster \
     -l job-name -o jsonpath='{.items[0].metadata.name}')

   oc logs $PROVISION_POD -n hive-test-azuregov -f

   # Wait for installation to complete (typically 45-60 minutes)
   oc wait --for=condition=Installed=true \
     clusterdeployment/azgov-nonzoned-cluster \
     -n hive-test-azuregov --timeout=75m
   ```
   **Expected:** ClusterDeployment progresses through provisioning phases, eventually reaching Installed=True condition within 60-75 minutes

6. **Action:** Verify cluster deployed to Azure Government region without zones
   ```bash
   # Check ClusterDeployment details
   oc get clusterdeployment azgov-nonzoned-cluster -n hive-test-azuregov \
     -o jsonpath='{.spec.platform.azure.region}'
   echo ""

   oc get clusterdeployment azgov-nonzoned-cluster -n hive-test-azuregov \
     -o jsonpath='{.spec.platform.azure.cloudName}'
   echo ""

   # Verify installation status
   oc get clusterdeployment azgov-nonzoned-cluster -n hive-test-azuregov \
     -o jsonpath='{.status.conditions[?(@.type=="Installed")]}'
   ```
   **Expected:** Region shows usgovtexas, cloudName shows AzureUSGovernmentCloud, Installed condition is True

7. **Action:** Create MachinePool in Azure Gov non-zoned region
   ```bash
   cat <<EOF | oc apply -f -
   apiVersion: hive.openshift.io/v1
   kind: MachinePool
   metadata:
     name: worker-azgov
     namespace: hive-test-azuregov
   spec:
     clusterDeploymentRef:
       name: azgov-nonzoned-cluster
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

8. **Action:** Monitor MachinePool reconciliation and verify Ready status
   ```bash
   # Watch MachinePool status
   oc get machinepool worker-azgov -n hive-test-azuregov -w

   # Wait for Ready condition
   oc wait --for=condition=Ready=true machinepool/worker-azgov \
     -n hive-test-azuregov --timeout=5m

   # Check status conditions
   oc get machinepool worker-azgov -n hive-test-azuregov \
     -o jsonpath='{.status.conditions[*]}' | jq
   ```
   **Expected:** MachinePool reaches Ready=True status within 5 minutes without errors

9. **Action:** Retrieve target cluster kubeconfig and verify MachineSet creation
   ```bash
   # Get kubeconfig
   oc get secret azgov-nonzoned-cluster-admin-kubeconfig \
     -n hive-test-azuregov \
     -o jsonpath='{.data.kubeconfig}' | base64 -d > /tmp/azgov-kubeconfig

   # Verify connectivity to target cluster
   oc --kubeconfig=/tmp/azgov-kubeconfig get nodes

   # Count MachineSets created for MachinePool
   MACHINESET_COUNT=$(oc --kubeconfig=/tmp/azgov-kubeconfig \
     get machinesets -n openshift-machine-api \
     --selector hive.openshift.io/machine-pool=worker --no-headers | wc -l)

   echo "MachineSet count: $MACHINESET_COUNT"

   # List MachineSets
   oc --kubeconfig=/tmp/azgov-kubeconfig \
     get machinesets -n openshift-machine-api \
     --selector hive.openshift.io/machine-pool=worker
   ```
   **Expected:** Exactly 1 MachineSet created (non-zoned deployment), MachineSet is visible and active

10. **Action:** Verify MachineSet configuration for non-zoned deployment
    ```bash
    # Get MachineSet name
    MACHINESET_NAME=$(oc --kubeconfig=/tmp/azgov-kubeconfig \
      get machinesets -n openshift-machine-api \
      --selector hive.openshift.io/machine-pool=worker \
      -o jsonpath='{.items[0].metadata.name}')

    echo "MachineSet name: $MACHINESET_NAME"

    # Check zone configuration (should be empty)
    ZONE_CONFIG=$(oc --kubeconfig=/tmp/azgov-kubeconfig \
      get machineset $MACHINESET_NAME -n openshift-machine-api \
      -o jsonpath='{.spec.template.spec.providerSpec.value.zone}')

    echo "Zone configuration: '$ZONE_CONFIG'"

    # Verify region in MachineSet
    MACHINESET_REGION=$(oc --kubeconfig=/tmp/azgov-kubeconfig \
      get machineset $MACHINESET_NAME -n openshift-machine-api \
      -o jsonpath='{.spec.template.spec.providerSpec.value.location}')

    echo "MachineSet region: $MACHINESET_REGION"

    # Check replica count
    MACHINESET_REPLICAS=$(oc --kubeconfig=/tmp/azgov-kubeconfig \
      get machineset $MACHINESET_NAME -n openshift-machine-api \
      -o jsonpath='{.spec.replicas}')

    echo "MachineSet replicas: $MACHINESET_REPLICAS"
    ```
    **Expected:** Zone field empty (non-zoned), region is usgovtexas, replicas equals 3 (matches MachinePool)

11. **Action:** Verify controller logs indicate non-zoned deployment for Azure Gov
    ```bash
    # Get hive-controller pod
    HIVE_POD=$(oc get pods -n hive -l control-plane=controller-manager \
      -o jsonpath='{.items[0].metadata.name}')

    # Search for non-zoned deployment message specific to usgovtexas
    oc logs $HIVE_POD -n hive | \
      grep "usgovtexas" | grep -i "zone"

    # Count non-zoned deployment log entries
    LOG_COUNT=$(oc logs $HIVE_POD -n hive | \
      grep -c "No availability zones detected for region. Using non-zoned deployment.")

    echo "Non-zoned deployment log count: $LOG_COUNT"
    ```
    **Expected:** Log message present indicating non-zoned deployment for usgovtexas region, at least 1 occurrence

12. **Action:** Verify no error messages related to "zero zones returned"
    ```bash
    # Search for old error message that should no longer appear
    ERROR_COUNT=$(oc logs $HIVE_POD -n hive | \
      grep "worker-azgov" | grep -c "zero zones returned for region usgovtexas")

    echo "Error message count (should be 0): $ERROR_COUNT"

    # Check for any error conditions in MachinePool
    oc get machinepool worker-azgov -n hive-test-azuregov \
      -o jsonpath='{.status.conditions[?(@.status=="False")]}'
    ```
    **Expected:** Zero occurrences of "zero zones returned" error, no error conditions in MachinePool status

13. **Action:** Verify Machine resources provisioned successfully
    ```bash
    # Wait for Machine provisioning
    sleep 180

    # List Machines
    oc --kubeconfig=/tmp/azgov-kubeconfig \
      get machines -n openshift-machine-api \
      --selector machine.openshift.io/cluster-api-machineset=$MACHINESET_NAME

    # Count Machines
    MACHINE_COUNT=$(oc --kubeconfig=/tmp/azgov-kubeconfig \
      get machines -n openshift-machine-api \
      --selector machine.openshift.io/cluster-api-machineset=$MACHINESET_NAME \
      --no-headers | wc -l)

    echo "Machine count: $MACHINE_COUNT"

    # Check Machine phases
    oc --kubeconfig=/tmp/azgov-kubeconfig \
      get machines -n openshift-machine-api \
      --selector machine.openshift.io/cluster-api-machineset=$MACHINESET_NAME \
      -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.status.phase}{"\n"}{end}'
    ```
    **Expected:** 3 Machine resources created, progressing to Running phase within 10 minutes

14. **Action:** Verify worker nodes join the cluster
    ```bash
    # Wait for nodes to become ready
    sleep 300

    # Check node count
    NODE_COUNT=$(oc --kubeconfig=/tmp/azgov-kubeconfig get nodes \
      -l node-role.kubernetes.io/worker --no-headers | wc -l)

    echo "Worker node count: $NODE_COUNT"

    # Check node readiness
    oc --kubeconfig=/tmp/azgov-kubeconfig get nodes \
      -l node-role.kubernetes.io/worker \
      -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.status.conditions[?(@.type=="Ready")].status}{"\n"}{end}'
    ```
    **Expected:** At least 3 worker nodes present, all nodes show Ready=True status

15. **Action:** Validate Azure resources created in correct region without zones
    ```bash
    # Access Azure Gov portal or use Azure CLI with Gov cloud endpoints
    # Verify VM resources created in usgovtexas region
    # Check that VMs are NOT assigned to availability zones

    # Using Azure CLI (if configured for Azure Gov):
    az cloud set --name AzureUSGovernment
    az login

    # Get resource group for cluster
    RG_NAME="azgov-nonzoned-<infra-id>-rg"

    # List VMs and check zone property
    az vm list -g $RG_NAME --query "[].{name:name, location:location, zones:zones}" -o table
    ```
    **Expected:** VMs created in usgovtexas region, zones field empty or null (non-zoned deployment confirmed at Azure level)

#### Manual Verification Checklist
- [ ] Azure Government credentials successfully configured
- [ ] ClusterDeployment installed in usgovtexas region
- [ ] MachinePool created without errors
- [ ] Exactly 1 MachineSet generated (non-zoned)
- [ ] MachineSet has no zone configuration
- [ ] Controller logs show "Using non-zoned deployment" message
- [ ] No "zero zones returned" errors in logs
- [ ] 3 Machine resources provisioned successfully
- [ ] 3 worker nodes joined cluster and are Ready
- [ ] Azure resources confirmed in non-zoned configuration

#### Notes
- Azure Government cloud requires special subscription access and cannot be tested in standard E2E environments
- Test requires manual verification of Azure Gov portal/CLI to confirm non-zoned resource deployment
- Similar procedure applies to other Azure Gov regions: usgovarizona, usdodeast, usdodcentral
- Test validates fix works across both Azure Commercial and Azure Government clouds
- Estimated total time: 90-120 minutes (60-75 min cluster install + 15-20 min MachinePool testing + verification)
