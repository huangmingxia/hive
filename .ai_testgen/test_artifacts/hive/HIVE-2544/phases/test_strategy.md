# Test Strategy

## Test Coverage Matrix

| Scenario | Platform | Cluster Type | Test Type | Reasoning | Priority |
|----------|----------|--------------|-----------|-----------|----------|
| MachinePool in non-zoned Azure region | Azure | Standard | E2E | Core fix validation - verify MachineSet generation succeeds in regions without AZ support | High |
| MachinePool in non-zoned Azure Gov region | Azure Government | Standard | Manual | Azure Gov requires special infrastructure access | High |
| User-specified zones in non-zoned region | Azure | Standard | E2E | Edge case validation - graceful degradation when zones specified but unavailable | Medium |
| Zoned region backward compatibility | Azure | Standard | E2E | Regression prevention - ensure zoned regions still work correctly | High |

## Test Scenarios

### Scenario 1: MachinePool Creation in Non-Zoned Azure Region
**Objective:** Validate MachinePool and MachineSet creation in Azure regions without availability zone support

**User Workflow:**
1. Operator provisions Hive-managed cluster in non-zoned Azure region
2. Operator creates MachinePool to add compute capacity
3. System generates MachineSet for worker node provisioning
4. Machines are created and join the cluster

**Test Approach:**
- Deploy ClusterDeployment to Azure region without zone support
- Create MachinePool resource without zone specification
- Monitor MachinePool reconciliation process
- Validate MachineSet creation with non-zoned configuration
- Verify Machine resources are created successfully

**Quantitative Validation:**
- MachinePool status condition: `Ready=True` within 5 minutes
- MachineSet count: exactly 1 (non-zoned deployment)
- MachineSet replicas: equals MachinePool replica count
- Controller logs: contains "No availability zones detected for region. Using non-zoned deployment."
- Machine count: equals requested replica count within 10 minutes

### Scenario 2: MachinePool in Non-Zoned Azure Government Region
**Objective:** Verify functionality in Azure Government cloud regions without zone support

**User Workflow:**
1. Operator manages OpenShift clusters in Azure Government cloud (regulated environments)
2. Operator deploys to regions like usgovtexas (no AZ support)
3. Operator creates MachinePool for workload scaling
4. System provisions compute resources in non-zoned configuration

**Test Approach:**
- Provision ClusterDeployment in Azure Gov region (e.g., usgovtexas)
- Create MachinePool with standard configuration
- Validate MachineSet generation
- Verify Machine provisioning completes

**Quantitative Validation:**
- ClusterDeployment installation: `Installed=True`
- MachinePool status: `Ready=True` within 5 minutes
- MachineSet count: 1 (non-zoned)
- Log pattern match: "Using non-zoned deployment" for usgovtexas region
- No error messages containing "zero zones returned"

### Scenario 3: User-Specified Zones in Non-Zoned Region
**Objective:** Validate graceful degradation when user configuration conflicts with region capabilities

**User Workflow:**
1. Operator creates MachinePool with zones specified in configuration
2. Target region does not support availability zones
3. System detects mismatch and proceeds with safe fallback
4. Deployment completes with warning logged

**Test Approach:**
- Deploy ClusterDeployment to non-zoned Azure region
- Create MachinePool with explicit zones field populated
- Monitor controller behavior and logging
- Verify deployment proceeds as non-zoned despite user input
- Validate warning messages in logs

**Quantitative Validation:**
- MachinePool reconciliation: succeeds (no error state)
- MachineSet zones field: empty or absent (user input overridden)
- Warning log count: at least 1 entry containing zone override message
- MachineSet count: 1 (not split by zones)
- MachinePool status: `Ready=True` (graceful degradation succeeded)

### Scenario 4: Backward Compatibility - Zoned Region Behavior
**Objective:** Ensure fix does not break existing functionality for regions with zone support

**User Workflow:**
1. Operator deploys cluster to Azure region with availability zones (e.g., eastus)
2. Operator creates MachinePool without specifying zones
3. System fetches available zones from Azure API
4. MachineSets created with zone distribution
5. Machines provisioned across multiple zones

**Test Approach:**
- Deploy ClusterDeployment to Azure region with zone support (e.g., eastus, westus2)
- Create MachinePool without zones specification
- Verify automatic zone detection
- Validate multiple MachineSet creation (one per zone)
- Confirm replica distribution across zones

**Quantitative Validation:**
- Zone detection: API returns >= 1 zone
- MachineSet count: equals number of available zones (typically 3)
- Each MachineSet zones field: contains exactly 1 zone
- Total replicas across MachineSets: equals MachinePool replica count
- Replica distribution: balanced (difference <= 1 between MachineSets)
- MachinePool status: `Ready=True` with proper zone information

## Validation Methods

### Status Validation
**Method:** Monitor Kubernetes resource status conditions
- Use `oc get machinepool -o jsonpath='{.status.conditions}'` to check Ready condition
- Verify `Ready=True` status with reason indicating successful reconciliation
- Check MachineSet owner references point to MachinePool
- Validate Machine count matches expected replica count

### Log Analysis
**Method:** Query MachinePool controller logs for specific patterns
- Search for "No availability zones detected for region. Using non-zoned deployment." in non-zoned scenarios
- Verify absence of "zero zones returned" error messages after fix
- Count log entries to ensure warnings appear when expected
- Confirm log messages reference correct region names

### Resource Count Verification
**Method:** Count generated Kubernetes resources
- Query MachineSets: `oc get machinesets -n openshift-machine-api --selector hive.openshift.io/machine-pool={pool-name}`
- Expected count for non-zoned: exactly 1
- Expected count for zoned: number of zones (typically 3)
- Validate replica distribution: `sum(MachineSet.spec.replicas) = MachinePool.spec.replicas`

### Configuration Inspection
**Method:** Examine MachineSet zone configuration
- Extract zones field: `oc get machineset {name} -o jsonpath='{.spec.template.spec.providerSpec.value.zone}'`
- Non-zoned regions: field should be empty or absent
- Zoned regions: field should contain valid zone identifier
- Cross-reference zone values with Azure region capabilities

### Metadata Tracking
**Method:** Monitor resource metadata changes
- Track MachinePool resourceVersion on status updates
- Measure time from creation to Ready=True condition
- Count reconciliation loops via log entries or status update timestamps
- Verify MachineSet generation happens within expected time window (< 5 minutes)

### Recovery Path Validation
**Method:** Test error to healthy state transitions
- Create MachinePool in non-zoned region (previously would fail)
- Verify system reaches healthy state without manual intervention
- Confirm no manual cleanup required
- Validate automatic retry behavior if transient failures occur
