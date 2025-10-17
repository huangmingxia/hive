---
description: Generate comprehensive test cases from JIRA issues
argument-hint: [JIRA_KEY]
---

## Name
test:gen-case

## Synopsis
```
/test:gen-case JIRA_KEY
```

## Description
The `test:gen-case` command analyzes a JIRA issue and generates comprehensive test cases including requirements analysis, test strategy, coverage matrix, and both E2E and manual test case specifications.

## Implementation
Executes the test case generation agent at `.ai_testgen/config/agents/test_case_generation.md`

The agent performs:
- Fetches JIRA issue details via Snowflake MCP
- Analyzes requirements and acceptance criteria
- Generates test strategy, coverage matrix, and test case specifications
- Saves outputs to: `.ai_testgen/test_artifacts/{COMPONENT}/{JIRA_KEY}/test_cases/`

## Return Value
- **Success**: Test case files generated:
  - `test_requirements_output.md`
  - `test_strategy.md`
  - `test_coverage_matrix.md`
  - `{JIRA_KEY}_e2e_test_case.md`
  - `{JIRA_KEY}_manual_test_case.md`
- **Failure**: Error message with JIRA fetch or parsing errors

## Examples

1. **Basic usage**:
   ```
   /test:gen-case HIVE-2883
   ```

2. **Different component**:
   ```
   /test:gen-case CCO-1234
   ```

## Arguments
- **$1** (required): JIRA issue key (e.g., HIVE-2883, CCO-1234)

## Prerequisites
- JIRA MCP server configured and accessible
- Valid JIRA issue with requirements/acceptance criteria

## Complete Workflow
1. `/test:gen-case HIVE-XXXX` ← Start here
2. `/test:setup-fork HIVE-XXXX`
3. `/test:gen-code HIVE-XXXX`
4. `/test:run HIVE-XXXX`
5. `/test:gen-report HIVE-XXXX`

## See Also
- `test_case_generation.md` - Agent implementation
- `/test:gen-code` - Next step in workflow
