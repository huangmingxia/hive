# Claude Code Slash Commands

This directory contains custom slash commands for AI-enhanced test generation workflow.

## Available Commands

### 1. `/setup-fork <github-username>`
**Purpose**: Configure your GitHub fork for E2E test generation

**When to use**: Run this ONCE before generating E2E tests

**Example**:
```
/setup-fork mihuang
```

**What it does**:
- Creates `.ai_testgen/.env` file
- Sets `GITHUB_FORK_OWNER` to your GitHub username
- Verifies your fork of openshift-tests-private exists

**Prerequisites**: You must have forked https://github.com/openshift/openshift-tests-private

---

### 2. `/generate-test-case <JIRA-KEY>`
**Purpose**: Generate comprehensive test cases from JIRA issues

**Example**:
```
/generate-test-case HIVE-2883
```

**What it does**:
- Fetches JIRA issue details
- Analyzes feature requirements
- Generates test strategy and requirements
- Creates E2E and manual test cases
- Builds test coverage matrix

**Outputs**:
- `test_requirements_output.md`
- `test_strategy.md`
- `test_coverage_matrix.md`
- `{JIRA_KEY}_e2e_test_case.md`
- `{JIRA_KEY}_manual_test_case.md`

**Prerequisites**: Valid JIRA issue key

---

### 3. `/generate-e2e-code <JIRA-KEY>`
**Purpose**: Generate executable E2E test code for openshift-tests-private

**Example**:
```
/generate-e2e-code HIVE-2883
```

**What it does**:
- Sets up openshift-tests-private repository from your fork
- Validates existing test coverage
- Generates E2E test code for all detected platforms (AWS, Azure, GCP, etc.)
- Performs quality checks and compilation validation

**Outputs**:
- Test code in `.ai_testgen/test_artifacts/{COMPONENT}/{JIRA_KEY}/openshift-tests-private/`
- Validation reports in `phases/` directory

**Prerequisites**:
- Fork configured via `/setup-fork`
- Test case file must exist (run `/generate-test-case` first)

---

### 4. `/run-e2e-tests <JIRA-KEY>`
**Purpose**: Execute E2E tests and capture results

**Example**:
```
/run-e2e-tests HIVE-2883
```

**What it does**:
- Navigates to test repository
- Builds test binaries
- Executes E2E tests
- Captures execution results and logs

**Prerequisites**:
- E2E code generated and committed
- Fork configured
- Access to OpenShift cluster (kubeconfig)

---

### 5. `/generate-test-report <JIRA-KEY>`
**Purpose**: Generate comprehensive test report

**Example**:
```
/generate-test-report HIVE-2883
```

**What it does**:
- Collects all test artifacts
- Summarizes test execution results
- Documents bugs found (product and E2E)
- Generates risk assessment
- Creates defect statistics

**Outputs**:
- `test_report.md` with complete test summary

**Prerequisites**:
- Test cases generated (optional)
- E2E code generated (optional)
- Test execution results (optional)

---

## Complete E2E Test Generation Workflow

### First Time Setup
```bash
# 1. Fork the repository on GitHub
#    Go to: https://github.com/openshift/openshift-tests-private
#    Click "Fork"

# 2. Configure your fork
/setup-fork your-github-username
```

### For Each JIRA Issue
```bash
# 1. Generate test cases
/generate-test-case HIVE-2883

# 2. Generate E2E test code (uses your fork)
/generate-e2e-code HIVE-2883

# 3. Run the tests (requires cluster access)
/run-e2e-tests HIVE-2883

# 4. Generate comprehensive report
/generate-test-report HIVE-2883
```

## Architecture

All commands follow this pattern:
1. Read agent configuration from `.ai_testgen/config/agents/`
2. Execute steps defined in agent config
3. Generate outputs to `.ai_testgen/test_artifacts/{COMPONENT}/{JIRA_KEY}/`

## Configuration Files

- **Agent configs**: `.ai_testgen/config/agents/*.md`
- **Rules**: `.ai_testgen/config/rules/`
- **Templates**: `.ai_testgen/config/templates/`
- **Environment**: `.ai_testgen/.env` (your fork configuration)

## Output Structure

```
.ai_testgen/test_artifacts/{COMPONENT}/{JIRA_KEY}/
├── phases/
│   ├── test_requirements_output.md
│   ├── test_strategy.md
│   └── validation_results.md
├── test_cases/
│   ├── {JIRA_KEY}_e2e_test_case.md
│   └── {JIRA_KEY}_manual_test_case.md
├── test_coverage_matrix.md
├── test_report.md
└── openshift-tests-private/          # Your fork clone
    └── test/extended/cluster_operator/{COMPONENT}/
        └── generated test files
```

## Troubleshooting

### "Fork not configured" error
**Solution**: Run `/setup-fork your-github-username`

### "Test case file missing" error
**Solution**: Run `/generate-test-case HIVE-XXXX` first

### "Repository not found" error
**Solution**:
1. Verify fork exists on GitHub
2. Check `.ai_testgen/.env` has correct `GITHUB_FORK_OWNER`
3. Re-run `/setup-fork your-github-username`

### "Compilation failed" error
**Solution**: The quality check agent will auto-fix up to 3 times. If it still fails, check the error logs in the phases directory.

## Tips

1. **Run setup once**: `/setup-fork` is a one-time setup per machine
2. **Sequential workflow**: Always run commands in order (test-case → e2e-code → run-tests → report)
3. **Check prerequisites**: Each command lists required prerequisites
4. **Review outputs**: Check generated files before proceeding to next step
5. **Use your fork**: All E2E code changes go to YOUR fork, not upstream

## Need Help?

- **Agent configs**: See `.ai_testgen/config/agents/` for detailed agent instructions
- **Examples**: Check `.ai_testgen/test_artifacts/hive/` for example outputs
- **Rules**: Review `.ai_testgen/config/rules/` for test generation guidelines
