# Project Instructions

## Mandatory Verification Steps

Before marking any software engineering task as complete, especially those involving code modifications, you **must** execute the following commands to ensure code quality and adherence to project standards:

1.  **Formatting & Linting:**
    Execute the following command to format code and run linting checks:
    ```bash
    invoke formatting
    ```

2.  **Type Checking:**
    Execute the following command to perform static type analysis:
    ```bash
    invoke typecheck
    ```

If either of these commands fails, fix the reported issues before proceeding.
