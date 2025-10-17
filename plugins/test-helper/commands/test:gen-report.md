---
description: Generate comprehensive test report with execution results
argument-hint: [JIRA_KEY]
---

## Name
test:gen-report

## Synopsis
```
/test:gen-report JIRA_KEY
```

## Description
The `test:gen-report` command creates a comprehensive test report that aggregates all test artifacts including test cases, generated code, execution results, and quality metrics.

## Implementation
Executes the test report generation agent at `.ai_testgen/config/agents/test_report_generation.md`

The agent performs:
- Collects artifacts from: `.ai_testgen/test_artifacts/{COMPONENT}/{JIRA_KEY}/`
- Analyzes test cases, code quality, and execution results
- Generates comprehensive markdown report
- Saves to: `.ai_testgen/test_artifacts/{COMPONENT}/{JIRA_KEY}/test_report.md`

## Return Value
- **Success**: Comprehensive test report generated
- **Failure**: Missing artifacts or report generation errors

## Examples

1. **Basic usage (after full workflow)**:
   ```
   /test:gen-report HIVE-2883
   ```

2. **After test execution**:
   ```
   /test:run HIVE-2883
   /test:gen-report HIVE-2883
   ```

3. **Partial report (without execution)**:
   ```
   /test:gen-case HIVE-2883
   /test:gen-report HIVE-2883
   ```

## Arguments
- **$1** (required): JIRA issue key (e.g., HIVE-2883, CCO-1234)

## Prerequisites
- Minimal: Test cases generated
- Optional: E2E code generated
- Optional: Tests executed

## Complete Workflow
1. `/test:gen-case HIVE-XXXX`
2. `/test:setup-fork HIVE-XXXX`
3. `/test:gen-code HIVE-XXXX`
4. `/test:run HIVE-XXXX`
5. `/test:gen-report HIVE-XXXX` ← Final step

## Report Contents
- JIRA Issue Summary
- Test Case Analysis
- Code Generation Metrics
- Execution Results
- Artifacts and Links
- Recommendations

## See Also
- `test_report_generation.md` - Agent implementation
- All other commands - Report aggregates their outputs
