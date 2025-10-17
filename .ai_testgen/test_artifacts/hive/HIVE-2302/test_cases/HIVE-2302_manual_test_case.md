# Test Case: HIVE-2302
**Component:** hive
**Summary:** Pass installer's metadata.json directly to destroyer

## Test Overview
- **Total Test Cases:** 8
- **Test Types:** Manual
- **Estimated Time:** 480-600 minutes (8-10 hours total)

## Test Cases

### Test Case HIVE-2302_M001
**Name:** AWS Cluster with HostedZoneRole - Shared VPC Deprovision
**Description:** Validate metadata.json correctly passes HostedZoneRole for AWS clusters using hosted zones in another account, test both modern and retrofit scenarios
**Type:** Manual
**Priority:** High

#### Prerequisites
- Two AWS accounts: Account A (cluster), Account B (hosted zone owner)
- IAM role in Account B allowing Account A to manage DNS records
- HostedZoneRole ARN configured in Account A
- Hive operator with HIVE-2302 changes deployed

#### Test Steps
1. **Action:** Prepare HostedZoneRole in Account B
   ```bash
   # In Account B: Create IAM role for DNS delegation
   # Note the role ARN: arn:aws:iam::ACCOUNT_B:role/HostedZoneRole

   # Configure trust relationship allowing Account A to assume role
   # Test role assumption from Account A
   aws sts assume-role --role-arn arn:aws:iam::ACCOUNT_B:role/HostedZoneRole \
     --role-session-name test-session
   ```
   **Expected:** Role exists and can be assumed from Account A

2. **Action:** Create install-config with HostedZoneRole
   ```bash
   cat > install-config.yaml <<EOF
   apiVersion: v1
   baseDomain: example.com
   metadata:
     name: aws-hostedzone-test
   platform:
     aws:
       region: us-east-1
       hostedZone: Z1234567890ABC  # Hosted zone ID in Account B
       hostedZoneRole: arn:aws:iam::ACCOUNT_B:role/HostedZoneRole
   pullSecret: '...'
   sshKey: '...'
   EOF

   oc create secret generic install-config \
     --from-file=install-config.yaml=install-config.yaml \
     -n hive-test-hostedzone
   ```
   **Expected:** install-config Secret created with HostedZoneRole configuration

3. **Action:** Provision cluster and monitor metadata.json Secret creation
   ```bash
   oc create -f clusterdeployment-hostedzone.yaml

   oc wait --for=condition=Installed=true \
     clusterdeployment/aws-hostedzone-test \
     -n hive-test-hostedzone --timeout=60m

   SECRET_NAME=$(oc get clusterdeployment aws-hostedzone-test \
     -n hive-test-hostedzone \
     -o jsonpath='{.spec.clusterMetadata.metadataJSONSecretRef.name}')
   ```
   **Expected:** Cluster provisions successfully, metadata.json Secret created

4. **Action:** Verify HostedZoneRole in metadata.json
   ```bash
   oc get secret $SECRET_NAME -n hive-test-hostedzone \
     -o jsonpath='{.data.metadata\.json}' | base64 -d | jq . > /tmp/hostedzone-metadata.json

   HOSTED_ZONE_ROLE=$(cat /tmp/hostedzone-metadata.json | jq -r '.aws.hostedZoneRole')
   echo "HostedZoneRole in metadata: $HOSTED_ZONE_ROLE"

   [[ "$HOSTED_ZONE_ROLE" == "arn:aws:iam::ACCOUNT_B:role/HostedZoneRole" ]] && \
     echo "HostedZoneRole match: SUCCESS" || echo "HostedZoneRole match: FAILED"
   ```
   **Expected:** metadata.json contains correct HostedZoneRole ARN

5. **Action:** Verify DNS records created in Account B hosted zone
   ```bash
   # Using Account B credentials
   aws route53 list-resource-record-sets --hosted-zone-id Z1234567890ABC \
     --query "ResourceRecordSets[?contains(Name, 'aws-hostedzone-test')]" > /tmp/dns-records.json

   cat /tmp/dns-records.json | jq '.[] | .Name'
   ```
   **Expected:** DNS records for cluster exist in Account B hosted zone

6. **Action:** Delete ClusterDeployment and verify deprovision job configuration
   ```bash
   oc delete clusterdeployment aws-hostedzone-test -n hive-test-hostedzone

   sleep 30

   DEPROVISION_POD=$(oc get pods -n hive-test-hostedzone \
     -l hive.openshift.io/cluster-deprovision-name \
     -o jsonpath='{.items[0].metadata.name}')

   oc logs $DEPROVISION_POD -n hive-test-hostedzone --tail=100 | \
     grep -i "hostedzonerole"
   ```
   **Expected:** Deprovision logs show HostedZoneRole being used for DNS cleanup

7. **Action:** Monitor DNS record deletion in Account B
   ```bash
   # Wait for deprovision to progress
   sleep 300

   # Check DNS records in Account B
   aws route53 list-resource-record-sets --hosted-zone-id Z1234567890ABC \
     --query "ResourceRecordSets[?contains(Name, 'aws-hostedzone-test')]"
   ```
   **Expected:** DNS records being deleted from Account B hosted zone

8. **Action:** Verify complete deprovision and cleanup
   ```bash
   DEPROVISION_NAME=$(oc get clusterdeprovision -n hive-test-hostedzone \
     -o jsonpath='{.items[0].metadata.name}')

   oc wait --for=condition=Completed=true \
     clusterdeprovision/$DEPROVISION_NAME \
     -n hive-test-hostedzone --timeout=30m

   # Final DNS check - should be empty
   aws route53 list-resource-record-sets --hosted-zone-id Z1234567890ABC \
     --query "ResourceRecordSets[?contains(Name, 'aws-hostedzone-test')]"
   ```
   **Expected:** Deprovision completes, no DNS records remain in Account B

9. **Action:** **RETROFIT SCENARIO** - Test legacy cluster with bespoke HostedZoneRole
   ```bash
   # Manually create ClusterDeployment with bespoke field populated
   # but without metadata.json Secret (simulating pre-HIVE-2302 cluster)

   # Populate CD.Spec.ClusterMetadata.Platform.AWS.HostedZoneRole
   # Trigger reconciliation to generate Secret
   # Verify HostedZoneRole flows into generated metadata.json
   # Deprovision and confirm DNS cleanup in Account B
   ```
   **Expected:** Retrofit scenario successfully generates metadata.json with HostedZoneRole, deprovision works

#### Manual Verification Checklist
- [ ] HostedZoneRole configured correctly in two-account setup
- [ ] Cluster provisions with DNS records in Account B
- [ ] metadata.json Secret contains HostedZoneRole ARN
- [ ] Deprovision uses HostedZoneRole to clean DNS in Account B
- [ ] All DNS records deleted from Account B hosted zone
- [ ] Retrofit scenario generates correct metadata.json
- [ ] No orphaned DNS records or AWS resources

### Test Case HIVE-2302_M002
**Name:** Azure Cluster with Custom ResourceGroupName
**Description:** Validate metadata.json handling of custom Azure resource group for both modern and retrofit scenarios
**Type:** Manual
**Priority:** High

#### Prerequisites
- Azure subscription with permission to create resources
- Pre-created Azure resource group for testing
- Hive operator with HIVE-2302 changes

#### Test Steps
1. **Action:** Create custom Azure resource group
   ```bash
   CUSTOM_RG="hive-custom-rg-test"
   az group create --name $CUSTOM_RG --location eastus

   az group show --name $CUSTOM_RG
   ```
   **Expected:** Custom resource group created successfully

2. **Action:** Create install-config with custom resourceGroupName
   ```bash
   cat > install-config.yaml <<EOF
   apiVersion: v1
   baseDomain: example.com
   metadata:
     name: azure-customrg-test
   platform:
     azure:
       region: eastus
       baseDomainResourceGroupName: azure-dns-rg
       resourceGroupName: $CUSTOM_RG
   pullSecret: '...'
   sshKey: '...'
   EOF

   oc create secret generic install-config \
     --from-file=install-config.yaml=install-config.yaml \
     -n hive-test-azure-rg
   ```
   **Expected:** install-config created with custom resourceGroupName

3. **Action:** Provision cluster targeting custom resource group
   ```bash
   oc create -f clusterdeployment-azure-customrg.yaml

   oc wait --for=condition=Installed=true \
     clusterdeployment/azure-customrg-test \
     -n hive-test-azure-rg --timeout=60m
   ```
   **Expected:** Cluster provisions into custom resource group

4. **Action:** Verify resources created in custom resource group
   ```bash
   az resource list --resource-group $CUSTOM_RG \
     --query '[].{name:name, type:type}' --output table

   RESOURCE_COUNT=$(az resource list --resource-group $CUSTOM_RG --query 'length(@)')
   echo "Resources in custom RG: $RESOURCE_COUNT"
   ```
   **Expected:** Cluster resources (VMs, NICs, disks, NSG, etc.) in custom resource group

5. **Action:** Verify metadata.json contains custom ResourceGroupName
   ```bash
   SECRET_NAME=$(oc get clusterdeployment azure-customrg-test \
     -n hive-test-azure-rg \
     -o jsonpath='{.spec.clusterMetadata.metadataJSONSecretRef.name}')

   oc get secret $SECRET_NAME -n hive-test-azure-rg \
     -o jsonpath='{.data.metadata\.json}' | base64 -d | \
     jq '.azure.resourceGroupName'
   ```
   **Expected:** metadata.json contains custom resource group name

6. **Action:** Delete ClusterDeployment and monitor deprovision
   ```bash
   oc delete clusterdeployment azure-customrg-test -n hive-test-azure-rg

   DEPROVISION_POD=$(oc get pods -n hive-test-azure-rg \
     -l hive.openshift.io/cluster-deprovision-name \
     -o jsonpath='{.items[0].metadata.name}')

   oc logs $DEPROVISION_POD -n hive-test-azure-rg -f --tail=100
   ```
   **Expected:** Deprovision targets custom resource group (logs should reference it)

7. **Action:** Verify resource deletion from custom resource group
   ```bash
   sleep 120

   az resource list --resource-group $CUSTOM_RG \
     --query '[].{name:name, type:type}' --output table

   RESOURCE_COUNT_DURING=$(az resource list --resource-group $CUSTOM_RG --query 'length(@)')
   echo "Resources during deprovision: $RESOURCE_COUNT_DURING"
   ```
   **Expected:** Resources being deleted from custom resource group

8. **Action:** Verify complete cleanup
   ```bash
   DEPROVISION_NAME=$(oc get clusterdeprovision -n hive-test-azure-rg \
     -o jsonpath='{.items[0].metadata.name}')

   oc wait --for=condition=Completed=true \
     clusterdeprovision/$DEPROVISION_NAME \
     -n hive-test-azure-rg --timeout=30m

   # Check if resource group still exists and is empty
   az group show --name $CUSTOM_RG 2>&1
   az resource list --resource-group $CUSTOM_RG --query 'length(@)'
   ```
   **Expected:** Resource group empty or deleted (depending on whether it was pre-existing)

9. **Action:** **RETROFIT SCENARIO** - Legacy cluster with bespoke ResourceGroupName
   ```bash
   # Create CD with CD.Spec.ClusterMetadata.Platform.Azure.ResourceGroupName populated
   # Without metadata.json Secret (legacy state)
   # Trigger reconciliation
   # Verify Secret generation includes ResourceGroupName
   # Deprovision and verify cleanup
   ```
   **Expected:** Retrofit correctly generates metadata.json with ResourceGroupName

#### Manual Verification Checklist
- [ ] Custom resource group created before cluster provisioning
- [ ] Cluster resources deployed to custom resource group
- [ ] metadata.json contains custom ResourceGroupName
- [ ] Deprovision targets correct resource group
- [ ] All resources deleted from custom resource group
- [ ] Retrofit scenario works correctly
- [ ] No resource leaks in Azure subscription

### Test Case HIVE-2302_M003
**Name:** GCP Cluster with NetworkProjectID - Shared VPC Deprovision
**Description:** Validate metadata.json handling of GCP shared VPC with network resources in separate project
**Type:** Manual
**Priority:** High

#### Prerequisites
- Two GCP projects: Project A (cluster), Project B (shared VPC network)
- Shared VPC configured with Project B as host, Project A as service project
- Appropriate IAM permissions for cross-project operations
- Hive operator with HIVE-2302 changes

#### Test Steps
1. **Action:** Configure GCP shared VPC between projects
   ```bash
   # In Project B (host project)
   gcloud compute shared-vpc enable $PROJECT_B

   # Add Project A as service project
   gcloud compute shared-vpc associated-projects add $PROJECT_A \
     --host-project=$PROJECT_B

   # Verify shared VPC configuration
   gcloud compute shared-vpc get-host-project $PROJECT_A
   ```
   **Expected:** Project A configured as service project of Project B shared VPC

2. **Action:** Create install-config with networkProjectID
   ```bash
   cat > install-config.yaml <<EOF
   apiVersion: v1
   baseDomain: example.com
   metadata:
     name: gcp-sharedvpc-test
   platform:
     gcp:
       region: us-east1
       projectID: $PROJECT_A
       networkProjectID: $PROJECT_B
       network: shared-vpc-network
   pullSecret: '...'
   sshKey: '...'
   EOF

   oc create secret generic install-config \
     --from-file=install-config.yaml=install-config.yaml \
     -n hive-test-gcp-sharedvpc
   ```
   **Expected:** install-config created with networkProjectID

3. **Action:** Provision cluster and verify network resources
   ```bash
   oc create -f clusterdeployment-gcp-sharedvpc.yaml

   oc wait --for=condition=Installed=true \
     clusterdeployment/gcp-sharedvpc-test \
     -n hive-test-gcp-sharedvpc --timeout=60m

   # Check firewall rules in network project (Project B)
   gcloud compute firewall-rules list --project=$PROJECT_B \
     --filter="name~gcp-sharedvpc-test" --format=table
   ```
   **Expected:** Cluster provisions, firewall rules created in Project B (network project)

4. **Action:** Verify metadata.json contains NetworkProjectID
   ```bash
   SECRET_NAME=$(oc get clusterdeployment gcp-sharedvpc-test \
     -n hive-test-gcp-sharedvpc \
     -o jsonpath='{.spec.clusterMetadata.metadataJSONSecretRef.name}')

   oc get secret $SECRET_NAME -n hive-test-gcp-sharedvpc \
     -o jsonpath='{.data.metadata\.json}' | base64 -d | jq . > /tmp/gcp-sharedvpc-metadata.json

   NETWORK_PROJECT=$(cat /tmp/gcp-sharedvpc-metadata.json | jq -r '.gcp.networkProjectID')
   echo "NetworkProjectID in metadata: $NETWORK_PROJECT"

   [[ "$NETWORK_PROJECT" == "$PROJECT_B" ]] && echo "Match: SUCCESS" || echo "Match: FAILED"
   ```
   **Expected:** metadata.json contains correct NetworkProjectID

5. **Action:** Verify compute resources in Project A, network resources in Project B
   ```bash
   # Instances in Project A
   gcloud compute instances list --project=$PROJECT_A \
     --filter="labels.kubernetes-io-cluster-*" --format=table

   # Firewall rules in Project B
   gcloud compute firewall-rules list --project=$PROJECT_B \
     --filter="name~gcp-sharedvpc-test" --format=table
   ```
   **Expected:** VMs in Project A, firewall rules in Project B

6. **Action:** Delete ClusterDeployment and monitor cross-project cleanup
   ```bash
   oc delete clusterdeployment gcp-sharedvpc-test -n hive-test-gcp-sharedvpc

   DEPROVISION_POD=$(oc get pods -n hive-test-gcp-sharedvpc \
     -l hive.openshift.io/cluster-deprovision-name \
     -o jsonpath='{.items[0].metadata.name}')

   oc logs $DEPROVISION_POD -n hive-test-gcp-sharedvpc -f --tail=100 | \
     grep -i "networkprojectid"
   ```
   **Expected:** Deprovision logs show operations in network project

7. **Action:** Verify firewall rules deleted from Project B
   ```bash
   sleep 180

   gcloud compute firewall-rules list --project=$PROJECT_B \
     --filter="name~gcp-sharedvpc-test" --format=table
   ```
   **Expected:** Firewall rules being deleted from Project B

8. **Action:** Verify complete cleanup in both projects
   ```bash
   DEPROVISION_NAME=$(oc get clusterdeprovision -n hive-test-gcp-sharedvpc \
     -o jsonpath='{.items[0].metadata.name}')

   oc wait --for=condition=Completed=true \
     clusterdeprovision/$DEPROVISION_NAME \
     -n hive-test-gcp-sharedvpc --timeout=30m

   # Check Project A - should have no instances
   gcloud compute instances list --project=$PROJECT_A \
     --filter="labels.kubernetes-io-cluster-*"

   # Check Project B - should have no cluster firewall rules
   gcloud compute firewall-rules list --project=$PROJECT_B \
     --filter="name~gcp-sharedvpc-test"
   ```
   **Expected:** No resources in Project A, no cluster firewall rules in Project B

9. **Action:** **RETROFIT SCENARIO** - Legacy cluster with bespoke NetworkProjectID
   ```bash
   # Create CD with CD.Spec.ClusterMetadata.Platform.GCP.NetworkProjectID
   # Without metadata.json Secret
   # Trigger reconciliation
   # Verify Secret includes NetworkProjectID
   # Deprovision and verify cross-project cleanup
   ```
   **Expected:** Retrofit generates correct metadata, cross-project cleanup succeeds

#### Manual Verification Checklist
- [ ] Shared VPC configured between two GCP projects
- [ ] Cluster provisions with compute in Project A, network in Project B
- [ ] metadata.json contains NetworkProjectID
- [ ] Deprovision cleans firewall rules from Project B
- [ ] Deprovision cleans compute resources from Project A
- [ ] Retrofit scenario works correctly
- [ ] No orphaned resources in either project

### Test Case HIVE-2302_M004
**Name:** vSphere Pre-Zonal Metadata Format Handling
**Description:** Verify retrofit and deprovision of vSphere clusters using pre-zonal metadata.json format
**Type:** Manual
**Priority:** Medium

#### Prerequisites
- vSphere infrastructure access
- OpenShift release before vSphere zonal support (e.g., 4.13 or earlier)
- Existing vSphere cluster or ability to provision one
- Hive operator with HIVE-2302 changes

#### Test Steps
1. **Action:** Identify or provision vSphere cluster from pre-zonal release
   ```bash
   # Use existing pre-zonal vSphere cluster or provision one with old release
   CLUSTER_NAME="vsphere-prezonal-test"
   NAMESPACE="hive-test-vsphere-prezonal"

   oc get clusterdeployment $CLUSTER_NAME -n $NAMESPACE
   ```
   **Expected:** vSphere cluster from pre-zonal release available

2. **Action:** Examine existing metadata structure (if available)
   ```bash
   PROVISION_NAME=$(oc get clusterprovision -n $NAMESPACE \
     --sort-by=.metadata.creationTimestamp \
     -o jsonpath='{.items[-1].metadata.name}')

   oc get clusterprovision $PROVISION_NAME -n $NAMESPACE \
     -o jsonpath='{.spec.metadataJSON}' | jq . > /tmp/vsphere-prezonal-metadata.json

   # Check structure - should have flat vsphere.username/password fields
   cat /tmp/vsphere-prezonal-metadata.json | jq '.vsphere'
   ```
   **Expected:** Pre-zonal structure with username/password at vsphere level (not in vcenters array)

3. **Action:** Trigger metadata.json Secret retrofit
   ```bash
   oc annotate clusterdeployment $CLUSTER_NAME -n $NAMESPACE \
     hive.openshift.io/reconcile-trigger="$(date +%s)" --overwrite

   sleep 10

   SECRET_NAME=$(oc get clusterdeployment $CLUSTER_NAME -n $NAMESPACE \
     -o jsonpath='{.spec.clusterMetadata.metadataJSONSecretRef.name}')

   oc get secret $SECRET_NAME -n $NAMESPACE \
     -o jsonpath='{.data.metadata\.json}' | base64 -d | jq . > /tmp/vsphere-retrofitted.json
   ```
   **Expected:** Secret created with pre-zonal metadata structure preserved

4. **Action:** Verify credentials scrubbed in Secret
   ```bash
   cat /tmp/vsphere-retrofitted.json | jq '.vsphere.username'
   cat /tmp/vsphere-retrofitted.json | jq '.vsphere.password'
   ```
   **Expected:** username and password fields empty or scrubbed

5. **Action:** Delete ClusterDeployment and verify deprovision job
   ```bash
   oc delete clusterdeployment $CLUSTER_NAME -n $NAMESPACE

   sleep 30

   DEPROVISION_POD=$(oc get pods -n $NAMESPACE \
     -l hive.openshift.io/cluster-deprovision-name \
     -o jsonpath='{.items[0].metadata.name}')

   # Check environment variables for credential injection
   oc get pod $DEPROVISION_POD -n $NAMESPACE \
     -o jsonpath='{.spec.containers[0].env[*].name}' | grep VSPHERE
   ```
   **Expected:** Deprovision pod has VSPHERE_USERNAME and VSPHERE_PASSWORD env vars for credential injection

6. **Action:** Monitor deprovision logs for pre-zonal format handling
   ```bash
   oc logs $DEPROVISION_POD -n $NAMESPACE -f --tail=100 | \
     grep -E "vsphere|credential"
   ```
   **Expected:** Logs show credential injection into pre-zonal metadata structure

7. **Action:** Verify vSphere resource cleanup
   ```bash
   # Access vSphere vCenter and check for VMs, folders, resource pools
   # Use vCenter UI or govc CLI
   # Verify cluster VMs are being deleted

   DEPROVISION_NAME=$(oc get clusterdeprovision -n $NAMESPACE \
     -o jsonpath='{.items[0].metadata.name}')

   oc wait --for=condition=Completed=true \
     clusterdeprovision/$DEPROVISION_NAME \
     -n $NAMESPACE --timeout=30m
   ```
   **Expected:** Deprovision completes, all VMs and folders deleted from vCenter

#### Manual Verification Checklist
- [ ] Pre-zonal vSphere cluster identified/provisioned
- [ ] metadata.json has flat vsphere.username/password structure
- [ ] Credentials scrubbed from Secret
- [ ] Deprovision job injects credentials via env vars
- [ ] Generic destroyer handles pre-zonal format correctly
- [ ] All vSphere resources (VMs, folders, resource pools) deleted
- [ ] No orphaned resources in vCenter inventory

### Test Case HIVE-2302_M005
**Name:** vSphere Post-Zonal Metadata Format Handling
**Description:** Verify vSphere clusters with zonal support using vCenters array metadata format
**Type:** Manual
**Priority:** Medium

#### Prerequisites
- vSphere infrastructure with zonal configuration support
- OpenShift release with vSphere zonal support (4.14+)
- Hive operator with HIVE-2302 changes

#### Test Steps
1. **Action:** Provision vSphere cluster with zonal configuration
   ```bash
   # Create install-config with multiple failure domains/zones
   cat > install-config.yaml <<EOF
   apiVersion: v1
   baseDomain: example.com
   metadata:
     name: vsphere-zonal-test
   platform:
     vsphere:
       vcenters:
       - server: vcenter1.example.com
         datacenters:
         - datacenter1
       - server: vcenter2.example.com
         datacenters:
         - datacenter2
       failureDomains:
       - name: zone-a
         zone: zone-a
         region: region1
       - name: zone-b
         zone: zone-b
         region: region1
   pullSecret: '...'
   sshKey: '...'
   EOF
   ```
   **Expected:** install-config with zonal vSphere configuration

2. **Action:** Deploy cluster and wait for installation
   ```bash
   oc create -f clusterdeployment-vsphere-zonal.yaml

   oc wait --for=condition=Installed=true \
     clusterdeployment/vsphere-zonal-test \
     -n hive-test-vsphere-zonal --timeout=60m
   ```
   **Expected:** Cluster provisions successfully with zonal configuration

3. **Action:** Verify post-zonal metadata.json structure
   ```bash
   SECRET_NAME=$(oc get clusterdeployment vsphere-zonal-test \
     -n hive-test-vsphere-zonal \
     -o jsonpath='{.spec.clusterMetadata.metadataJSONSecretRef.name}')

   oc get secret $SECRET_NAME -n hive-test-vsphere-zonal \
     -o jsonpath='{.data.metadata\.json}' | base64 -d | jq . > /tmp/vsphere-zonal-metadata.json

   # Check for vcenters array structure
   cat /tmp/vsphere-zonal-metadata.json | jq '.vsphere.vcenters'
   ```
   **Expected:** metadata.json contains vcenters array with per-vCenter configuration

4. **Action:** Verify credentials scrubbed from vCenters array
   ```bash
   cat /tmp/vsphere-zonal-metadata.json | jq '.vsphere.vcenters[].username'
   cat /tmp/vsphere-zonal-metadata.json | jq '.vsphere.vcenters[].password'
   ```
   **Expected:** username/password fields in each vCenter element are empty/scrubbed

5. **Action:** Verify cluster resources across zones
   ```bash
   # Check vCenter inventory across both vCenters/datacenters
   # Verify VMs distributed across zones
   # Use vCenter UI or govc to list VMs in each datacenter
   ```
   **Expected:** Cluster VMs distributed across configured zones/vCenters

6. **Action:** Delete ClusterDeployment and verify deprovision configuration
   ```bash
   oc delete clusterdeployment vsphere-zonal-test -n hive-test-vsphere-zonal

   sleep 30

   DEPROVISION_POD=$(oc get pods -n hive-test-vsphere-zonal \
     -l hive.openshift.io/cluster-deprovision-name \
     -o jsonpath='{.items[0].metadata.name}')

   oc get pod $DEPROVISION_POD -n hive-test-vsphere-zonal \
     -o jsonpath='{.spec.containers[0].env[*].name}' | grep VSPHERE
   ```
   **Expected:** Deprovision pod has credential environment variables

7. **Action:** Monitor deprovision for multi-vCenter handling
   ```bash
   oc logs $DEPROVISION_POD -n hive-test-vsphere-zonal -f --tail=100 | \
     grep -E "vcenter|zone|credential"
   ```
   **Expected:** Logs show credentials injected into vcenters array, operations across multiple vCenters

8. **Action:** Verify cleanup across all zones
   ```bash
   DEPROVISION_NAME=$(oc get clusterdeprovision -n hive-test-vsphere-zonal \
     -o jsonpath='{.items[0].metadata.name}')

   oc wait --for=condition=Completed=true \
     clusterdeprovision/$DEPROVISION_NAME \
     -n hive-test-vsphere-zonal --timeout=30m

   # Check both vCenters for remaining resources
   # Verify VMs deleted from all datacenters/zones
   ```
   **Expected:** Deprovision completes, resources deleted from all zones/vCenters

#### Manual Verification Checklist
- [ ] vSphere cluster provisioned with zonal configuration
- [ ] metadata.json contains vcenters array structure
- [ ] Credentials scrubbed from each vCenter element
- [ ] Deprovision injects credentials into array elements
- [ ] Resources deleted from all zones/vCenters
- [ ] No orphaned VMs, folders, or resource pools
- [ ] Generic destroyer handles post-zonal format correctly

### Test Case HIVE-2302_M006
**Name:** IBMCloud Cluster Deprovision with metadata.json
**Description:** Validate IBMCloud cluster deprovision using generic destroyer with metadata.json
**Type:** Manual
**Priority:** Medium

#### Prerequisites
- IBMCloud account and credentials
- Hive operator with HIVE-2302 changes
- IBMCloud API key configured

#### Test Steps
1. **Action:** Provision IBMCloud cluster
   ```bash
   oc create namespace hive-test-ibmcloud

   oc create secret generic ibmcloud-creds \
     --from-literal=ibmcloud_api_key=$IBMCLOUD_API_KEY \
     -n hive-test-ibmcloud

   oc create -f clusterdeployment-ibmcloud.yaml
   ```
   **Expected:** ClusterDeployment created for IBMCloud

2. **Action:** Wait for installation and verify metadata.json Secret
   ```bash
   oc wait --for=condition=Installed=true \
     clusterdeployment/ibmcloud-test \
     -n hive-test-ibmcloud --timeout=60m

   SECRET_NAME=$(oc get clusterdeployment ibmcloud-test \
     -n hive-test-ibmcloud \
     -o jsonpath='{.spec.clusterMetadata.metadataJSONSecretRef.name}')

   oc get secret $SECRET_NAME -n hive-test-ibmcloud \
     -o jsonpath='{.data.metadata\.json}' | base64 -d | jq .
   ```
   **Expected:** Cluster installs, metadata.json Secret contains IBMCloud-specific fields

3. **Action:** Deprovision and verify generic destroyer usage
   ```bash
   oc delete clusterdeployment ibmcloud-test -n hive-test-ibmcloud

   DEPROVISION_POD=$(oc get pods -n hive-test-ibmcloud \
     -l hive.openshift.io/cluster-deprovision-name \
     -o jsonpath='{.items[0].metadata.name}')

   oc logs $DEPROVISION_POD -n hive-test-ibmcloud -f
   ```
   **Expected:** Deprovision uses generic destroyer, completes successfully

4. **Action:** Verify IBMCloud resource cleanup
   ```bash
   # Use ibmcloud CLI to check for remaining resources
   ibmcloud resource service-instances --output json | \
     jq '.[] | select(.name | contains("ibmcloud-test"))'
   ```
   **Expected:** No remaining IBMCloud resources for cluster

#### Manual Verification Checklist
- [ ] IBMCloud cluster provisions successfully
- [ ] metadata.json Secret created
- [ ] Deprovision uses generic destroyer
- [ ] All IBMCloud resources cleaned up

### Test Case HIVE-2302_M007
**Name:** Nutanix Cluster Deprovision with metadata.json
**Description:** Validate Nutanix cluster deprovision with credential injection into metadata.json
**Type:** Manual
**Priority:** Medium

#### Prerequisites
- Nutanix infrastructure access
- Prism Central credentials
- Hive operator with HIVE-2302 changes

#### Test Steps
1. **Action:** Provision Nutanix cluster
   ```bash
   oc create namespace hive-test-nutanix

   oc create secret generic nutanix-creds \
     --from-literal=username=$NUTANIX_USERNAME \
     --from-literal=password=$NUTANIX_PASSWORD \
     -n hive-test-nutanix

   oc create -f clusterdeployment-nutanix.yaml
   ```
   **Expected:** ClusterDeployment created for Nutanix

2. **Action:** Verify metadata.json with scrubbed credentials
   ```bash
   SECRET_NAME=$(oc get clusterdeployment nutanix-test \
     -n hive-test-nutanix \
     -o jsonpath='{.spec.clusterMetadata.metadataJSONSecretRef.name}')

   oc get secret $SECRET_NAME -n hive-test-nutanix \
     -o jsonpath='{.data.metadata\.json}' | base64 -d | jq '.nutanix'
   ```
   **Expected:** Nutanix credentials scrubbed from metadata.json

3. **Action:** Deprovision with credential injection
   ```bash
   oc delete clusterdeployment nutanix-test -n hive-test-nutanix

   DEPROVISION_POD=$(oc get pods -n hive-test-nutanix \
     -l hive.openshift.io/cluster-deprovision-name \
     -o jsonpath='{.items[0].metadata.name}')

   oc get pod $DEPROVISION_POD -n hive-test-nutanix \
     -o jsonpath='{.spec.containers[0].env[*].name}' | grep NUTANIX
   ```
   **Expected:** Deprovision pod has NUTANIX_USERNAME and NUTANIX_PASSWORD env vars

4. **Action:** Verify Nutanix resource cleanup
   ```bash
   # Check Prism Central for VM cleanup
   # Verify no orphaned VMs or storage
   ```
   **Expected:** All Nutanix resources deleted

#### Manual Verification Checklist
- [ ] Nutanix cluster provisions successfully
- [ ] Credentials scrubbed from metadata.json
- [ ] Deprovision injects credentials via environment
- [ ] All Nutanix VMs and resources cleaned up

### Test Case HIVE-2302_M008
**Name:** OpenStack Cluster Deprovision with metadata.json
**Description:** Validate OpenStack cluster deprovision using generic destroyer
**Type:** Manual
**Priority:** Medium

#### Prerequisites
- OpenStack environment access
- OpenStack credentials (clouds.yaml)
- Hive operator with HIVE-2302 changes

#### Test Steps
1. **Action:** Provision OpenStack cluster
   ```bash
   oc create namespace hive-test-openstack

   oc create secret generic openstack-creds \
     --from-file=clouds.yaml=/path/to/clouds.yaml \
     -n hive-test-openstack

   oc create -f clusterdeployment-openstack.yaml
   ```
   **Expected:** ClusterDeployment created for OpenStack

2. **Action:** Verify metadata.json Secret creation
   ```bash
   SECRET_NAME=$(oc get clusterdeployment openstack-test \
     -n hive-test-openstack \
     -o jsonpath='{.spec.clusterMetadata.metadataJSONSecretRef.name}')

   oc get secret $SECRET_NAME -n hive-test-openstack \
     -o jsonpath='{.data.metadata\.json}' | base64 -d | jq .
   ```
   **Expected:** metadata.json contains OpenStack-specific information

3. **Action:** Deprovision and verify cleanup
   ```bash
   oc delete clusterdeployment openstack-test -n hive-test-openstack

   DEPROVISION_POD=$(oc get pods -n hive-test-openstack \
     -l hive.openshift.io/cluster-deprovision-name \
     -o jsonpath='{.items[0].metadata.name}')

   oc logs $DEPROVISION_POD -n hive-test-openstack -f
   ```
   **Expected:** Generic destroyer runs, completes successfully

4. **Action:** Verify OpenStack resource cleanup
   ```bash
   # Use openstack CLI to check for remaining resources
   openstack server list | grep openstack-test
   openstack network list | grep openstack-test
   ```
   **Expected:** No remaining OpenStack instances or networks

#### Manual Verification Checklist
- [ ] OpenStack cluster provisions successfully
- [ ] metadata.json Secret created
- [ ] Deprovision uses generic destroyer
- [ ] All OpenStack resources (instances, networks, volumes) cleaned up
