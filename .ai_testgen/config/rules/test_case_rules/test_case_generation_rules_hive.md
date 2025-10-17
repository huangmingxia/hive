# Hive MANDATORY Rules Validation Checklist

## Rules Summary
**Source**: `config/rules/test_case_rules/test_case_generation_rules_hive.md` Line 14-21

1. **Platform-Based**: OpenStack, IBMCloud, Nutanix → Manual
2. **Scenario-Based**: Upgrade/retrofit/legacy, Shared VPC, Resource Group → Manual  
3. **DNSZone Pattern**: Use hiveutil + managedDNS, NOT direct DNSZone creation

---

## Generation-Time Validation 

### Before Writing Test Cases - Verify Classification

**For each test scenario generated, check:**

1. **Platform Check**:
   - If platform is **OpenStack, IBMCloud, or Nutanix** → **MUST be Manual**
   - If platform is **AWS, Azure, Azure Government, GCP, vSphere** → **Default to E2E**  

2. **Scenario Check**:
   - If keywords: **upgrade, retrofit, legacy** → Classify as **Manual**
   - If requires: **HostedZoneRole, NetworkProjectID, ResourceGroupName** → Classify as **Manual**
   - Otherwise → Can be **E2E**

3. **DNSZone Check**:
   - If test involves DNSZone → Must use **hiveutil + managedDNS** pattern
   - Must NOT use direct DNSZone creation

4. **MANDATORY COMMAND RULE**: For test step commands, ONLY use oc with jsonpath output format. FORBIDDEN: jq, python -m json.tool, or any other JSON tools. Use oc jsonpath exclusively (e.g., oc get secret -o jsonpath='{.data.metadata\\.json}' for extraction, oc get -o jsonpath for field validation)."


### Pre-Write Verification Question

**Before writing to E2E file, ask:**
- "Does this test involve OpenStack/IBMCloud/Nutanix?" → YES = Move to Manual
- "Does this test involve upgrade/retrofit/legacy?" → YES = Move to Manual
- "Does this test require shared VPC or existing resource group?" → YES = Move to Manual

**Only write to E2E file if ALL answers are NO**

---

## Post-Generation Validation (For fixing errors)

**After test cases generated, verify files:**

```bash
JIRA_KEY="HIVE-2302"  # Replace with actual key
E2E_FILE="test_artifacts/hive/${JIRA_KEY}/test_cases/${JIRA_KEY}_e2e_test_case.md"

# Check for violations in E2E file
echo "Checking for platform violations..."
grep -E "Platform.*(OpenStack|IBMCloud|Nutanix)" "$E2E_FILE"

echo "Checking for scenario violations..."
grep -iE "upgrade|retrofit|legacy|HostedZoneRole|NetworkProjectID|ResourceGroupName" "$E2E_FILE"
```

**Expected**: Both should return empty (no matches found)

**If violations found**: Use "Quick Fix" section below to move test cases to Manual file

---

## MANDATORY Rules Detail

### Rule 1: Platform-Based Classification
**OpenStack, IBMCloud, Nutanix** → Must be Manual (require special infrastructure)

### Rule 2: Scenario-Based Classification
- **Upgrade/Retrofit/Legacy** → Manual (Upgrade need version control)
- **AWS HostedZoneRole** → Manual (need shared VPC setup)
- **GCP NetworkProjectID** → Manual (need shared VPC setup)
- **Azure ResourceGroupName** → Manual (need pre-created RG)

### Rule 3: DNSZone Test Pattern
- ✅ Use `hiveutil` to configure managedDomains in HiveConfig
- ✅ Create ClusterDeployment with managedDNS=true (auto-creates DNSZone)
- ⛔ FORBIDDEN: Direct DNSZone creation
- ✅ Reference: e2e case 24088

---

## Quick Fix

If violations found:
1. Extract test case from E2E file
2. Change **Type: E2E** → **Type: Manual**
3. Renumber as **HIVE-XXXX_M0XX**
4. Add to Manual file
5. Remove from E2E file

---

## Validation Script

```bash
#!/bin/bash
# validate_hive_rules.sh

JIRA_KEY=$1
E2E_FILE="test_artifacts/hive/${JIRA_KEY}/test_cases/${JIRA_KEY}_e2e_test_case.md"

VIOLATIONS=$(grep -E "Platform.*(OpenStack|IBMCloud|Nutanix)" "$E2E_FILE" 2>/dev/null)

if [ -n "$VIOLATIONS" ]; then
    echo "❌ VIOLATION: OpenStack/IBMCloud/Nutanix in E2E file"
    echo "$VIOLATIONS"
    exit 1
else
    echo "✅ All rules validated"
    exit 0
fi
```

**Usage**: `./validate_hive_rules.sh HIVE-2302`
