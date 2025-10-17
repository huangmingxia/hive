# Test Requirements Output

## Component Name
hive

## Card Summary
[Hive/Azure/Azure Gov] Machinepool: Unable to generate machinesets in regions without availability zone support.

## Test Requirements

### Root Cause Analysis
When creating a MachinePool in Azure regions that do not support availability zones (e.g., usgovtexas), the MachinePool controller returned an error "zero zones returned for region" and failed to generate MachineSet resources. The controller previously required zones to be present and did not handle non-zoned regions.

### Technical Changes
The fix modifies the Azure actuator logic in `pkg/controller/machinepool/azureactuator.go`:
- When no zones are specified by the user and the region returns zero zones, the system now passes through "no zones" to the MachineSet generator
- A log message is added for visibility when deploying to non-zoned regions
- The MachineSet generator handles non-zoned deployments by creating a single MachineSet without zone specification
- Unit tests updated to verify non-zoned deployment behavior

### Test Objectives
1. Verify MachinePool creation succeeds in Azure regions without availability zone support
2. Validate MachineSet generation for non-zoned regions
3. Confirm proper logging when deploying to regions without zones
4. Ensure backward compatibility with zoned regions

## Affected Platforms
- Azure (Commercial Cloud)
- Azure Government Cloud

**Regions Without AZ Support:**
- Azure Gov: usgovtexas, usgovarizona, usdodeast, usdodcentral
- Azure Commercial: Various regional locations without zone support

## Test Scenarios

### Scenario 1: MachinePool Creation in Non-Zoned Azure Region
- Create a ClusterDeployment in an Azure region without availability zones
- Create a MachinePool without specifying zones
- Verify MachineSet is created successfully as non-zoned deployment
- Validate log message indicates non-zoned deployment

### Scenario 2: MachinePool Creation in Non-Zoned Azure Gov Region
- Create a ClusterDeployment in an Azure Government region without availability zones (e.g., usgovtexas)
- Create a MachinePool without specifying zones
- Verify MachineSet generation succeeds
- Confirm deployment completes successfully

### Scenario 3: MachinePool with Explicit Zones in Non-Zoned Region
- Create a ClusterDeployment in a non-zoned Azure region
- Create a MachinePool with explicit zone specification
- Verify system logs warning and deploys as non-zoned (graceful degradation)
- Validate MachineSet creation succeeds without zones

### Scenario 4: Backward Compatibility - Zoned Region Behavior
- Create a ClusterDeployment in an Azure region with availability zone support
- Create a MachinePool without specifying zones
- Verify zones are automatically fetched and applied
- Confirm multiple MachineSet resources created (one per zone)

## Edge Cases

### EC1: Region Zone Detection
- Validate correct zone detection for regions with and without AZ support
- Verify zone list fetching from Azure API
- Ensure proper handling when Azure API returns empty zone list

### EC2: User Zone Override Handling
- When user specifies zones but region doesn't support them
- Verify graceful degradation with warning message
- Ensure deployment proceeds as non-zoned

### EC3: MachineSet Replica Distribution
- Non-zoned region: Single MachineSet with full replica count
- Zoned region: Multiple MachineSets with replicas distributed across zones
- Validate replica count accuracy in both scenarios

### EC4: Cross-Platform Consistency
- Compare behavior between Azure Commercial and Azure Government clouds
- Verify consistent handling of non-zoned regions across both platforms
