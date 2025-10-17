---
description: Setup GitHub fork and clone openshift-tests-private repository
argument-hint: [JIRA_KEY]
---

## Name
test:setup-fork

## Synopsis
```
/test:setup-fork JIRA_KEY
```

## Description
The `test:setup-fork` command prepares the openshift-tests-private repository for E2E test generation by setting up the GitHub fork, cloning the repository, configuring remotes, and creating a working branch.

## Implementation
Executes the repository setup agent at `.ai_testgen/config/agents/repository_setup_agent.md`

The agent performs:
- Detects fork owner from `.env` or `gh` CLI
- Verifies/creates fork using `gh repo fork`
- Clones repository to: `.ai_testgen/test_artifacts/{COMPONENT}/{JIRA_KEY}/openshift-tests-private`
- Syncs with upstream and creates working branch: `ai-e2e-{JIRA_KEY}`

## Return Value
- **Success**: Repository cloned and configured
- **Failure**: Error message with troubleshooting steps

## Examples

1. **Basic usage**:
   ```
   /test:setup-fork HIVE-2883
   ```

2. **With custom fork owner**:
   ```bash
   echo "GITHUB_FORK_OWNER=my-org-name" > .ai_testgen/.env
   /test:setup-fork CCO-1234
   ```

## Arguments
- **$1** (required): JIRA issue key (e.g., HIVE-2883, CCO-1234)

## Prerequisites
- `gh` CLI installed and authenticated (`gh auth login`)
- Network access to GitHub

## Next Steps
- `/test:gen-case {JIRA_KEY}` - Generate test cases
- `/test:gen-code {JIRA_KEY}` - Generate E2E test code

## See Also
- `repository_setup_agent.md` - Agent implementation
