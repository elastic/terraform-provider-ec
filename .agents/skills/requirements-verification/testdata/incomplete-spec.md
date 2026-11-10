<!--
Intentionally invalid. Not a real capability spec.
Used only to exercise `.agents/skills/requirements-verification`.
Do not copy into openspec/specs/. Do not "fix" this file.
make check-openspec does not read this path.
-->

# `ec_example_incomplete`

This fixture has no `Resource implementation:` line.

## Schema

```hcl
resource "ec_example_incomplete" "example" {
  name = <optional, string>
}
```

## Requirements

### Requirement: Name is required

The resource uses name as the identifier.

### Requirement: Read something

The resource SHALL refresh state from the API.

#### Scenario: Read

- GIVEN an existing remote object
- WHEN Read is called
- THEN the provider SHALL set `name` in state
