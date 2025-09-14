# CLAUDE.md

## EXECUTION EFFICIENCY REQUIREMENTS (MANDATORY)

**CRITICAL: Claude must execute with maximum efficiency and minimal verbosity**

**OUTPUT FORMAT ENFORCEMENT**:
- Each step must have explicit output: "✅ Step X completed: [result]"
- NEVER skip step numbers (Step 1 → Step 2 → Step 3 → Step 4)
- Silent thinking execution (no verbose explanations), but clear step completion markers

**VIOLATION ENFORCEMENT**: If Claude provides ANY explanatory text, process descriptions, or step-by-step commentary, user should immediately interrupt with "EFFICIENCY VIOLATION" and Claude must restart with direct execution only.

## MANDATORY Protocol

**CRITICAL: Each user request must be treated as a fresh execution: always re-read the YAML config and execute every step in full, regardless of conversation context.**  
**CRITICAL: NEVER use Task tool for agents** 

Before executing ANY user request, Claude MUST:
1. **MANDATORY: Read config/agents/workflow_orchestrator.yaml FIRST**
2. **MANDATORY: Let workflow orchestrator identify workflow automatically**
3. Read STARTUP_CHECKLIST.md (located at ./STARTUP_CHECKLIST.md in current directory)
4. Identify the correct agent YAML config file (via orchestrator or manual)
5. Read the agent YAML config directly using Read tool
6. Execute EACH step in the agent configuration manually using available tools
7. Verify each step's output before proceeding to next step

**MANDATORY** = Required, cannot be skipped

### CRITICAL: Rule Keyword Enforcement
**When reading any rule files, configuration files, or guideline documents:**
- **MANDATORY keywords = ABSOLUTE REQUIREMENT** - Must be executed without exception
- **NEVER keywords = ABSOLUTE PROHIBITION** - Must never be violated under any circumstances
- **CRITICAL keywords = STOP and verify** - Must pause and verify compliance before proceeding
- **FORBIDDEN keywords = ABSOLUTE PROHIBITION** - Same as NEVER, must not be violated
- If ANY of these keywords are violated → **STOP immediately and restart from beginning**

---

## Fresh Start Enforcement

**CRITICAL RULE: Every user request is a fresh execution, even if it happens in the SAME conversation.**

- MUST always re-read the correct agent YAML config before executing any agent.  
- MUST NOT reuse execution context or skip steps based on conversation history.  
- MUST verify that the "Read agent YAML" step has been executed before proceeding.  
- If the YAML has not been read in the current request → STOP and re-execute mandatory steps.  

### Implementation Notes
- Introduce a request-level hook: each user request = enforce mandatory startup protocol.  
- NEVER rely on Task Tool (causes step skipping and incomplete execution).  
- Log confirmation before agent execution:  
  `Agent config loaded: config/agents/<agent>.yaml`  

---

## Why NOT to use Task Tool
- Task tool uses general-purpose subagent that doesn't follow YAML configs exactly
- Steps get skipped, simplified, or executed incorrectly
- Results in incomplete or wrong outputs
- Agent configurations are designed for direct execution only

## Supported Agents

**Complete Flow**: test_case_generation → e2e_test_generation_openshift_private → test-executor → pr-submitter

### 1. test_case_generation
- **Description**: Generates test cases from JIRA issues for Polarion integration
- **Input**: JIRA issue key (e.g., "Create test case for HIVE-2883")
- **Output**: Test cases in markdown format
- **Config**: `config/agents/test_case_generation.yaml`

### 2. e2e_test_generation_openshift_private
- **Description**: Creates E2E test code integrated into openshift-tests-private repository
- **Input**: JIRA issue key (e.g., "Generate E2E test for HIVE-2883")
- **Output**: E2E test code in openshift-tests-private
- **Config**: `config/agents/e2e_test_generation_openshift_private.yaml`

<!--
### 2.1. e2e_test_generation_product_repo [Future]
- **Description**: Generates E2E test code directly in product repository (e.g., openshift/hive)
- **Note**: This agent will be enabled after E2E test case migration is completed
- **Config**: `config/agents/e2e_test_generation_product_repo.yaml`
-->

### 3. test-executor
- **Description**: Executes E2E tests and generates comprehensive test reports
- **Input**: JIRA issue key (e.g., "Run E2E tests for HIVE-2883")
- **Output**: Test execution results and reports
- **Config**: `config/agents/test-executor.yaml`

### 4. pr-submitter
- **Description**: Submits E2E test code as pull requests to openshift-tests-private
- **Input**: JIRA issue key (e.g., "Create PR for HIVE-2883")
- **Output**: Pull request created
- **Config**: `config/agents/pr-submitter.yaml`

## Agent Usage Examples

**CRITICAL: When a user requests to "generate test cases and run them", Claude MUST execute test_case_generation → e2e_test_generation_openshift_private → test-executor sequentially, performing all steps and verifying outputs at each stage, even if this request occurs in the same conversation. DO NOT STOP ANY STEPS.**

### MANDATORY STEP-BY-STEP EXECUTION PROTOCOL

**BEFORE STARTING ANY AGENT:**
1. **Read agent YAML config completely** - Use Read tool to load entire config file
2. **Execute steps in EXACT sequence** - Must complete step N before starting step N+1
3. **Verify each step's output** - Confirm step completion before proceeding to next step
4. **Use EXACT commands from config** - Copy commands exactly as written in agent YAML
5. **NEVER skip, simplify, or combine steps** - Each step must be executed individually

**VIOLATION PREVENTION RULES:**
- If ANY step is skipped → STOP immediately and restart from skipped step
- Each subtask completion MUST verify ALL steps executed before proceeding to next subtask
- If using different commands than specified in config → VIOLATION, must use exact config commands
- If combining multiple steps into one action → VIOLATION, must execute separately


### Example: Complete Flow for HIVE-2883
```
User: "Create test case for HIVE-2883"
→ Execute test_case_generation agent (with prerequisite check)

User: "Create E2E case for HIVE-2883"
→ Execute e2e_test_generation_openshift_private agent (with prerequisite check)

User: "Re-create test case for HIVE-2883"
→ Execute test_case_generation agent (SKIP prerequisite check, regenerate)

User: "Recreate E2E case for HIVE-2883"
→ Execute e2e_test_generation_openshift_private agent (SKIP E2E prerequisite check, regenerate)

User: "Run the E2E tests for HIVE-2883"
→ Execute test-executor agent

User: "Generate test cases and run them for HIVE-2883"
→ Execute test_case_generation → e2e_test_generation_openshift_private → test-executor

User: "Create a Pull Request for HIVE-2883 E2E tests"
→ Execute PR submission process

User: "Add QE comment to HIVE-2883"
→ Execute JIRA QE comment update process
```

### Regenerate Keywords Detection
**CRITICAL**: The following keywords in user input trigger REGENERATE mode (skip prerequisite checks):
- "re-create" / "recreate"
- "re-generate" / "regenerate"

**Examples of REGENERATE requests:**
- "re-create e2e case for HIVE-2923"
- "regenerate test case for HIVE-2883"
- "recreate E2E test for HIVE-2544"
- "re-create HIVE-2923 的 E2E test"

## Key Rules
- Always read agent YAML configs directly
- Execute each step manually using available tools
- Verify each step before proceeding
- Follow agent configurations exactly

