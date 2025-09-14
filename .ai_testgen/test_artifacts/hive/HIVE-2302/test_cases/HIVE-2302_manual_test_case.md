# Test Case: HIVE-2302
**Component:** hive
**Summary:** Pass installer's metadata.json directly to destroyer

## Test Overview
- **Total Test Cases:** 5
- **Test Types:** Manual
- **Estimated Time:** 180 minutes

## Test Cases

### Test Case HIVE-2302_M001
**Name:** Legacy Cluster Retrofit with metadata.json Secret Generation and Deprovision
**Description:** Verify that clusters created before HIVE-2302 implementation (legacy clusters) are successfully retrofitted with metadata.json Secret and deprovision cleanly.
**Type:** Manual
**Priority:** High

#### Prerequisites
- Hive management cluster running version BEFORE HIVE-2302 implementation
- Existing ClusterDeployments created on legacy Hive version
- Access to upgrade Hive to version WITH HIVE-2302 implementation
- ClusterProvision and ClusterDeployment.Spec.ClusterMetadata contain bespoke platform fields

#### Test Steps
1. **Action:** Verify legacy cluster exists without metadata.json Secret.
   ```bash
   # Check existing ClusterDeployment
   oc get clusterdeployment legacy-test-cluster -o yaml | grep metadataJSONSecretRef

   # Verify no metadata-json Secret exists
   oc get secret legacy-test-cluster-metadata-json 2>&1 | grep "NotFound"
   ```
   **Expected:** metadataJSONSecretRef field does not exist or is empty. Secret legacy-test-cluster-metadata-json does not exist (NotFound error).

2. **Action:** Upgrade Hive to version containing HIVE-2302 implementation.
   ```bash
   # Deploy new Hive version with HIVE-2302
   oc apply -f hive-deployment-2302.yaml

   # Wait for Hive controller rollout
   oc rollout status deployment/hive-controllers -n hive --timeout=5m

   # Verify new Hive version running
   oc get deployment hive-controllers -n hive -o jsonpath='{.spec.template.spec.containers[0].image}'
   ```
   **Expected:** Hive controllers deployment updated successfully. New image version shows HIVE-2302 or later. Hive controller pods are running.

3. **Action:** Trigger ClusterDeployment controller reconciliation to retrofit metadata.json Secret.
   ```bash
   # Add annotation to trigger reconciliation
   oc annotate clusterdeployment legacy-test-cluster hive.openshift.io/reconcile="$(date +%s)"

   # Wait for controller to reconcile
   sleep 30

   # Check if metadata.json Secret created
   oc get secret legacy-test-cluster-metadata-json
   ```
   **Expected:** metadata.json Secret is created automatically by retrofit logic. Secret name matches pattern: {clustername}-metadata-json

4. **Action:** Verify retrofitted Secret content generated from ClusterProvision or ClusterDeployment.Spec.ClusterMetadata.
   ```bash
   # Extract and verify metadata.json content
   oc get secret legacy-test-cluster-metadata-json -o jsonpath='{.data.metadata\.json}' | base64 -d > /tmp/retrofitted-metadata.json

   # Verify infraID matches
   LEGACY_INFRA=$(oc get clusterdeployment legacy-test-cluster -o jsonpath='{.spec.clusterMetadata.infraID}')
   grep -q "$LEGACY_INFRA" /tmp/retrofitted-metadata.json && echo "PASS: infraID matches" || echo "FAIL: infraID mismatch"

   # Verify platform field exists
   grep -q '"platform"' /tmp/retrofitted-metadata.json && echo "PASS: platform field present" || echo "FAIL: platform field missing"
   ```
   **Expected:** metadata.json contains infraID matching ClusterDeployment.Spec.ClusterMetadata.infraID. Output shows: PASS: infraID matches. Platform field exists in retrofitted metadata.json. Output shows: PASS: platform field present

5. **Action:** Verify ClusterDeployment.Spec.ClusterMetadata.MetadataJSONSecretRef field is populated by retrofit.
   ```bash
   # Check MetadataJSONSecretRef populated
   oc get clusterdeployment legacy-test-cluster -o jsonpath='{.spec.clusterMetadata.metadataJSONSecretRef.name}'
   ```
   **Expected:** MetadataJSONSecretRef.name field shows: legacy-test-cluster-metadata-json

6. **Action:** Delete legacy ClusterDeployment to trigger deprovision using retrofitted metadata.json.
   ```bash
   # Delete ClusterDeployment
   oc delete clusterdeployment legacy-test-cluster

   # Verify ClusterDeprovision created
   sleep 5
   LEGACY_CD=$(oc get clusterdeprovision -l hive.openshift.io/cluster-deployment-name=legacy-test-cluster -o jsonpath='{.items[0].metadata.name}')
   echo "ClusterDeprovision: $LEGACY_CD"
   ```
   **Expected:** ClusterDeprovision resource created. ClusterDeprovision references retrofitted metadata.json Secret.

7. **Action:** Monitor deprovision completion using retrofitted metadata.json.
   ```bash
   # Wait for deprovision completion
   oc wait --for=condition=Completed clusterdeprovision/$LEGACY_CD --timeout=30m

   # Check completion status
   oc get clusterdeprovision $LEGACY_CD -o jsonpath='{.status.conditions[?(@.type=="Completed")].status}'
   ```
   **Expected:** ClusterDeprovision completes successfully. Completion condition status shows: True

8. **Action:** Verify no cloud resources remain after legacy cluster deprovision.
   ```bash
   # Verify resource cleanup (platform-specific)
   # Example for AWS:
   aws ec2 describe-instances --filters "Name=tag:kubernetes.io/cluster/${LEGACY_INFRA},Values=owned" --query 'Reservations[*].Instances[*].[InstanceId]' --output text | wc -l
   ```
   **Expected:** Zero resources remain with legacy cluster infraID. Retrofit and deprovision completed cleanly.

---

### Test Case HIVE-2302_M002
**Name:** AWS Shared VPC with HostedZoneRole Deprovision
**Description:** Verify that AWS clusters with HostedZoneRole (shared VPC DNS scenario) preserve HostedZoneRole in metadata.json Secret and deprovision DNS resources correctly.
**Type:** Manual
**Priority:** High

#### Prerequisites
- AWS shared VPC infrastructure pre-configured with cross-account hosted zone
- HostedZoneRole IAM role created in DNS account
- Hive management cluster with HIVE-2302 implementation
- Permissions to create clusters with HostedZoneRole configuration

#### Test Steps
1. **Action:** Create ClusterDeployment with AWS HostedZoneRole for shared VPC DNS.
   ```bash
   # Create ClusterDeployment with HostedZoneRole
   cat <<EOF | oc create -f -
   apiVersion: hive.openshift.io/v1
   kind: ClusterDeployment
   metadata:
     name: test-aws-sharedvpc
   spec:
     platform:
       aws:
         hostedZoneRole: arn:aws:iam::123456789012:role/cross-account-dns-role
   ...
   EOF

   # Wait for provisioning
   oc wait --for=condition=Provisioned clusterdeployment/test-aws-sharedvpc --timeout=60m
   ```
   **Expected:** ClusterDeployment completes provisioning with HostedZoneRole configuration. Cluster reaches "Provisioned" status.

2. **Action:** Verify metadata.json Secret contains HostedZoneRole field.
   ```bash
   # Extract metadata.json and check HostedZoneRole
   oc get secret test-aws-sharedvpc-metadata-json -o jsonpath='{.data.metadata\.json}' | base64 -d > /tmp/aws-sharedvpc-metadata.json

   # Verify HostedZoneRole preserved
   grep -q "hostedZoneRole" /tmp/aws-sharedvpc-metadata.json && echo "PASS: HostedZoneRole in metadata.json" || echo "FAIL: HostedZoneRole missing"

   # Check specific role ARN
   grep "cross-account-dns-role" /tmp/aws-sharedvpc-metadata.json
   ```
   **Expected:** metadata.json contains hostedZoneRole field. Output shows: PASS: HostedZoneRole in metadata.json. Correct IAM role ARN present in metadata.json.

3. **Action:** Delete ClusterDeployment to trigger deprovision of shared VPC cluster.
   ```bash
   # Delete cluster
   oc delete clusterdeployment test-aws-sharedvpc

   # Get ClusterDeprovision
   sleep 5
   SHAREDVPC_CD=$(oc get clusterdeprovision -l hive.openshift.io/cluster-deployment-name=test-aws-sharedvpc -o jsonpath='{.items[0].metadata.name}')
   ```
   **Expected:** ClusterDeprovision created with metadata.json Secret reference.

4. **Action:** Verify deprovision uses HostedZoneRole to clean DNS records in cross-account hosted zone.
   ```bash
   # Monitor deprovision
   oc wait --for=condition=Completed clusterdeprovision/$SHAREDVPC_CD --timeout=30m

   # Check deprovision pod logs for HostedZoneRole usage
   DEPROV_POD=$(oc get pods -l hive.openshift.io/cluster-deprovision=$SHAREDVPC_CD -o jsonpath='{.items[0].metadata.name}')
   oc logs $DEPROV_POD | grep -i "hostedZoneRole"
   ```
   **Expected:** ClusterDeprovision completes successfully. Deprovision logs show HostedZoneRole being used for DNS cleanup. Cross-account DNS records deleted.

5. **Action:** Verify shared VPC DNS hosted zone records are cleaned up.
   ```bash
   # Check hosted zone for remaining cluster DNS records
   HOSTED_ZONE_ID="Z1234567890ABC"  # Replace with actual hosted zone ID
   aws route53 list-resource-record-sets --hosted-zone-id $HOSTED_ZONE_ID | grep test-aws-sharedvpc | wc -l
   ```
   **Expected:** Zero DNS records remain for cluster in shared hosted zone. HostedZoneRole enabled correct cross-account cleanup.

---

### Test Case HIVE-2302_M003
**Name:** GCP Shared VPC with NetworkProjectID Deprovision
**Description:** Verify that GCP clusters with NetworkProjectID (shared VPC scenario) preserve NetworkProjectID in metadata.json Secret and deprovision network resources correctly.
**Type:** Manual
**Priority:** High

#### Prerequisites
- GCP shared VPC infrastructure with host project and service project
- NetworkProjectID configured for host project
- Hive management cluster with HIVE-2302 implementation
- GCP credentials with permissions to create clusters in shared VPC

#### Test Steps
1. **Action:** Create ClusterDeployment with GCP NetworkProjectID for shared VPC.
   ```bash
   # Create ClusterDeployment with NetworkProjectID
   cat <<EOF | oc create -f -
   apiVersion: hive.openshift.io/v1
   kind: ClusterDeployment
   metadata:
     name: test-gcp-sharedvpc
   spec:
     platform:
       gcp:
         networkProjectID: gcp-shared-vpc-host-project
   ...
   EOF

   # Wait for provisioning
   oc wait --for=condition=Provisioned clusterdeployment/test-gcp-sharedvpc --timeout=60m
   ```
   **Expected:** ClusterDeployment completes with NetworkProjectID configuration. Cluster provisions in GCP shared VPC.

2. **Action:** Verify metadata.json Secret contains NetworkProjectID field.
   ```bash
   # Extract and verify NetworkProjectID
   oc get secret test-gcp-sharedvpc-metadata-json -o jsonpath='{.data.metadata\.json}' | base64 -d > /tmp/gcp-sharedvpc-metadata.json

   grep -q "networkProjectID" /tmp/gcp-sharedvpc-metadata.json && echo "PASS: NetworkProjectID in metadata.json" || echo "FAIL: NetworkProjectID missing"

   # Verify project ID value
   grep "gcp-shared-vpc-host-project" /tmp/gcp-sharedvpc-metadata.json
   ```
   **Expected:** metadata.json contains networkProjectID field. Output shows: PASS: NetworkProjectID in metadata.json. Correct host project ID present.

3. **Action:** Delete ClusterDeployment to trigger deprovision of shared VPC cluster.
   ```bash
   # Delete cluster
   oc delete clusterdeployment test-gcp-sharedvpc

   # Get ClusterDeprovision
   sleep 5
   GCP_SHAREDVPC_CD=$(oc get clusterdeprovision -l hive.openshift.io/cluster-deployment-name=test-gcp-sharedvpc -o jsonpath='{.items[0].metadata.name}')
   ```
   **Expected:** ClusterDeprovision created.

4. **Action:** Monitor deprovision and verify NetworkProjectID used for resource cleanup.
   ```bash
   # Wait for completion
   oc wait --for=condition=Completed clusterdeprovision/$GCP_SHAREDVPC_CD --timeout=30m

   # Verify completion
   oc get clusterdeprovision $GCP_SHAREDVPC_CD -o jsonpath='{.status.conditions[?(@.type=="Completed")].status}'
   ```
   **Expected:** ClusterDeprovision completes successfully. Status shows: True

5. **Action:** Verify GCP shared VPC resources cleaned up correctly.
   ```bash
   # Get cluster infraID
   GCP_INFRA=$(oc get secret test-gcp-sharedvpc-metadata-json -o jsonpath='{.data.metadata\.json}' | base64 -d | grep -o '"infraID":"[^"]*' | cut -d'"' -f4)

   # Check for remaining instances in shared VPC
   gcloud compute instances list --project=gcp-shared-vpc-host-project --filter="labels.kubernetes-io-cluster-${GCP_INFRA}=owned" --format="value(name)" | wc -l
   ```
   **Expected:** Zero instances remain in shared VPC host project. NetworkProjectID enabled correct shared VPC cleanup.

---

### Test Case HIVE-2302_M004
**Name:** Azure Existing Resource Group Deprovision
**Description:** Verify that Azure clusters with ResourceGroupName (existing resource group scenario) preserve ResourceGroupName in metadata.json Secret and deprovision resources correctly without deleting the pre-existing resource group.
**Type:** Manual
**Priority:** High

#### Prerequisites
- Pre-created Azure resource group for cluster deployment
- ResourceGroupName configured in ClusterDeployment
- Hive management cluster with HIVE-2302 implementation
- Azure credentials with permissions to use existing resource groups

#### Test Steps
1. **Action:** Pre-create Azure resource group for testing.
   ```bash
   # Create resource group
   az group create --name hive-existing-rg-test --location eastus

   # Verify resource group exists
   az group show --name hive-existing-rg-test --query "name" -o tsv
   ```
   **Expected:** Resource group hive-existing-rg-test created successfully. Name shows: hive-existing-rg-test

2. **Action:** Create ClusterDeployment with Azure ResourceGroupName pointing to existing resource group.
   ```bash
   # Create ClusterDeployment
   cat <<EOF | oc create -f -
   apiVersion: hive.openshift.io/v1
   kind: ClusterDeployment
   metadata:
     name: test-azure-existingrg
   spec:
     platform:
       azure:
         resourceGroupName: hive-existing-rg-test
   ...
   EOF

   # Wait for provisioning
   oc wait --for=condition=Provisioned clusterdeployment/test-azure-existingrg --timeout=60m
   ```
   **Expected:** ClusterDeployment provisions into existing resource group. Cluster reaches "Provisioned" status.

3. **Action:** Verify metadata.json Secret contains ResourceGroupName field.
   ```bash
   # Extract metadata.json
   oc get secret test-azure-existingrg-metadata-json -o jsonpath='{.data.metadata\.json}' | base64 -d > /tmp/azure-existingrg-metadata.json

   # Verify ResourceGroupName
   grep -q "resourceGroupName" /tmp/azure-existingrg-metadata.json && echo "PASS: ResourceGroupName in metadata.json" || echo "FAIL: ResourceGroupName missing"

   # Check specific resource group name
   grep "hive-existing-rg-test" /tmp/azure-existingrg-metadata.json
   ```
   **Expected:** metadata.json contains resourceGroupName field. Output shows: PASS: ResourceGroupName in metadata.json. Correct resource group name present.

4. **Action:** Delete ClusterDeployment to trigger deprovision.
   ```bash
   # Delete cluster
   oc delete clusterdeployment test-azure-existingrg

   # Get ClusterDeprovision
   sleep 5
   AZURE_EXISTINGRG_CD=$(oc get clusterdeprovision -l hive.openshift.io/cluster-deployment-name=test-azure-existingrg -o jsonpath='{.items[0].metadata.name}')
   ```
   **Expected:** ClusterDeprovision created.

5. **Action:** Monitor deprovision completion.
   ```bash
   # Wait for completion
   oc wait --for=condition=Completed clusterdeprovision/$AZURE_EXISTINGRG_CD --timeout=30m
   ```
   **Expected:** ClusterDeprovision completes successfully.

6. **Action:** Verify cluster resources deleted but existing resource group PRESERVED.
   ```bash
   # Check resource group still exists
   az group show --name hive-existing-rg-test --query "name" -o tsv

   # Check resource group is empty or only contains non-cluster resources
   az resource list --resource-group hive-existing-rg-test --query "length(@)"
   ```
   **Expected:** Resource group hive-existing-rg-test still exists (NOT deleted). Resource group is empty or contains only pre-existing non-cluster resources. Cluster resources cleaned up without deleting the resource group itself.

---

### Test Case HIVE-2302_M005
**Name:** vSphere Pre-Zonal and Post-Zonal Metadata Format Retrofit and Deprovision
**Description:** Verify that vSphere clusters created both before and after zonal support (different metadata.json schemas) are retrofitted correctly and deprovision successfully.
**Type:** Manual
**Priority:** Medium

#### Prerequisites
- Access to vSphere environments
- Ability to deploy specific OpenShift versions (pre-zonal and post-zonal)
- Hive management cluster with HIVE-2302 implementation
- vSphere credentials for both scenarios

#### Test Steps
1. **Action:** Create or identify vSphere cluster created BEFORE zonal support (pre-zonal metadata format).
   ```bash
   # Verify pre-zonal cluster exists
   oc get clusterdeployment vsphere-pre-zonal-cluster

   # Check metadata format (pre-zonal has single vCenter config)
   oc get secret vsphere-pre-zonal-cluster-metadata-json -o jsonpath='{.data.metadata\.json}' | base64 -d | grep -c "vcenters" || echo "Pre-zonal format (no vcenters array)"
   ```
   **Expected:** Pre-zonal cluster exists. metadata.json uses old format without vcenters array.

2. **Action:** Verify pre-zonal metadata.json Secret contains credentials in old format.
   ```bash
   # Extract metadata.json
   oc get secret vsphere-pre-zonal-cluster-metadata-json -o jsonpath='{.data.metadata\.json}' | base64 -d > /tmp/vsphere-pre-zonal-metadata.json

   # Check for username/password fields at root level
   grep -q '"username"' /tmp/vsphere-pre-zonal-metadata.json && echo "PASS: username field present (pre-zonal)" || echo "FAIL: username missing"
   ```
   **Expected:** metadata.json contains username and password fields at VSphere root level. Output shows: PASS: username field present (pre-zonal)

3. **Action:** Delete pre-zonal vSphere cluster to trigger deprovision.
   ```bash
   # Delete cluster
   oc delete clusterdeployment vsphere-pre-zonal-cluster

   # Get ClusterDeprovision
   sleep 5
   VSPHERE_PREZONAL_CD=$(oc get clusterdeprovision -l hive.openshift.io/cluster-deployment-name=vsphere-pre-zonal-cluster -o jsonpath='{.items[0].metadata.name}')
   ```
   **Expected:** ClusterDeprovision created for pre-zonal cluster.

4. **Action:** Monitor pre-zonal cluster deprovision with credential re-injection.
   ```bash
   # Wait for completion
   oc wait --for=condition=Completed clusterdeprovision/$VSPHERE_PREZONAL_CD --timeout=30m

   # Check deprovision pod logs for credential injection
   PREZONAL_POD=$(oc get pods -l hive.openshift.io/cluster-deprovision=$VSPHERE_PREZONAL_CD -o jsonpath='{.items[0].metadata.name}')
   oc logs $PREZONAL_POD | grep -i "username\|password" | head -5
   ```
   **Expected:** ClusterDeprovision completes successfully. Logs show credentials re-injected from environment variables.

5. **Action:** Create or identify vSphere cluster created AFTER zonal support (post-zonal metadata format).
   ```bash
   # Verify post-zonal cluster exists
   oc get clusterdeployment vsphere-post-zonal-cluster

   # Check metadata format (post-zonal has vcenters array)
   oc get secret vsphere-post-zonal-cluster-metadata-json -o jsonpath='{.data.metadata\.json}' | base64 -d | grep -q "vcenters" && echo "PASS: Post-zonal format (vcenters array)" || echo "FAIL: vcenters array missing"
   ```
   **Expected:** Post-zonal cluster exists. metadata.json uses new format with vcenters array. Output shows: PASS: Post-zonal format (vcenters array)

6. **Action:** Verify post-zonal metadata.json Secret contains credentials in new format (vcenters array).
   ```bash
   # Extract metadata.json
   oc get secret vsphere-post-zonal-cluster-metadata-json -o jsonpath='{.data.metadata\.json}' | base64 -d > /tmp/vsphere-post-zonal-metadata.json

   # Check for vcenters array with username/password
   grep -A 5 '"vcenters"' /tmp/vsphere-post-zonal-metadata.json | grep -q '"username"' && echo "PASS: username in vcenters array (post-zonal)" || echo "FAIL: username missing in vcenters"
   ```
   **Expected:** metadata.json contains vcenters array with username/password fields. Output shows: PASS: username in vcenters array (post-zonal)

7. **Action:** Delete post-zonal vSphere cluster to trigger deprovision.
   ```bash
   # Delete cluster
   oc delete clusterdeployment vsphere-post-zonal-cluster

   # Get ClusterDeprovision
   sleep 5
   VSPHERE_POSTZONAL_CD=$(oc get clusterdeprovision -l hive.openshift.io/cluster-deployment-name=vsphere-post-zonal-cluster -o jsonpath='{.items[0].metadata.name}')
   ```
   **Expected:** ClusterDeprovision created for post-zonal cluster.

8. **Action:** Monitor post-zonal cluster deprovision with credential re-injection for vcenters array.
   ```bash
   # Wait for completion
   oc wait --for=condition=Completed clusterdeprovision/$VSPHERE_POSTZONAL_CD --timeout=30m

   # Verify completion
   oc get clusterdeprovision $VSPHERE_POSTZONAL_CD -o jsonpath='{.status.conditions[?(@.type=="Completed")].status}'
   ```
   **Expected:** ClusterDeprovision completes successfully. Status shows: True. Credentials re-injected correctly for vcenters array format.

9. **Action:** Verify both pre-zonal and post-zonal vSphere clusters deprovisioned cleanly without resource leaks.
   ```bash
   # Check for remaining VMs (both clusters should be clean)
   # This requires govc configured for vSphere access

   # Verify both cluster folders are cleaned up
   echo "Verifying both vSphere cluster formats deprovisioned successfully"
   ```
   **Expected:** Both pre-zonal and post-zonal vSphere clusters deprovisioned cleanly. Metadata format differences handled correctly by generic destroyer.
