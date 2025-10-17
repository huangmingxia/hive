# Test Case: HIVE-2302
**Component:** hive
**Summary:** Pass installer's metadata.json directly to destroyer

## Test Overview
- **Total Test Cases:** 6
- **Test Types:** E2E
- **Estimated Time:** 180-240 minutes

## Test Cases

### Test Case HIVE-2302_001
**Name:** AWS Modern Cluster - metadata.json Secret Creation and Generic Deprovision
**Description:** Validate automatic metadata.json Secret creation during AWS cluster provisioning and successful deprovision using generic destroyer
**Type:** E2E
**Priority:** High

#### Prerequisites
- Access to AWS cloud credentials with permission to create resources
- Hive operator deployed with HIVE-2302 changes
- Base domain configured for cluster DNS

#### Test Steps
1. **Action:** Create AWS ClusterDeployment
   ```bash
   oc create namespace hive-test-aws-modern

   oc create secret generic aws-creds \
     --from-literal=aws_access_key_id=$AWS_ACCESS_KEY_ID \
     --from-literal=aws_secret_access_key=$AWS_SECRET_ACCESS_KEY \
     -n hive-test-aws-modern

   cat <<EOF | oc apply -f -
   apiVersion: hive.openshift.io/v1
   kind: ClusterDeployment
   metadata:
     name: aws-modern-test
     namespace: hive-test-aws-modern
   spec:
     baseDomain: example.com
     clusterName: aws-modern-test
     platform:
       aws:
         credentialsSecretRef:
           name: aws-creds
         region: us-east-1
     provisioning:
       installConfigSecretRef:
         name: install-config
       imageSetRef:
         name: openshift-v4.16.0
     pullSecretRef:
       name: pull-secret
   EOF
   ```
   **Expected:** ClusterDeployment created successfully

2. **Action:** Wait for cluster installation and verify Secret creation
   ```bash
   oc wait --for=condition=Installed=true clusterdeployment/aws-modern-test \
     -n hive-test-aws-modern --timeout=60m

   SECRET_NAME=$(oc get clusterdeployment aws-modern-test -n hive-test-aws-modern \
     -o jsonpath='{.spec.clusterMetadata.metadataJSONSecretRef.name}')

   echo "Metadata JSON Secret name: $SECRET_NAME"
   ```
   **Expected:** Cluster installs successfully within 60 minutes, MetadataJSONSecretRef populated with Secret name

3. **Action:** Verify metadata.json Secret exists and contains valid content
   ```bash
   oc get secret $SECRET_NAME -n hive-test-aws-modern

   oc get secret $SECRET_NAME -n hive-test-aws-modern \
     -o jsonpath='{.data.metadata\.json}' | base64 -d > /tmp/metadata.json

   cat /tmp/metadata.json | jq .

   cat /tmp/metadata.json | jq -r '.infraID'
   ```
   **Expected:** Secret exists, contains metadata.json key with valid JSON, infraID field present

4. **Action:** Verify Secret has ownerReference to ClusterDeployment
   ```bash
   oc get secret $SECRET_NAME -n hive-test-aws-modern \
     -o jsonpath='{.metadata.ownerReferences[0].kind}'

   oc get secret $SECRET_NAME -n hive-test-aws-modern \
     -o jsonpath='{.metadata.ownerReferences[0].name}'
   ```
   **Expected:** ownerReference kind is ClusterDeployment, name is aws-modern-test

5. **Action:** Extract infraID and count AWS resources before deprovision
   ```bash
   INFRA_ID=$(cat /tmp/metadata.json | jq -r '.infraID')
   echo "InfraID: $INFRA_ID"

   aws resourcegroupstaggingapi get-resources \
     --tag-filters "Key=kubernetes.io/cluster/${INFRA_ID},Values=owned" \
     --region us-east-1 \
     --query 'ResourceTagMappingList[*].ResourceARN' \
     --output json > /tmp/resources-before.json

   RESOURCE_COUNT_BEFORE=$(cat /tmp/resources-before.json | jq '. | length')
   echo "Resources before deprovision: $RESOURCE_COUNT_BEFORE"
   ```
   **Expected:** Multiple AWS resources tagged with cluster infraID (typically 50-100+)

6. **Action:** Delete ClusterDeployment to trigger deprovision
   ```bash
   oc delete clusterdeployment aws-modern-test -n hive-test-aws-modern

   sleep 30

   DEPROVISION_NAME=$(oc get clusterdeprovision -n hive-test-aws-modern \
     -o jsonpath='{.items[0].metadata.name}')

   echo "ClusterDeprovision: $DEPROVISION_NAME"
   ```
   **Expected:** ClusterDeployment deletion initiated, ClusterDeprovision resource created

7. **Action:** Verify deprovision job uses generic destroyer with metadata.json Secret
   ```bash
   DEPROVISION_POD=$(oc get pods -n hive-test-aws-modern \
     -l hive.openshift.io/cluster-deprovision-name=$DEPROVISION_NAME \
     -o jsonpath='{.items[0].metadata.name}')

   echo "Deprovision pod: $DEPROVISION_POD"

   oc get pod $DEPROVISION_POD -n hive-test-aws-modern \
     -o jsonpath='{.spec.volumes[*].name}' | grep -o "metadata-json"

   oc get pod $DEPROVISION_POD -n hive-test-aws-modern \
     -o jsonpath='{.spec.containers[0].command[*]}' | grep "deprovision"
   ```
   **Expected:** Pod has volume named metadata-json, command includes "hiveutil deprovision --metadata-json-secret-name"

8. **Action:** Monitor deprovision logs for generic destroyer usage
   ```bash
   oc logs $DEPROVISION_POD -n hive-test-aws-modern -f --tail=50 | head -100

   oc logs $DEPROVISION_POD -n hive-test-aws-modern | \
     grep -i "metadata-json-secret-name"
   ```
   **Expected:** Logs show metadata.json Secret mounting and unmarshaling, generic destroyer initialization

9. **Action:** Wait for deprovision completion
   ```bash
   oc wait --for=condition=Completed=true \
     clusterdeprovision/$DEPROVISION_NAME \
     -n hive-test-aws-modern --timeout=30m
   ```
   **Expected:** ClusterDeprovision completes successfully within 30 minutes

10. **Action:** Verify all AWS resources cleaned up
    ```bash
    sleep 60

    aws resourcegroupstaggingapi get-resources \
      --tag-filters "Key=kubernetes.io/cluster/${INFRA_ID},Values=owned" \
      --region us-east-1 \
      --query 'ResourceTagMappingList[*].ResourceARN' \
      --output json > /tmp/resources-after.json

    RESOURCE_COUNT_AFTER=$(cat /tmp/resources-after.json | jq '. | length')
    echo "Resources after deprovision: $RESOURCE_COUNT_AFTER"
    ```
    **Expected:** 0 AWS resources remaining with cluster infraID tag

### Test Case HIVE-2302_002
**Name:** Azure Modern Cluster - metadata.json Secret and Generic Deprovision
**Description:** Validate Azure cluster provisioning creates metadata.json Secret and deprovision uses generic destroyer
**Type:** E2E
**Priority:** High

#### Prerequisites
- Access to Azure cloud credentials
- Hive operator deployed with HIVE-2302 changes
- Base domain configured

#### Test Steps
1. **Action:** Create Azure ClusterDeployment
   ```bash
   oc create namespace hive-test-azure-modern

   oc create secret generic azure-creds \
     --from-literal=osServicePrincipal.json='{"clientId":"xxx","clientSecret":"xxx","tenantId":"xxx","subscriptionId":"xxx"}' \
     -n hive-test-azure-modern

   cat <<EOF | oc apply -f -
   apiVersion: hive.openshift.io/v1
   kind: ClusterDeployment
   metadata:
     name: azure-modern-test
     namespace: hive-test-azure-modern
   spec:
     baseDomain: example.com
     clusterName: azure-modern-test
     platform:
       azure:
         baseDomainResourceGroupName: azure-dns-rg
         region: eastus
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
   **Expected:** ClusterDeployment created successfully

2. **Action:** Wait for installation and verify metadata.json Secret creation
   ```bash
   oc wait --for=condition=Installed=true clusterdeployment/azure-modern-test \
     -n hive-test-azure-modern --timeout=60m

   SECRET_NAME=$(oc get clusterdeployment azure-modern-test -n hive-test-azure-modern \
     -o jsonpath='{.spec.clusterMetadata.metadataJSONSecretRef.name}')

   oc get secret $SECRET_NAME -n hive-test-azure-modern \
     -o jsonpath='{.data.metadata\.json}' | base64 -d | jq .
   ```
   **Expected:** Installation completes, Secret created with valid Azure metadata.json

3. **Action:** Extract infraID and resourceGroupName
   ```bash
   METADATA_JSON=$(oc get secret $SECRET_NAME -n hive-test-azure-modern \
     -o jsonpath='{.data.metadata\.json}' | base64 -d)

   INFRA_ID=$(echo "$METADATA_JSON" | jq -r '.infraID')
   RESOURCE_GROUP=$(echo "$METADATA_JSON" | jq -r '.azure.resourceGroupName')

   echo "InfraID: $INFRA_ID"
   echo "Resource Group: $RESOURCE_GROUP"
   ```
   **Expected:** infraID and resourceGroupName extracted successfully

4. **Action:** Count Azure resources before deprovision
   ```bash
   az resource list --resource-group $RESOURCE_GROUP \
     --query '[].{name:name, type:type}' > /tmp/azure-resources-before.json

   RESOURCE_COUNT=$(cat /tmp/azure-resources-before.json | jq '. | length')
   echo "Azure resources before deprovision: $RESOURCE_COUNT"
   ```
   **Expected:** Multiple resources in resource group (VMs, NICs, disks, etc.)

5. **Action:** Delete ClusterDeployment and verify deprovision job configuration
   ```bash
   oc delete clusterdeployment azure-modern-test -n hive-test-azure-modern

   sleep 30

   DEPROVISION_POD=$(oc get pods -n hive-test-azure-modern \
     -l hive.openshift.io/cluster-deprovision-name \
     -o jsonpath='{.items[0].metadata.name}')

   oc get pod $DEPROVISION_POD -n hive-test-azure-modern \
     -o jsonpath='{.spec.containers[0].command[*]}' | grep "deprovision --metadata-json-secret-name"
   ```
   **Expected:** Deprovision job uses generic destroyer command with metadata-json-secret-name parameter

6. **Action:** Monitor deprovision completion
   ```bash
   DEPROVISION_NAME=$(oc get clusterdeprovision -n hive-test-azure-modern \
     -o jsonpath='{.items[0].metadata.name}')

   oc wait --for=condition=Completed=true \
     clusterdeprovision/$DEPROVISION_NAME \
     -n hive-test-azure-modern --timeout=30m
   ```
   **Expected:** ClusterDeprovision completes within 30 minutes

7. **Action:** Verify Azure resource group cleaned up
   ```bash
   sleep 60

   az group show --name $RESOURCE_GROUP 2>&1 || echo "Resource group deleted"
   ```
   **Expected:** Resource group deleted or empty (depending on whether it was installer-created)

### Test Case HIVE-2302_003
**Name:** GCP Modern Cluster - metadata.json Secret and Generic Deprovision
**Description:** Validate GCP cluster provisioning and deprovision with metadata.json Secret
**Type:** E2E
**Priority:** High

#### Prerequisites
- Access to GCP credentials
- Hive operator deployed with HIVE-2302 changes
- GCP project with appropriate quotas

#### Test Steps
1. **Action:** Create GCP ClusterDeployment
   ```bash
   oc create namespace hive-test-gcp-modern

   oc create secret generic gcp-creds \
     --from-file=osServiceAccount.json=/path/to/gcp-service-account.json \
     -n hive-test-gcp-modern

   cat <<EOF | oc apply -f -
   apiVersion: hive.openshift.io/v1
   kind: ClusterDeployment
   metadata:
     name: gcp-modern-test
     namespace: hive-test-gcp-modern
   spec:
     baseDomain: example.com
     clusterName: gcp-modern-test
     platform:
       gcp:
         credentialsSecretRef:
           name: gcp-creds
         region: us-east1
     provisioning:
       installConfigSecretRef:
         name: install-config
       imageSetRef:
         name: openshift-v4.16.0
     pullSecretRef:
       name: pull-secret
   EOF
   ```
   **Expected:** ClusterDeployment created successfully

2. **Action:** Wait for installation and validate Secret creation
   ```bash
   oc wait --for=condition=Installed=true clusterdeployment/gcp-modern-test \
     -n hive-test-gcp-modern --timeout=60m

   SECRET_NAME=$(oc get clusterdeployment gcp-modern-test -n hive-test-gcp-modern \
     -o jsonpath='{.spec.clusterMetadata.metadataJSONSecretRef.name}')

   oc get secret $SECRET_NAME -n hive-test-gcp-modern
   ```
   **Expected:** Installation completes, metadata.json Secret created

3. **Action:** Verify metadata.json structure for GCP
   ```bash
   oc get secret $SECRET_NAME -n hive-test-gcp-modern \
     -o jsonpath='{.data.metadata\.json}' | base64 -d | jq . > /tmp/gcp-metadata.json

   INFRA_ID=$(cat /tmp/gcp-metadata.json | jq -r '.infraID')
   GCP_PROJECT=$(cat /tmp/gcp-metadata.json | jq -r '.gcp.projectID')

   echo "InfraID: $INFRA_ID"
   echo "GCP Project: $GCP_PROJECT"
   ```
   **Expected:** Valid GCP metadata with infraID and projectID

4. **Action:** Count GCP resources before deprovision
   ```bash
   gcloud compute instances list --project=$GCP_PROJECT \
     --filter="labels.kubernetes-io-cluster-${INFRA_ID}=owned" \
     --format=json > /tmp/gcp-instances-before.json

   INSTANCE_COUNT=$(cat /tmp/gcp-instances-before.json | jq '. | length')
   echo "GCP instances before deprovision: $INSTANCE_COUNT"
   ```
   **Expected:** Multiple GCP instances for cluster nodes

5. **Action:** Trigger deprovision and verify generic destroyer usage
   ```bash
   oc delete clusterdeployment gcp-modern-test -n hive-test-gcp-modern

   sleep 30

   DEPROVISION_POD=$(oc get pods -n hive-test-gcp-modern \
     -l hive.openshift.io/cluster-deprovision-name \
     -o jsonpath='{.items[0].metadata.name}')

   oc logs $DEPROVISION_POD -n hive-test-gcp-modern --tail=100 | \
     grep "metadata-json-secret-name"
   ```
   **Expected:** Deprovision logs show metadata.json Secret usage

6. **Action:** Wait for completion and verify cleanup
   ```bash
   DEPROVISION_NAME=$(oc get clusterdeprovision -n hive-test-gcp-modern \
     -o jsonpath='{.items[0].metadata.name}')

   oc wait --for=condition=Completed=true \
     clusterdeprovision/$DEPROVISION_NAME \
     -n hive-test-gcp-modern --timeout=30m

   sleep 60

   gcloud compute instances list --project=$GCP_PROJECT \
     --filter="labels.kubernetes-io-cluster-${INFRA_ID}=owned" \
     --format=json > /tmp/gcp-instances-after.json

   INSTANCE_COUNT_AFTER=$(cat /tmp/gcp-instances-after.json | jq '. | length')
   echo "GCP instances after deprovision: $INSTANCE_COUNT_AFTER"
   ```
   **Expected:** ClusterDeprovision completes successfully, 0 GCP instances remaining

### Test Case HIVE-2302_004
**Name:** Legacy Cluster Retrofit from ClusterProvision
**Description:** Verify automatic metadata.json Secret generation from existing ClusterProvision.Spec.MetadataJSON for legacy clusters
**Type:** E2E
**Priority:** High

#### Prerequisites
- Existing AWS cluster with ClusterProvision containing metadataJSON
- Hive operator upgraded to include HIVE-2302 changes
- ClusterDeployment does not have MetadataJSONSecretRef (legacy state)

#### Test Steps
1. **Action:** Verify legacy cluster state - has ClusterProvision but no Secret
   ```bash
   NAMESPACE="hive-legacy-provision-test"
   CLUSTER_NAME="legacy-provision-cluster"

   oc get clusterdeployment $CLUSTER_NAME -n $NAMESPACE \
     -o jsonpath='{.spec.clusterMetadata.metadataJSONSecretRef}'

   PROVISION_NAME=$(oc get clusterprovision -n $NAMESPACE \
     --sort-by=.metadata.creationTimestamp \
     -o jsonpath='{.items[-1].metadata.name}')

   oc get clusterprovision $PROVISION_NAME -n $NAMESPACE \
     -o jsonpath='{.spec.metadataJSON}' | jq . > /tmp/original-metadata.json

   cat /tmp/original-metadata.json
   ```
   **Expected:** ClusterDeployment has no MetadataJSONSecretRef, ClusterProvision.Spec.MetadataJSON contains valid JSON

2. **Action:** Trigger controller reconciliation to retrofit Secret
   ```bash
   oc annotate clusterdeployment $CLUSTER_NAME -n $NAMESPACE \
     hive.openshift.io/reconcile-trigger="$(date +%s)" --overwrite

   sleep 10

   SECRET_NAME=$(oc get clusterdeployment $CLUSTER_NAME -n $NAMESPACE \
     -o jsonpath='{.spec.clusterMetadata.metadataJSONSecretRef.name}')

   echo "Retrofitted Secret name: $SECRET_NAME"
   ```
   **Expected:** After reconciliation, ClusterDeployment.Spec.ClusterMetadata.MetadataJSONSecretRef populated

3. **Action:** Verify retrofitted Secret content matches ClusterProvision
   ```bash
   oc get secret $SECRET_NAME -n $NAMESPACE \
     -o jsonpath='{.data.metadata\.json}' | base64 -d > /tmp/retrofitted-metadata.json

   diff /tmp/original-metadata.json /tmp/retrofitted-metadata.json

   echo "Metadata comparison: $(diff /tmp/original-metadata.json /tmp/retrofitted-metadata.json && echo 'IDENTICAL' || echo 'DIFFERENT')"
   ```
   **Expected:** Retrofitted Secret content exactly matches ClusterProvision.Spec.MetadataJSON

4. **Action:** Verify bespoke fields remain populated (backward compatibility)
   ```bash
   oc get clusterdeployment $CLUSTER_NAME -n $NAMESPACE \
     -o jsonpath='{.spec.clusterMetadata.infraID}'

   oc get clusterdeployment $CLUSTER_NAME -n $NAMESPACE \
     -o jsonpath='{.spec.clusterMetadata.platform.aws}'
   ```
   **Expected:** Bespoke fields (infraID, platform-specific fields) still present

5. **Action:** Deprovision using generic destroyer
   ```bash
   oc delete clusterdeployment $CLUSTER_NAME -n $NAMESPACE

   sleep 30

   DEPROVISION_POD=$(oc get pods -n $NAMESPACE \
     -l hive.openshift.io/cluster-deprovision-name \
     -o jsonpath='{.items[0].metadata.name}')

   oc get pod $DEPROVISION_POD -n $NAMESPACE \
     -o jsonpath='{.spec.containers[0].command[*]}' | \
     grep "deprovision --metadata-json-secret-name"
   ```
   **Expected:** Deprovision job uses generic destroyer with retrofitted Secret

6. **Action:** Monitor deprovision completion and resource cleanup
   ```bash
   DEPROVISION_NAME=$(oc get clusterdeprovision -n $NAMESPACE \
     -o jsonpath='{.items[0].metadata.name}')

   oc wait --for=condition=Completed=true \
     clusterdeprovision/$DEPROVISION_NAME \
     -n $NAMESPACE --timeout=30m

   INFRA_ID=$(cat /tmp/retrofitted-metadata.json | jq -r '.infraID')

   aws resourcegroupstaggingapi get-resources \
     --tag-filters "Key=kubernetes.io/cluster/${INFRA_ID},Values=owned" \
     --region us-east-1 \
     --query 'ResourceTagMappingList[*].ResourceARN' | jq '. | length'
   ```
   **Expected:** Deprovision completes successfully, 0 AWS resources remaining

### Test Case HIVE-2302_005
**Name:** Legacy Cluster Retrofit from Bespoke Fields
**Description:** Validate metadata.json Secret generation from bespoke ClusterMetadata fields when ClusterProvision unavailable
**Type:** E2E
**Priority:** High

#### Prerequisites
- Ancient cluster with only bespoke ClusterMetadata fields
- No ClusterProvision or ClusterProvision.Spec.MetadataJSON is empty
- Hive operator with HIVE-2302 changes

#### Test Steps
1. **Action:** Verify legacy cluster state - no ClusterProvision, only bespoke fields
   ```bash
   NAMESPACE="hive-legacy-bespoke-test"
   CLUSTER_NAME="legacy-bespoke-cluster"

   oc get clusterprovision -n $NAMESPACE 2>&1 || echo "No ClusterProvision found"

   INFRA_ID=$(oc get clusterdeployment $CLUSTER_NAME -n $NAMESPACE \
     -o jsonpath='{.spec.clusterMetadata.infraID}')

   echo "InfraID from bespoke field: $INFRA_ID"

   oc get clusterdeployment $CLUSTER_NAME -n $NAMESPACE \
     -o jsonpath='{.spec.clusterMetadata}' | jq .
   ```
   **Expected:** No ClusterProvision exists, bespoke fields (infraID, platform) populated

2. **Action:** Trigger controller to retrofit Secret from bespoke fields
   ```bash
   oc annotate clusterdeployment $CLUSTER_NAME -n $NAMESPACE \
     hive.openshift.io/reconcile-trigger="$(date +%s)" --overwrite

   sleep 10

   SECRET_NAME=$(oc get clusterdeployment $CLUSTER_NAME -n $NAMESPACE \
     -o jsonpath='{.spec.clusterMetadata.metadataJSONSecretRef.name}')

   echo "Generated Secret name: $SECRET_NAME"
   ```
   **Expected:** Secret reference populated after reconciliation

3. **Action:** Verify generated metadata.json structure
   ```bash
   oc get secret $SECRET_NAME -n $NAMESPACE \
     -o jsonpath='{.data.metadata\.json}' | base64 -d | jq . > /tmp/generated-metadata.json

   cat /tmp/generated-metadata.json

   GENERATED_INFRA_ID=$(cat /tmp/generated-metadata.json | jq -r '.infraID')
   echo "Generated infraID: $GENERATED_INFRA_ID"

   [[ "$GENERATED_INFRA_ID" == "$INFRA_ID" ]] && echo "InfraID match: SUCCESS" || echo "InfraID match: FAILED"
   ```
   **Expected:** Generated metadata.json contains infraID matching bespoke field, platform-specific data included

4. **Action:** Verify metadata.json contains platform-specific fields from bespoke data
   ```bash
   PLATFORM=$(cat /tmp/generated-metadata.json | jq -r 'keys[]' | grep -E 'aws|azure|gcp')
   echo "Platform in metadata: $PLATFORM"

   cat /tmp/generated-metadata.json | jq ".$PLATFORM"
   ```
   **Expected:** Platform section present with data reconstructed from bespoke fields

5. **Action:** Test deprovision with retrofitted metadata
   ```bash
   oc delete clusterdeployment $CLUSTER_NAME -n $NAMESPACE

   DEPROVISION_POD=$(oc get pods -n $NAMESPACE \
     -l hive.openshift.io/cluster-deprovision-name \
     -o jsonpath='{.items[0].metadata.name}')

   oc logs $DEPROVISION_POD -n $NAMESPACE -f --tail=100 | head -50
   ```
   **Expected:** Deprovision pod starts and uses generated metadata.json

6. **Action:** Verify successful deprovision completion
   ```bash
   DEPROVISION_NAME=$(oc get clusterdeprovision -n $NAMESPACE \
     -o jsonpath='{.items[0].metadata.name}')

   oc wait --for=condition=Completed=true \
     clusterdeprovision/$DEPROVISION_NAME \
     -n $NAMESPACE --timeout=30m
   ```
   **Expected:** Deprovision completes successfully using retrofitted metadata

### Test Case HIVE-2302_006
**Name:** Legacy Deprovision Escape Hatch with Annotation
**Description:** Verify annotation enables legacy platform-specific destroyer instead of generic destroyer
**Type:** E2E
**Priority:** Medium

#### Prerequisites
- AWS cluster with metadata.json Secret (modern cluster)
- Hive operator with HIVE-2302 changes

#### Test Steps
1. **Action:** Create AWS cluster and verify metadata.json Secret exists
   ```bash
   NAMESPACE="hive-test-escape-hatch"
   CLUSTER_NAME="escape-hatch-cluster"

   # Cluster should be provisioned normally with metadata.json Secret
   SECRET_NAME=$(oc get clusterdeployment $CLUSTER_NAME -n $NAMESPACE \
     -o jsonpath='{.spec.clusterMetadata.metadataJSONSecretRef.name}')

   echo "Metadata Secret: $SECRET_NAME"
   oc get secret $SECRET_NAME -n $NAMESPACE
   ```
   **Expected:** Modern cluster has metadata.json Secret

2. **Action:** Add legacy-deprovision annotation to ClusterDeployment
   ```bash
   oc annotate clusterdeployment $CLUSTER_NAME -n $NAMESPACE \
     hive.openshift.io/legacy-deprovision="true"

   oc get clusterdeployment $CLUSTER_NAME -n $NAMESPACE \
     -o jsonpath='{.metadata.annotations.hive\.openshift\.io/legacy-deprovision}'
   ```
   **Expected:** Annotation added successfully with value "true"

3. **Action:** Delete ClusterDeployment and check deprovision job command
   ```bash
   oc delete clusterdeployment $CLUSTER_NAME -n $NAMESPACE

   sleep 30

   DEPROVISION_POD=$(oc get pods -n $NAMESPACE \
     -l hive.openshift.io/cluster-deprovision-name \
     -o jsonpath='{.items[0].metadata.name}')

   oc get pod $DEPROVISION_POD -n $NAMESPACE \
     -o jsonpath='{.spec.containers[0].command[*]}' > /tmp/deprovision-command.txt

   cat /tmp/deprovision-command.txt
   ```
   **Expected:** Command uses legacy AWS deprovision (e.g., "hiveutil aws deprovision" NOT "hiveutil deprovision --metadata-json-secret-name")

4. **Action:** Verify volume mounts do NOT include metadata-json Secret
   ```bash
   oc get pod $DEPROVISION_POD -n $NAMESPACE \
     -o jsonpath='{.spec.volumes[*].name}' | grep "metadata-json" || echo "No metadata-json volume (expected for legacy mode)"
   ```
   **Expected:** No metadata-json volume mounted (legacy mode doesn't use it)

5. **Action:** Check deprovision logs for legacy destroyer usage
   ```bash
   oc logs $DEPROVISION_POD -n $NAMESPACE --tail=100 | \
     grep -E "aws deprovision|ClusterUninstaller"
   ```
   **Expected:** Logs show legacy AWS-specific destroyer initialization

6. **Action:** Monitor completion and verify successful cleanup
   ```bash
   DEPROVISION_NAME=$(oc get clusterdeprovision -n $NAMESPACE \
     -o jsonpath='{.items[0].metadata.name}')

   oc wait --for=condition=Completed=true \
     clusterdeprovision/$DEPROVISION_NAME \
     -n $NAMESPACE --timeout=30m

   INFRA_ID=$(oc get clusterdeprovision $DEPROVISION_NAME -n $NAMESPACE \
     -o jsonpath='{.spec.infraID}')

   aws resourcegroupstaggingapi get-resources \
     --tag-filters "Key=kubernetes.io/cluster/${INFRA_ID},Values=owned" \
     --region us-east-1 \
     --query 'ResourceTagMappingList[*].ResourceARN' | jq '. | length'
   ```
   **Expected:** Deprovision completes successfully using legacy path, 0 resources remaining
