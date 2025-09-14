# Test Case: HIVE-2302
**Component:** hive
**Summary:** Pass installer's metadata.json directly to destroyer

## Test Overview
- **Total Test Cases:** 2
- **Test Types:** E2E
- **Estimated Time:** 120 minutes

## Test Cases

### Test Case HIVE-2302_E001
**Name:** Modern Cluster Provision with metadata.json Secret Creation and Deprovision
**Description:** Verify that clusters created after HIVE-2302 implementation automatically generate metadata.json Secret and successfully deprovision using the new generic destroyer.
**Type:** E2E
**Priority:** Critical

#### Prerequisites
- Hive management cluster is running with HIVE-2302 changes deployed
- Cloud provider credentials are configured (AWS/Azure/GCP/vSphere)
- Access to Hive namespace and permissions to create ClusterDeployment

#### Test Steps
1. **Action:** Create a ClusterDeployment on AWS platform.
   ```bash
   # Create standard ClusterDeployment
   oc create -f clusterdeployment-aws.yaml

   # Wait for cluster installation to complete
   oc wait --for=condition=Provisioned clusterdeployment/test-cluster-aws --timeout=60m

   # Verify cluster reaches Running state
   oc get clusterdeployment test-cluster-aws -o jsonpath='{.status.powerState}'
   ```
   **Expected:** ClusterDeployment installation completes successfully. Cluster reaches "Provisioned" condition. powerState shows: Running

2. **Action:** Verify metadata.json Secret is created automatically after provision.
   ```bash
   # Get cluster name
   CLUSTER_NAME=$(oc get clusterdeployment test-cluster-aws -o jsonpath='{.metadata.name}')

   # Check for metadata.json Secret
   oc get secret ${CLUSTER_NAME}-metadata-json -n test-namespace

   # Verify Secret has metadata.json key
   oc get secret ${CLUSTER_NAME}-metadata-json -n test-namespace -o jsonpath='{.data.metadata\.json}' | base64 -d | head -10
   ```
   **Expected:** Secret ${CLUSTER_NAME}-metadata-json exists in cluster namespace. Secret contains key "metadata.json". Decoded content is valid JSON with infraID and platform fields visible.

3. **Action:** Verify ClusterDeployment.Spec.ClusterMetadata.MetadataJSONSecretRef field is populated.
   ```bash
   # Check MetadataJSONSecretRef field
   oc get clusterdeployment test-cluster-aws -o jsonpath='{.spec.clusterMetadata.metadataJSONSecretRef.name}'
   ```
   **Expected:** MetadataJSONSecretRef.name field shows: ${CLUSTER_NAME}-metadata-json

4. **Action:** Verify metadata.json Secret contains installer's metadata verbatim.
   ```bash
   # Extract metadata.json from Secret
   oc get secret ${CLUSTER_NAME}-metadata-json -n test-namespace -o jsonpath='{.data.metadata\.json}' | base64 -d > /tmp/metadata.json

   # Verify infraID field exists
   INFRA_ID=$(oc get clusterdeployment test-cluster-aws -o jsonpath='{.spec.clusterMetadata.infraID}')
   grep -q "$INFRA_ID" /tmp/metadata.json && echo "PASS: infraID present" || echo "FAIL: infraID missing"

   # Verify platform field exists
   grep -q '"platform"' /tmp/metadata.json && echo "PASS: platform field present" || echo "FAIL: platform field missing"
   ```
   **Expected:** infraID from ClusterDeployment matches infraID in metadata.json. Output shows: PASS: infraID present. Platform field exists in metadata.json. Output shows: PASS: platform field present

5. **Action:** Delete ClusterDeployment to trigger deprovision.
   ```bash
   # Delete ClusterDeployment
   oc delete clusterdeployment test-cluster-aws

   # Verify ClusterDeprovision is created
   sleep 5
   oc get clusterdeprovision -l hive.openshift.io/cluster-deployment-name=test-cluster-aws
   ```
   **Expected:** ClusterDeprovision resource is created automatically. ClusterDeprovision name matches pattern related to cluster name.

6. **Action:** Verify ClusterDeprovision references metadata.json Secret.
   ```bash
   # Get ClusterDeprovision name
   CD_NAME=$(oc get clusterdeprovision -l hive.openshift.io/cluster-deployment-name=test-cluster-aws -o jsonpath='{.items[0].metadata.name}')

   # Check MetadataJSONSecretRef in ClusterDeprovision
   oc get clusterdeprovision $CD_NAME -o jsonpath='{.spec.metadataJSONSecretRef.name}'
   ```
   **Expected:** ClusterDeprovision.Spec.MetadataJSONSecretRef.name shows: ${CLUSTER_NAME}-metadata-json

7. **Action:** Monitor ClusterDeprovision completion.
   ```bash
   # Wait for ClusterDeprovision to complete
   oc wait --for=condition=Completed clusterdeprovision/$CD_NAME --timeout=30m

   # Verify completion status
   oc get clusterdeprovision $CD_NAME -o jsonpath='{.status.conditions[?(@.type=="Completed")].status}'
   ```
   **Expected:** ClusterDeprovision completes within 30 minutes. Completion condition status shows: True

8. **Action:** Verify generic destroyer (hiveutil deprovision) was used instead of legacy platform-specific destroyer.
   ```bash
   # Check deprovision pod logs for generic destroyer usage
   DEPROV_POD=$(oc get pods -n test-namespace -l hive.openshift.io/cluster-deprovision=$CD_NAME -o jsonpath='{.items[0].metadata.name}')

   # Verify hiveutil deprovision command used
   oc logs -n test-namespace $DEPROV_POD | grep -i "metadata-json-secret"
   ```
   **Expected:** Deprovision pod logs show --metadata-json-secret-name parameter usage. Confirms generic destroyer code path executed.

9. **Action:** Verify no cloud resources remain (leak detection).
   ```bash
   # Get infraID for resource tagging
   INFRA_ID=$(oc get clusterdeployment test-cluster-aws -o jsonpath='{.spec.clusterMetadata.infraID}' 2>/dev/null || grep infraID /tmp/metadata.json | awk -F'"' '{print $4}')

   # Check AWS resources with infraID tag (AWS example)
   aws ec2 describe-instances --filters "Name=tag:kubernetes.io/cluster/${INFRA_ID},Values=owned" --query 'Reservations[*].Instances[*].[InstanceId]' --output text | wc -l
   ```
   **Expected:** Zero resources with cluster infraID tag remain. Output shows: 0

---

### Test Case HIVE-2302_E002
**Name:** Deprovision Resource Leak Detection Across All Platforms
**Description:** Comprehensive regression test to verify deprovision completes cleanly without resource leaks on AWS, Azure, GCP, and vSphere platforms using the new metadata.json-based destroyer.
**Type:** E2E
**Priority:** Critical

#### Prerequisites
- Hive management cluster is running with HIVE-2302 changes
- Cloud credentials configured for multiple platforms: AWS, Azure, GCP, vSphere
- CLI tools installed: aws, az, gcloud, govc
- Access to cloud provider accounts for resource verification

#### Test Steps
1. **Action:** Create ClusterDeployments on all supported platforms.
   ```bash
   # Create AWS cluster
   oc create -f clusterdeployment-aws-leak-test.yaml

   # Create Azure cluster
   oc create -f clusterdeployment-azure-leak-test.yaml

   # Create GCP cluster
   oc create -f clusterdeployment-gcp-leak-test.yaml

   # Create vSphere cluster
   oc create -f clusterdeployment-vsphere-leak-test.yaml

   # Wait for all clusters to complete provisioning
   oc wait --for=condition=Provisioned clusterdeployment/test-leak-aws --timeout=60m
   oc wait --for=condition=Provisioned clusterdeployment/test-leak-azure --timeout=60m
   oc wait --for=condition=Provisioned clusterdeployment/test-leak-gcp --timeout=60m
   oc wait --for=condition=Provisioned clusterdeployment/test-leak-vsphere --timeout=60m
   ```
   **Expected:** All ClusterDeployments complete provisioning successfully. All clusters reach "Provisioned" condition. All clusters show powerState: Running

2. **Action:** Capture infraID and metadata.json Secret names for each platform.
   ```bash
   # AWS
   AWS_INFRA=$(oc get clusterdeployment test-leak-aws -o jsonpath='{.spec.clusterMetadata.infraID}')
   AWS_SECRET=$(oc get clusterdeployment test-leak-aws -o jsonpath='{.spec.clusterMetadata.metadataJSONSecretRef.name}')

   # Azure
   AZURE_INFRA=$(oc get clusterdeployment test-leak-azure -o jsonpath='{.spec.clusterMetadata.infraID}')
   AZURE_SECRET=$(oc get clusterdeployment test-leak-azure -o jsonpath='{.spec.clusterMetadata.metadataJSONSecretRef.name}')

   # GCP
   GCP_INFRA=$(oc get clusterdeployment test-leak-gcp -o jsonpath='{.spec.clusterMetadata.infraID}')
   GCP_SECRET=$(oc get clusterdeployment test-leak-gcp -o jsonpath='{.spec.clusterMetadata.metadataJSONSecretRef.name}')

   # vSphere
   VSPHERE_INFRA=$(oc get clusterdeployment test-leak-vsphere -o jsonpath='{.spec.clusterMetadata.infraID}')
   VSPHERE_SECRET=$(oc get clusterdeployment test-leak-vsphere -o jsonpath='{.spec.clusterMetadata.metadataJSONSecretRef.name}')

   echo "AWS: $AWS_INFRA, Secret: $AWS_SECRET"
   echo "Azure: $AZURE_INFRA, Secret: $AZURE_SECRET"
   echo "GCP: $GCP_INFRA, Secret: $GCP_SECRET"
   echo "vSphere: $VSPHERE_INFRA, Secret: $VSPHERE_SECRET"
   ```
   **Expected:** All infraID values are non-empty. All metadata.json Secret names are populated. Output shows 4 lines with infraID and Secret names.

3. **Action:** Delete all ClusterDeployments simultaneously to trigger deprovision.
   ```bash
   # Delete all clusters
   oc delete clusterdeployment test-leak-aws test-leak-azure test-leak-gcp test-leak-vsphere

   # Verify ClusterDeprovisions created
   sleep 10
   oc get clusterdeprovision
   ```
   **Expected:** 4 ClusterDeprovision resources are created (one per platform). ClusterDeprovisions reference respective metadata.json Secrets.

4. **Action:** Monitor all ClusterDeprovisions for completion.
   ```bash
   # Get ClusterDeprovision names
   AWS_CD=$(oc get clusterdeprovision -l hive.openshift.io/cluster-deployment-name=test-leak-aws -o jsonpath='{.items[0].metadata.name}')
   AZURE_CD=$(oc get clusterdeprovision -l hive.openshift.io/cluster-deployment-name=test-leak-azure -o jsonpath='{.items[0].metadata.name}')
   GCP_CD=$(oc get clusterdeprovision -l hive.openshift.io/cluster-deployment-name=test-leak-gcp -o jsonpath='{.items[0].metadata.name}')
   VSPHERE_CD=$(oc get clusterdeprovision -l hive.openshift.io/cluster-deployment-name=test-leak-vsphere -o jsonpath='{.items[0].metadata.name}')

   # Wait for completions
   oc wait --for=condition=Completed clusterdeprovision/$AWS_CD --timeout=30m &
   oc wait --for=condition=Completed clusterdeprovision/$AZURE_CD --timeout=30m &
   oc wait --for=condition=Completed clusterdeprovision/$GCP_CD --timeout=30m &
   oc wait --for=condition=Completed clusterdeprovision/$VSPHERE_CD --timeout=30m &
   wait
   ```
   **Expected:** All 4 ClusterDeprovisions complete successfully. All reach Completed condition within 30 minutes.

5. **Action:** Verify AWS resource cleanup (zero leaks).
   ```bash
   # Check EC2 instances
   AWS_INSTANCES=$(aws ec2 describe-instances --filters "Name=tag:kubernetes.io/cluster/${AWS_INFRA},Values=owned" --query 'Reservations[*].Instances[*].[InstanceId]' --output text | wc -l)

   # Check VPCs
   AWS_VPCS=$(aws ec2 describe-vpcs --filters "Name=tag:kubernetes.io/cluster/${AWS_INFRA},Values=owned" --query 'Vpcs[*].[VpcId]' --output text | wc -l)

   # Check S3 buckets
   AWS_BUCKETS=$(aws s3api list-buckets --query "Buckets[?contains(Name, '${AWS_INFRA}')].Name" --output text | wc -l)

   echo "AWS leaked resources: Instances=$AWS_INSTANCES, VPCs=$AWS_VPCS, Buckets=$AWS_BUCKETS"
   test $AWS_INSTANCES -eq 0 && test $AWS_VPCS -eq 0 && test $AWS_BUCKETS -eq 0 && echo "PASS: AWS clean" || echo "FAIL: AWS leaks detected"
   ```
   **Expected:** Zero instances, VPCs, and S3 buckets remain. Output shows: AWS leaked resources: Instances=0, VPCs=0, Buckets=0. Output shows: PASS: AWS clean

6. **Action:** Verify Azure resource cleanup (zero leaks).
   ```bash
   # Check resource group
   AZURE_RG=$(oc get secret ${AZURE_SECRET} -o jsonpath='{.data.metadata\.json}' | base64 -d | grep -o '"resourceGroupName":"[^"]*' | cut -d'"' -f4)

   # Verify resource group deleted
   az group show --name $AZURE_RG 2>&1 | grep -q "ResourceGroupNotFound" && echo "PASS: Azure resource group deleted" || echo "FAIL: Azure resource group still exists"
   ```
   **Expected:** Azure resource group is deleted. Output shows: PASS: Azure resource group deleted

7. **Action:** Verify GCP resource cleanup (zero leaks).
   ```bash
   # Check GCP instances
   GCP_PROJECT=$(oc get secret ${GCP_SECRET} -o jsonpath='{.data.metadata\.json}' | base64 -d | grep -o '"projectID":"[^"]*' | cut -d'"' -f4)

   GCP_INSTANCES=$(gcloud compute instances list --project=$GCP_PROJECT --filter="labels.kubernetes-io-cluster-${GCP_INFRA}=owned" --format="value(name)" | wc -l)

   echo "GCP leaked instances: $GCP_INSTANCES"
   test $GCP_INSTANCES -eq 0 && echo "PASS: GCP clean" || echo "FAIL: GCP leaks detected"
   ```
   **Expected:** Zero GCP instances remain with cluster infraID label. Output shows: GCP leaked instances: 0. Output shows: PASS: GCP clean

8. **Action:** Verify vSphere resource cleanup (zero leaks).
   ```bash
   # Extract vSphere folder from metadata.json
   VSPHERE_FOLDER=$(oc get secret ${VSPHERE_SECRET} -o jsonpath='{.data.metadata\.json}' | base64 -d | grep -o '"folder":"[^"]*' | cut -d'"' -f4)

   # Check for VMs in cluster folder (requires govc configured)
   VMS_COUNT=$(govc find -type VirtualMachine "/${VSPHERE_FOLDER}" 2>/dev/null | wc -l)

   echo "vSphere leaked VMs: $VMS_COUNT"
   test $VMS_COUNT -eq 0 && echo "PASS: vSphere clean" || echo "FAIL: vSphere leaks detected"
   ```
   **Expected:** Zero VMs remain in vSphere cluster folder. Output shows: vSphere leaked VMs: 0. Output shows: PASS: vSphere clean

9. **Action:** Generate comprehensive leak detection summary report.
   ```bash
   # Summary of all platforms
   echo "=== Deprovision Resource Leak Report ==="
   echo "AWS: Instances=$AWS_INSTANCES, VPCs=$AWS_VPCS, Buckets=$AWS_BUCKETS"
   echo "Azure: Resource Group Deleted"
   echo "GCP: Instances=$GCP_INSTANCES"
   echo "vSphere: VMs=$VMS_COUNT"

   # Overall pass/fail
   TOTAL_LEAKS=$((AWS_INSTANCES + AWS_VPCS + AWS_BUCKETS + GCP_INSTANCES + VMS_COUNT))
   test $TOTAL_LEAKS -eq 0 && echo "OVERALL: PASS - Zero leaks across all platforms" || echo "OVERALL: FAIL - $TOTAL_LEAKS leaks detected"
   ```
   **Expected:** Total leak count is 0. All platforms show zero remaining resources. Output shows: OVERALL: PASS - Zero leaks across all platforms
