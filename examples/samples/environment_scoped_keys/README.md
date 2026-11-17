# Environment-scoped Management Keys

Gives the automation of each Environment a key which can only reach its own Environment.

1. [central](central/main.tf) uses a team-scoped Management Key to create the Environments, their Ingest Keys, and an environment-scoped Management Key for each of them.
2. [environment](environment/main.tf) is the automation of one Environment, which is given only the key from the first step.

## What an environment-scoped key can do today

* It authenticates the provider, and `honeycombio_auth_metadata` reports its team and Environment.
* It cannot manage any resource of the provider yet. Those use either the v1 API, which needs a Configuration Key until environment-scoped keys are accepted there, or team-scoped v2 routes. A resource which needs a team-scoped key fails with an error which says so.
* It cannot create other API Keys.
