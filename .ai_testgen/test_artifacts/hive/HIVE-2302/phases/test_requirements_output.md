# Test Requirements Output

## Component Name
hive

## Card Summary
Pass installer's metadata.json directly to destroyer

## Test Requirements

### Root Cause
Previously, Hive required plumbing individual metadata fields from the installer's metadata.json through a complex sequence (ClusterProvision → ClusterDeprovision → uninstall pod → destroyer) whenever the installer added new fields. This created maintenance burden and required code changes for every new metadata field.

### Technical Changes
The solution passes the entire metadata.json as an opaque blob through a new Secret:
1. **New Secret Creation**: Creates `{clustername}-metadata-json` Secret containing the metadata.json from the installer
2. **New API Field**: Adds `ClusterDeployment.Spec.ClusterMetadata.MetadataJSONSecretRef` to reference the Secret
3. **Generic Destroyer**: New `hiveutil deprovision` command that unmarshals metadata.json directly into the destroyer
4. **Legacy Retrofit**: For clusters created before this change, retrofits the Secret from existing bespoke fields or ClusterProvision
5. **Escape Hatch**: Annotation `hive.openshift.io/legacy-deprovision: "true"` to use platform-specific legacy destroyer
6. **Deprecated Fields**: Platform-specific metadata fields (AWS HostedZoneRole, Azure ResourceGroupName, GCP NetworkProjectID) marked as deprecated

### Test Objectives
1. Verify metadata.json Secret creation during cluster provisioning
2. Validate cluster deprovisioning using new generic destroyer with metadata.json
3. Confirm legacy cluster retrofit mechanism works correctly
4. Test legacy deprovision escape hatch annotation
5. Ensure all platforms (AWS, Azure, GCP, vSphere, IBMCloud, Nutanix, OpenStack) deprovision cleanly
6. Verify special platform-specific fields (HostedZoneRole, ResourceGroupName, NetworkProjectID) work through both modern and retrofit scenarios
7. Test vSphere pre-zonal and post-zonal metadata.json formats

## Affected Platforms
- AWS (with HostedZoneRole support)
- Azure (with ResourceGroupName support)
- GCP (with NetworkProjectID support for shared VPC)
- vSphere (pre-zonal and post-zonal formats)
- IBMCloud
- Nutanix
- OpenStack

## Test Scenarios

### Scenario 1: Modern Cluster Provisioning and Deprovisioning
- Provision new cluster (creates metadata.json Secret automatically)
- Verify metadata.json Secret creation with correct content
- Verify ClusterDeployment.Spec.ClusterMetadata.MetadataJSONSecretRef populated
- Deprovision cluster using generic destroyer
- Confirm complete resource cleanup

### Scenario 2: Legacy Cluster Retrofit from ClusterProvision
- Use existing cluster with ClusterProvision containing metadataJSON
- Upgrade Hive to version with this feature
- Verify automatic Secret creation from ClusterProvision.Spec.MetadataJSON
- Deprovision using new generic destroyer
- Validate complete cleanup

### Scenario 3: Legacy Cluster Retrofit from Bespoke Fields
- Use cluster created before ClusterProvision.Spec.MetadataJSON existed
- Only bespoke ClusterDeployment.Spec.ClusterMetadata fields available
- Verify Secret retrofitted from bespoke fields
- Deprovision successfully
- Confirm no resource leaks

### Scenario 4: Legacy Deprovision Escape Hatch
- Provision cluster with metadata.json Secret
- Add annotation `hive.openshift.io/legacy-deprovision: "true"`
- Verify legacy platform-specific destroyer is used
- Confirm successful deprovision with complete cleanup

### Scenario 5: Platform-Specific Special Fields (AWS HostedZoneRole)
- Provision AWS cluster with HostedZoneRole for shared VPC
- Verify metadata.json contains HostedZoneRole
- Test both modern and retrofit scenarios
- Deprovision successfully with HostedZoneRole utilized

### Scenario 6: Platform-Specific Special Fields (Azure ResourceGroupName)
- Provision Azure cluster with custom ResourceGroupName
- Verify metadata.json includes ResourceGroupName
- Test modern and retrofit paths
- Deprovision successfully

### Scenario 7: Platform-Specific Special Fields (GCP NetworkProjectID)
- Provision GCP cluster with shared VPC (NetworkProjectID)
- Verify metadata.json contains NetworkProjectID
- Test modern and retrofit scenarios
- Deprovision cleanly

### Scenario 8: vSphere Pre-Zonal Format
- Use vSphere cluster from release before zonal support
- Verify pre-zonal metadata.json format handling
- Retrofit Secret correctly
- Deprovision successfully

### Scenario 9: vSphere Post-Zonal Format
- Provision vSphere cluster with zonal support
- Verify post-zonal metadata.json format
- Deprovision using generic destroyer
- Confirm cleanup

## Edge Cases

### EC1: Missing ClusterProvision and Bespoke Fields
- Cluster with no ClusterProvision and incomplete bespoke fields
- Verify retrofit behavior (best effort)
- Handle gracefully with appropriate error messages

### EC2: Credential Scrubbing in metadata.json
- Verify sensitive credentials removed from metadata.json Secret (HIVE-2804)
- Ensure credentials re-injected during deprovision for platforms requiring them (vSphere, Nutanix)

### EC3: Pre-Existing Secret
- metadata.json Secret already exists (user-modified)
- Verify retrofit logic skips regeneration
- User modifications preserved

### EC4: Secret Deletion and Regeneration
- User deletes metadata.json Secret
- Trigger retrofit logic to regenerate from available sources
- Verify regenerated Secret validity

### EC5: Platform Migration or Inconsistency
- metadata.json platform differs from ClusterDeployment platform specification
- Handle error case appropriately

### EC6: Incomplete metadata.json
- metadata.json missing required fields for platform
- Deprovision fails with clear error message
- No partial cleanup leaving orphaned resources

### EC7: Backward Compatibility
- Clusters transitioning through multiple Hive versions
- Verify bespoke fields still populated
- Ensure both old and new code paths functional during upgrade window
