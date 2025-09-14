# AI-Assisted Test Generation

This proposal aims to integrate an AI-assisted test generation system into the Hive upstream repository.  
To achieve this, we suggest creating a dedicated directory `.ai_testgen` within the repository to store prompt files and related resources.  
This ensures that contributors who already have the Hive repository can directly leverage these assets without the need to clone or maintain a separate repository.  

## Two possible implementation options are described below:  

**1. Clone external repository**  
- Migrate the contents of `git@github.com:huangmingxia/test-generation-assistant.git` into the `.ai_testgen` folder.  
    ```bash
    git clone git@github.com:huangmingxia/test-generation-assistant.git
    ```
- Populate `.ai_testgen/config/rules` with upstream-specific rules to be used for generating test cases and end-to-end (e2e) tests.


**2. Rules-only integration**  
- Only add the `rules/` directory inside `.ai_testgen/rules`.  
- Configure test case or e2e generation to load these rules by default.   
**Limitation**:  
  - For pre-merge testing, `jira-mcp-snowflake` currently lacks support for extracting associated PR links from Jira issues.    
  - Additionally, DeepWiki is updated on a weekly basis, which means that code changes within this interval cannot be retrieved.    

### Recommended approach [TBD]
Currently recommend approach 1 (cloning the external repository), as it ensures all code changes, including pre-merge changes, are available during test generation.
