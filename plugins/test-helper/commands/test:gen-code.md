---
description: Generate E2E test code for openshift-tests-private repository
argument-hint: [JIRA_KEY]
---

## Name
test:gen-code

## Synopsis
```
/test:gen-code JIRA_KEY
```

## Description
The `test:gen-code` command generates executable Ginkgo-based E2E test code for the openshift-tests-private repository based on test case specifications.

## Implementation
Executes the E2E test generation agent at `.ai_testgen/config/agents/e2e_test_generation_openshift_private.md`

The agent performs:
- Reads test case from: `.ai_testgen/test_artifacts/{COMPONENT}/{JIRA_KEY}/test_cases/{JIRA_KEY}_e2e_test_case.md`
- Analyzes repository patterns in your fork
- Generates Ginkgo E2E test code
- Saves to: `.ai_testgen/test_artifacts/{COMPONENT}/{JIRA_KEY}/openshift-tests-private/`

## Return Value
- **Success**: E2E test code generated in forked repository
- **Failure**: Error message with generation or validation errors

## Examples

1. **Basic usage**:
   ```
   /test:gen-code HIVE-2883
   ```

2. **After test case generation**:
   ```
   /test:gen-case HIVE-2883
   /test:gen-code HIVE-2883
   ```

## Arguments
- **$1** (required): JIRA issue key (e.g., HIVE-2883, CCO-1234)

## Prerequisites
- Test case file exists
- Repository set up via `/test:setup-fork`
- Fork configured in `.ai_testgen/.env`

## Complete Workflow
1. `/test:gen-case HIVE-XXXX`
2. `/test:setup-fork HIVE-XXXX`
3. `/test:gen-code HIVE-XXXX` ← This command
4. `/test:run HIVE-XXXX`
5. `/test:gen-report HIVE-XXXX`

## See Also
- `e2e_test_generation_openshift_private.md` - Agent implementation
- `/test:run` - Next step in workflow
