# 🧪 Test Generation Assistant

## 🎯 Overview

AI-assisted test generation and execution system that automates the complete testing workflow:

- **Test Case Generation**: Generate comprehensive test cases from JIRA tickets
- **E2E Test Code Generation**: Create executable E2E test code
- **Test Execution**: Run E2E tests and capture results
- **PR Submission**: Automatically create pull requests for generated test code
- **JIRA Integration**: Update JIRA tickets with test status and results

This system uses multiple specialized agents to handle each phase of the testing process, ensuring consistent and thorough test coverage for OpenShift components.

## 📋 Prerequisites

Before you can use this test generation system, you need to set up the following integrations:

### 1. Install Claude Code, Cursor, or Gemini
- **Claude Code access**: Follow [Claude Code Install Instructions](https://docs.google.com/document/d/1eNARy9CI28o09E7Foq01e5WD5MvEj3LSBnXqFcprxjo/edit?usp=drivesdk)
- **Cursor access**: Download and install [Cursor](https://cursor.sh/) for an IDE-based experience with AI assistance
- **Gemini access**: Refer to [Gemini documentation](https://gemini.google.com/) for AI assistance
- **Important**: Choose Claude Code for command-line interaction, Cursor for a full IDE experience, or Gemini for alternative AI assistance. All can be used to run the test generation system.

### 2. JIRA MCP Access
- **Apply for jira-mcp-snowflake token**: Refer to [Jira-MCP-Snowflake-Token-Guide](https://docs.google.com/document/d/1pg6TkwezhIahppp5k0md0Zx0CC-4f_RWQHaH9cTl1Mo/edit?tab=t.0#heading=h.xyjdx8nsdjql)
 
- **Config jira-mcp-snowflake mcp server**:
  ```bash
  claude mcp add jira-mcp-snowflake https://jira-mcp-snowflake.mcp-playground-poc.devshift.net/sse --transport sse -H "X-Snowflake-Token: your_token_here"
  ```
### 3. DeepWiki MCP Connection 
  ```bash
  claude mcp add -s user -t http deepwiki https://mcp.deepwiki.com/mcp
  ```
### 4. Prepare Test Generation Rules
- **Purpose**: Configure rules to guide AI in generating realistic and executable test cases
- **Location**: `config/rules/`
- **Content**: Component-specific testing guidelines, validation criteria, and best practices

### 5. Configure GitHub CLI (gh)
- **Purpose**: Enable PR creation and GitHub repository operations
- **Installation**: 
  ```bash
  # macOS
  brew install gh
  
  # Linux/Windows
  # Follow: https://cli.github.com/manual/installation
  ```
  
### 6. Configure E2E Repository Fork
**Purpose**: E2E test generation requires write access to openshift-tests-private repository  
**Requirement**: Update the repository URL in the agent configuration to use your own fork

**Steps to configure:**
```
Edit `config/agents/e2e_test_generation_openshift_private.yaml`: 
"If not exists: Clone https://github.com/YOUR_USERNAME/openshift-tests-private.git to temp_repos/openshift-tests-private/"
```
**Verify access**: Ensure you have write permissions to your fork for PR creation

### 7. Execution Method

**CRITICAL**: All agents must be executed by reading agent YAML configurations directly.  

**Execution Pattern**:
1. Read `CLAUDE.md` (for Claude Code), `GEMINI.md` (for Gemini), or Cursor documentation (see Cursor Configuration in Prerequisites) for agent instructions
2. Read specific agent YAML config (e.g., `config/agents/test_case_generation.yaml`)
3. Execute each step manually using available tools
4. Verify each step before proceeding

## 🎛️ Workflow Orchestrator (PRIMARY ENTRY POINT)

**MANDATORY: The workflow orchestrator is the PRIMARY entry point for ALL test generation requests.**

**CRITICAL: Before executing ANY user request, AI MUST:**
1. **Read `config/agents/workflow_orchestrator.yaml` FIRST**
2. **Let orchestrator identify workflow automatically**
3. **Execute agents in correct sequence**

### What is it?
A top-level agent that automatically:
- ✅ Parses user requests and identifies intent
- ✅ Matches requests to predefined workflows
- ✅ Validates prerequisites automatically
- ✅ Executes agents in correct sequence
- ✅ Handles REGENERATE mode
- ✅ Generates comprehensive execution reports

### Configuration
**Location**: `config/agents/workflow_orchestrator.yaml`  
**Guide**: `config/agents/WORKFLOW_ORCHESTRATOR_GUIDE.md`

### Rule Keyword Enforcement
When AI reads any configuration or guideline files:
- **MANDATORY** = ABSOLUTE REQUIREMENT (must be executed without exception)
- **NEVER** = ABSOLUTE PROHIBITION (must never be violated)
- **CRITICAL** = STOP and verify (must pause before proceeding)
- **FORBIDDEN** = ABSOLUTE PROHIBITION (same as NEVER)
- Violation of any keyword = Complete execution failure, must restart

### Supported Workflows
1. **full_flow** - Test Case → E2E Code → Execution
2. **test_case_only** - Test Case Generation
3. **e2e_generation** - E2E Test Generation
4. **test_execution** - Test Execution Only
5. **pr_submission** - PR Submission
6. **e2e_and_run** - E2E Generation + Execution

### Quick Start
The orchestrator automatically identifies your intent from natural language:
```bash
# Complete flow (3 agents executed automatically)
"Generate test cases and run them for HIVE-2883"

# Single agent via orchestrator
"Create test case for HIVE-2883"

# Regenerate mode (skips prerequisite checks)
"re-create e2e test for HIVE-2883"
```

**See** `config/agents/WORKFLOW_ORCHESTRATOR_GUIDE.md` for detailed usage instructions.

---

## 🚀  Agents

### 1. Generate Test Cases
**Purpose**: Generate manual test cases from JIRA tickets  
**Agent Config**: `config/agents/test_case_generation.yaml`  
**Input**: JIRA issue key (e.g., "HIVE-2883")  
**Output**: 
- `test_requirements_output.yaml` - Detailed requirements analysis
- `test_strategy.yaml` - Test strategy based on requirements  
- `{JIRA_KEY}_test_case.md` - Executable test cases in markdown format

**Example Usage**:
```
"Create test case for HIVE-2883"
"Generate test case for JIRA issue HIVE-2883"
"Generate test cases and run them for HIVE-2883"  # Full flow
```

**Features**:
- Systematic thinking framework (4 mandatory phases)
- JIRA data extraction and analysis
- Architecture analysis with DeepWiki integration
- User-reality driven test scenarios
- Quantitative validation methods  

### 2. Generate E2E Test Code
**Purpose**: Generate E2E test code based on existing test cases  
**Agent Config**: `config/agents/e2e_test_generation_openshift_private.yaml`  
**Output**: E2E test code integrated into openshift-tests-private repository  

### 2.1 Execute E2E Tests
**Purpose**: Execute generated E2E tests and capture results  
**Agent Config**: `config/agents/test-executor.yaml`  
**Output**: Test execution results and reports  

### 3. Submit E2E Pull Request
**Purpose**: Create PR for generated E2E tests  
**Tools**: Git, GitHub CLI (gh)  
**Output**: Created pull request with E2E test code  

### 4. Update JIRA QE Comment [In Progress]


### Agent Dependencies
- test-executor requires completion of e2e_test_generation_openshift_private
- PR submission requires completion of e2e_test_generation_openshift_private (and optionally test-executor)
- JIRA QE comment update can be executed after any agent completion

## 🚀 Usage Options

### Option 1: Claude Code Usage

#### Quick Start

1. **Setup Prerequisites** (see above sections):
   - Install Claude Code
   - Configure JIRA MCP access
   - Configure DeepWiki MCP
   - Configure GitHub CLI (gh)

2. **Open Claude Code**:
   ```bash
   claude
   ```

3. **Navigate to project directory**:
   ```bash
   cd /path/to/.ai_testgen
   ```

4. **Execute agents directly**:
   ```
   # Generate test cases
   "Create test case for HIVE-2883"
   
   # Generate E2E code
   "Generate E2E case for HIVE-2883"
   
   # Execute E2E tests
   "Run E2E tests for HIVE-2883"

   # Complete end-to-end flow (executes all 3 agents in sequence)
   "Generate test cases and run them for HIVE-2883"
   # This automatically executes: test_case_generation → e2e_test_generation_openshift_private → test-executor
   
   # Create PR
   "Create PR for HIVE-2883 E2E tests"
   
   # Add QE comment
   "Add QE comment to HIVE-2883"
   ```

### How It Works

1. **Claude reads CLAUDE.md** - Gets agent instructions
2. **Reads agent YAML configs** - Gets specific execution steps  
3. **Executes steps** - Uses available tools (JIRA MCP, WebFetch, Bash, etc.)
4. **Verifies outputs** - Ensures each step completes successfully
5. **Generates artifacts** - Creates test cases, E2E code, PRs, etc.

### No API Server Needed
- Everything runs through Claude Code directly
- No need for separate microservice or Docker containers
- All agents executed via natural language commands
- Real-time interaction and feedback

### Option 2: Cursor IDE Usage

#### Quick Start

1. **Setup Prerequisites** (see above sections):
   - Install Cursor IDE
   - Configure JIRA MCP access (if using Cursor's AI features)
   - Configure DeepWiki MCP
   - Configure GitHub CLI (gh)

2. **Open Cursor IDE**:
   ```bash
   cursor /path/to/ai_testgen
   ```

3. **Use Cursor's AI Chat**:
   - Open Cursor's AI chat panel (Cmd+L or Ctrl+L)
   - Execute test generation commands directly in the chat
   - Use Cursor's integrated terminal for command execution

4. **Execute agents directly**:
   ```
   # Generate test cases
   "Create test case for HIVE-2883"
   
   # Generate E2E code
   "Generate E2E case for HIVE-2883"
   
   # Execute E2E tests
   "Run E2E tests for HIVE-2883"

   # Complete end-to-end flow (executes all 3 agents in sequence)
   "Generate test cases and run them for HIVE-2883"
   # This automatically executes: test_case_generation → e2e_test_generation_openshift_private → test-executor
   
   # Create PR
   "Create PR for HIVE-2883 E2E tests"
   
   # Add QE comment
   "Add QE comment to HIVE-2883"
   ```

#### Cursor Advantages
- **Integrated Development Environment**: Full IDE experience with AI assistance
- **Code Navigation**: Easy file browsing and code navigation
- **Terminal Integration**: Built-in terminal for command execution
- **AI-Powered Editing**: Smart code completion and refactoring
- **Project Management**: Better project structure visualization

## 🔗 Integrations

### JIRA Integration
- **MCP Endpoint**: `jira-mcp-snowflake`
- **Purpose**: Extract ticket requirements and details

### DeepWiki Integration
- **MCP Endpoint**: `DeepWiki MCP`
- **Purpose**: Code change analysis and testing implications

### GitHub Integration
- **Repository Management**: Automatic branch creation
- **File Operations**: Direct file creation in E2E repositories

## 📚 Documentation

- **[CLAUDE.md](CLAUDE.md)** - Complete agent instructions and execution guide
