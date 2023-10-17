# Patching Automation Tool - Development Notes

## Replacing a go module

In the event that a go module does not work as desired, we can temporarily replace
the module with our own fork.

* Create a fork of the repo and update with the fix desired
* Replace the module (temporarily?) until the PR has been merged.
  * `go mod edit -replace ${ORIG_GO_MOD}=${NEW_GO_MOD}`
  * `go mod tidy`

Example:

```bash
go mod edit -replace github.com/puppetlabs/go-pe-client=github.com/TJM/go-pe-client@cmd_task_options
go mod tidy
```
