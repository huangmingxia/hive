# Test Strategy

## Test Coverage Matrix

| Scenario | Platform | Cluster Type | Test Type | Reasoning | Priority |
|----------|----------|--------------|-----------|-----------|----------|
| Modern cluster provision/deprovision | AWS, Azure, GCP, vSphere | Standard | E2E | Core functionality validation - new Secret creation and generic destroyer | High |
| Legacy retrofit from ClusterProvision | AWS, Azure, GCP | Standard | E2E | Validate automatic Secret generation from existing metadataJSON | High |
| Legacy retrofit from bespoke fields | AWS, Azure, GCP | Standard | E2E | Validate fallback Secret generation from legacy fields | High |
| Legacy deprovision escape hatch | AWS | Standard | E2E | Verify annotation-based legacy destroyer still works | Medium |
| AWS with HostedZoneRole | AWS | Shared VPC | Manual | Requires pre-existing hosted zone in another account | High |
| Azure with ResourceGroupName | Azure | Custom RG | Manual | Requires pre-created resource group | High |
| GCP with NetworkProjectID | GCP | Shared VPC | Manual | Requires shared VPC setup with separate network project | High |
| vSphere pre-zonal format | vSphere | Standard | Manual | Requires vSphere infrastructure and pre-zonal release version | Medium |
| vSphere post-zonal format | vSphere | Standard | Manual | Requires vSphere infrastructure with zonal support | Medium |
| IBMCloud deprovision | IBMCloud | Standard | Manual | Requires IBMCloud infrastructure | Medium |
| Nutanix deprovision | Nutanix | Standard | Manual | Requires Nutanix infrastructure | Medium |
| OpenStack deprovision | OpenStack | Standard | Manual | Requires OpenStack infrastructure | Medium |

## Test Scenarios

### Scenario 1: Modern Cluster Provisioning with metadata.json Secret
**Objective:** Validate new clusters automatically create metadata.json Secret and use generic destroyer

**User Workflow:**
1. Operator provisions new Hive-managed OpenShift cluster
2. Installer generates metadata.json during installation
3. Hive creates Secret containing metadata.json
4. ClusterDeployment references Secret via MetadataJSONSecretRef
5. Operator deletes ClusterDeployment to trigger deprovision
6. Generic destroyer reads metadata.json from Secret
7. Cluster resources cleaned up completely

**Test Approach:**
- Create ClusterDeployment on target platform
- Wait for installation to complete
- Verify Secret creation and population
- Inspect Secret content matches installer output
- Delete ClusterDeployment
- Monitor deprovision job using generic destroyer
- Verify complete resource cleanup

**Quantitative Validation:**
- Secret `{clustername}-metadata-json` exists in CD namespace
- Secret contains key `metadata.json` with valid JSON content
- ClusterDeployment.Spec.ClusterMetadata.MetadataJSONSecretRef.Name equals `{clustername}-metadata-json`
- Deprovision job mounts Secret as volume
- Deprovision logs show "hiveutil deprovision --metadata-json-secret-name"
- All cloud resources deleted (0 resources tagged with infraID remain)
- ClusterDeprovision finalizes with Completed=True within expected timeframe

### Scenario 2: Legacy Cluster Retrofit from ClusterProvision
**Objective:** Verify automatic Secret generation for clusters with existing ClusterProvision.Spec.MetadataJSON

**User Workflow:**
1. Cluster exists from before this feature (has ClusterProvision with metadataJSON)
2. Hive operator upgraded to include this feature
3. ClusterDeployment controller detects missing Secret
4. Controller extracts metadataJSON from ClusterProvision
5. Controller creates metadata.json Secret
6. Deprovision uses generic destroyer successfully

**Test Approach:**
- Simulate legacy cluster by creating ClusterDeployment without Secret but with ClusterProvision
- Manually populate ClusterProvision.Spec.MetadataJSON
- Trigger controller reconciliation
- Verify Secret created from ClusterProvision data
- Deprovision cluster
- Validate cleanup

**Quantitative Validation:**
- ClusterProvision.Spec.MetadataJSON field populated with valid JSON
- After reconciliation: Secret created within 60 seconds
- Secret.Data["metadata.json"] exactly matches ClusterProvision.Spec.MetadataJSON
- Bespoke fields (InfraID, platform-specific) remain populated (backward compatibility)
- Deprovision completes successfully using generic destroyer
- No resource leaks (0 cloud resources remain)

### Scenario 3: Legacy Cluster Retrofit from Bespoke Fields
**Objective:** Validate Secret reconstruction from ClusterMetadata bespoke fields when ClusterProvision unavailable

**User Workflow:**
1. Ancient cluster from before ClusterProvision.Spec.MetadataJSON existed
2. Only bespoke ClusterDeployment.Spec.ClusterMetadata fields available
3. Hive controller reconstructs metadata.json from bespoke fields
4. Creates Secret with reconstructed content
5. Deprovision succeeds using generic destroyer

**Test Approach:**
- Create ClusterDeployment with only bespoke fields populated (no ClusterProvision)
- Delete or omit ClusterProvision
- Trigger reconciliation
- Verify Secret generation from bespoke field logic
- Compare generated metadata.json structure
- Deprovision successfully

**Quantitative Validation:**
- ClusterProvision does not exist or has empty metadataJSON
- Bespoke fields populated: InfraID, platform-specific fields
- Secret created within 60 seconds of reconciliation
- Secret.Data["metadata.json"] contains valid structure matching platform
- Generated metadata includes: ClusterName, InfraID, ClusterID, Platform-specific data
- Deprovision job succeeds
- All platform resources cleaned up (0 tagged resources remaining)

### Scenario 4: Legacy Deprovision Escape Hatch
**Objective:** Ensure annotation allows reverting to platform-specific legacy destroyer

**User Workflow:**
1. Operator provisions cluster normally (has metadata.json Secret)
2. Issue discovered with generic destroyer for specific platform
3. Operator adds annotation `hive.openshift.io/legacy-deprovision: "true"`
4. Delete ClusterDeployment
5. Platform-specific legacy destroyer used instead of generic
6. Cluster deprovisioned successfully

**Test Approach:**
- Provision cluster (creates metadata.json Secret)
- Add legacy-deprovision annotation to ClusterDeployment
- Delete ClusterDeployment
- Verify deprovision job uses legacy destroyer path
- Check logs for platform-specific deprovision commands (not generic)
- Confirm cleanup

**Quantitative Validation:**
- Annotation `hive.openshift.io/legacy-deprovision: "true"` present on CD
- Deprovision job spec uses legacy command (e.g., "hiveutil aws deprovision" not "hiveutil deprovision")
- Logs show platform-specific destroyer initialization
- All resources cleaned (0 tagged resources)
- No errors related to metadata.json Secret mounting

### Scenario 5: AWS Cluster with HostedZoneRole (Shared VPC)
**Objective:** Validate HostedZoneRole field flows through metadata.json in both modern and retrofit scenarios

**User Workflow:**
1. Operator provisions AWS cluster with hosted zone in different account
2. Install-config specifies HostedZoneRole ARN
3. Installer writes HostedZoneRole to metadata.json
4. Hive captures full metadata including HostedZoneRole
5. Deprovision uses HostedZoneRole to clean DNS records in external account

**Test Approach:**
- **Modern**: Provision cluster with HostedZoneRole in install-config
  - Verify metadata.json Secret contains HostedZoneRole
  - Deprovision and check DNS cleanup in external account
- **Retrofit**: Simulate legacy cluster with bespoke AWS.HostedZoneRole field
  - Verify retrofitted Secret includes HostedZoneRole
  - Deprovision successfully

**Quantitative Validation:**
- Modern path: Secret.Data["metadata.json"] contains `aws.hostedZoneRole` field
- Retrofit path: Secret generated with HostedZoneRole from CD.Spec.ClusterMetadata.Platform.AWS.HostedZoneRole
- Deprovision logs show AssumeRole operation for HostedZoneRole
- DNS records in external account deleted (0 records for cluster domain)
- No permissions errors during deprovision

### Scenario 6: Azure Cluster with ResourceGroupName
**Objective:** Verify custom ResourceGroupName handled correctly through metadata.json

**User Workflow:**
1. Operator provisions Azure cluster with pre-created resource group
2. Install-config specifies resourceGroupName
3. Installer records ResourceGroupName in metadata.json
4. Hive preserves ResourceGroupName in Secret
5. Deprovision targets correct resource group

**Test Approach:**
- **Modern**: Provision with custom ResourceGroupName
  - Verify metadata.json contains correct RG name
  - Deprovision and verify RG handling
- **Retrofit**: Legacy cluster with bespoke Azure.ResourceGroupName
  - Generate Secret from bespoke field
  - Deprovision successfully

**Quantitative Validation:**
- Modern: Secret contains `azure.resourceGroupName` matching install-config
- Retrofit: Secret populated with value from CD.Spec.ClusterMetadata.Platform.Azure.ResourceGroupName
- Deprovision targets specified resource group (not default infraID-rg)
- All resources in resource group deleted
- Resource group itself deleted if installer-created, preserved if pre-existing

### Scenario 7: GCP Cluster with NetworkProjectID (Shared VPC)
**Objective:** Validate NetworkProjectID for shared VPC deployments flows through correctly

**User Workflow:**
1. Operator provisions GCP cluster using shared VPC from another project
2. Install-config specifies networkProjectID
3. Installer includes NetworkProjectID in metadata.json
4. Deprovision accesses network resources in separate project

**Test Approach:**
- **Modern**: Provision with NetworkProjectID in install-config
  - Verify metadata.json includes NetworkProjectID
  - Deprovision and check cross-project network cleanup
- **Retrofit**: Legacy cluster with bespoke GCP.NetworkProjectID
  - Generate Secret from bespoke field
  - Deprovision successfully

**Quantitative Validation:**
- Modern: Secret.Data["metadata.json"] contains `gcp.networkProjectID`
- Retrofit: Secret includes value from CD.Spec.ClusterMetadata.Platform.GCP.NetworkProjectID
- Deprovision logs show operations in network project
- Firewall rules in network project deleted
- Compute resources in cluster project deleted
- No orphaned network resources

### Scenario 8: vSphere Pre-Zonal Metadata Format
**Objective:** Ensure pre-zonal vSphere metadata.json format handled correctly during retrofit

**User Workflow:**
1. vSphere cluster from release before zonal support
2. metadata.json has pre-zonal structure with single username/password at top level
3. Retrofit creates Secret from legacy format
4. Generic destroyer handles pre-zonal structure
5. Deprovision succeeds with credential injection

**Test Approach:**
- Use vSphere cluster from pre-zonal release
- Verify metadata.json structure (flat credentials)
- Trigger retrofit if needed
- Deprovision using generic destroyer
- Validate credential injection and cleanup

**Quantitative Validation:**
- metadata.json contains `vsphere.username` and `vsphere.password` fields (scrubbed)
- Deprovision job injects credentials from environment variables
- Deprovision logs show credential injection for pre-zonal format
- All VMs, folders, resource pools deleted
- vCenter inventory cleaned completely

### Scenario 9: vSphere Post-Zonal Metadata Format
**Objective:** Validate post-zonal vSphere metadata.json with vCenters array structure

**User Workflow:**
1. vSphere cluster with zonal support enabled
2. metadata.json has vCenters array with per-vCenter credentials
3. Secret created with zonal structure
4. Deprovision handles vCenters array correctly
5. Credentials injected into array elements

**Test Approach:**
- Provision vSphere cluster with zonal configuration
- Verify metadata.json has vCenters array structure
- Deprovision cluster
- Validate credential injection for each vCenter
- Confirm cleanup across zones

**Quantitative Validation:**
- metadata.json contains `vsphere.vcenters[]` array
- Each vCenter has username/password fields (scrubbed)
- Deprovision injects credentials into each array element
- Resources deleted from all zones/vCenters
- Complete inventory cleanup (0 resources tagged with infraID)

## Validation Methods

### Secret Inspection
**Method:** Examine metadata.json Secret content and references
- Use `oc get secret {clustername}-metadata-json -o jsonpath='{.data.metadata\.json}' | base64 -d | jq` to view content
- Verify JSON structure matches platform expectations
- Check ClusterDeployment.Spec.ClusterMetadata.MetadataJSONSecretRef points to correct Secret
- Confirm Secret has ownerReference to ClusterDeployment for garbage collection

### Deprovision Job Analysis
**Method:** Inspect deprovision job configuration and logs
- Check job spec for volume mounts of metadata.json Secret
- Verify command: `hiveutil deprovision --metadata-json-secret-name` vs legacy platform-specific commands
- Analyze logs for Secret unmarshaling success
- Track destroyer initialization with correct metadata

### Cloud Resource Verification
**Method:** Query cloud provider APIs for resource cleanup
- AWS: `aws resourcegroupstaggingapi get-resources --tag-filters Key=kubernetes.io/cluster/{infraID},Values=owned` should return 0 resources
- Azure: `az resource list --tag kubernetes.io-cluster-{infraID}=owned` should be empty
- GCP: `gcloud compute instances list --filter="labels.kubernetes-io-cluster-{infraID}=owned"` should return nothing
- Count resources before and after deprovision (delta should equal all created resources)

### Retrofit Logic Validation
**Method:** Test Secret generation from various source states
- Scenario: ClusterProvision exists with metadataJSON → verify byte-for-byte copy
- Scenario: No ClusterProvision → verify generation from bespoke fields
- Scenario: Secret already exists → verify no regeneration (idempotent)
- Track controller logs for retrofit decision path

### Backward Compatibility Check
**Method:** Verify bespoke fields remain populated
- After Secret creation, verify CD.Spec.ClusterMetadata.InfraID still present
- Check platform-specific fields (AWS.HostedZoneRole, Azure.ResourceGroupName, GCP.NetworkProjectID) remain
- Ensure deprecated fields continue working during transition period

### Credential Scrubbing Validation
**Method:** Confirm sensitive data removed from Secret
- vSphere: Check metadata.json has empty username/password fields
- Nutanix: Verify credentials scrubbed
- Compare Secret content with original installer metadata.json (if available)
- Ensure no plaintext passwords in Secret

### Legacy Destroyer Activation
**Method:** Verify annotation triggers platform-specific code path
- Add annotation `hive.openshift.io/legacy-deprovision: "true"`
- Check deprovision job uses legacy command syntax
- Verify logs show platform-specific destroyer (not generic)
- Confirm no attempts to mount metadata.json Secret in legacy mode
