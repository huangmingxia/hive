---
description: Execute E2E tests and capture results
argument-hint: [JIRA_KEY]
---

## Name
test:run

## Synopsis
```
/test:run JIRA_KEY
```

## Description
The `test:run` command executes generated E2E tests against an OpenShift cluster and captures detailed execution results.

## Implementation
Executes the test executor agent at `.ai_testgen/config/agents/test-executor.md`

The agent performs:
- Compiles E2E test code from the forked repository
- Executes tests against OpenShift cluster (requires kubeconfig)
- Captures execution results, logs, and artifacts
- Saves outputs to: `.ai_testgen/test_artifacts/{COMPONENT}/{JIRA_KEY}/test_results/`

## Return Value
- **Success**: Test execution completed with results captured
- **Failure**: Compilation errors, cluster access issues, or test failures

## Examples

1. **Basic usage**:
   ```
   /test:run HIVE-2883
   ```

2. **After code generation**:
   ```
   /test:gen-code HIVE-2883
   /test:run HIVE-2883
   ```

## Arguments
- **$1** (required): JIRA issue key (e.g., HIVE-2883, CCO-1234)

## Prerequisites
- E2E test code generated
- OpenShift cluster access with valid kubeconfig
- Go toolchain installed for compilation

## Complete Workflow
1. `/test:gen-case HIVE-XXXX`
2. `/test:setup-fork HIVE-XXXX`
3. `/test:gen-code HIVE-XXXX`
4. `/test:run HIVE-XXXX` ← This command
5. `/test:gen-report HIVE-XXXX`

## See Also
- `test-executor.md` - Agent implementation
- `/test:gen-report` - Next step in workflow
