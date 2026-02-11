# Changelog

## v1.8.15

- Update dependencies

## v1.8.14

- Update dependencies

## v1.8.13

- Update dependencies

## v1.8.12

- Updated dependencies

## v1.8.11

- Updated dependencies
- Added an option to filter a multi alert monitor

## v1.8.10

- Handle event requests asynchronously, to avoid blocking the agent
- Update dependencies

## v1.8.9

- Use uid instead of name for user statement in Dockerfil
- Update dependencies

## v1.8.8

- Set new `Technology` property in extension description
- Update dependencies (go 1.23)

## v1.8.7

- Update dependencies (go 1.22)

## v1.8.6

- Update dependencies
- Use datadog API V2 for downtime creation

## v1.8.5

- Update dependencies

## v1.8.4

- update dependencies
- Fix warnings `Could not find step infos for step execution ...` in logs

## v1.8.3

- update dependencies

## v1.8.2

- Removed link to Steadybit homepage from event messages
- use discovery_kit_sdk for discoveries
- update dependencies

## v1.8.1

- Update dependencies
- Fix event creation when linux package is used

## v1.8.0

- Updated to use the new `steadybit/event-kit-api` version `1.3.0`
	- Requires Platform `>= 1.0.94`

## v1.7.10

- Possibility to exclude attributes from discovery

## v1.7.9

- expected status can be a list of status

## v1.7.8

- update dependencies
- added https_proxy support documentation

## v1.7.6

- migration to new unified steadybit actionIds and targetTypes

## v1.7.5

- update dependencies

## v1.7.4

- Add DateHappened to submitted DataDog events
- Correctly select StepExecution for event creation

## v1.7.3

- Add linux package build

## v1.7.2

- Added service tag to Datadog events

## v1.7.1

- Added DEBUG logging for monitor discovery

## v1.7.0

- Links to Datadogs monitors are now using the timeframe of the experiment execution.
- "Monitor Status Check" has a new parameter `Status Check Mode`. Supported values are `All the time` (default)
	and `At least once`.
- New Action to create a Downtime for a monitor during an experiment execution.
- Details about step executions are sent to Datadog as events.

## v1.6.0

- Monitor shouldn't have a blast radius
- Run as non-root user
- Update dependencies

## v1.5.0

- Print build information on extension startup.

## v1.4.0

- Support creation of a TLS server through the environment variables `STEADYBIT_EXTENSION_TLS_SERVER_CERT`
	and `STEADYBIT_EXTENSION_TLS_SERVER_KEY`. Both environment variables must refer to files containing the certificate
	and key in PEM format.
- Support mutual TLS through the environment variable `STEADYBIT_EXTENSION_TLS_CLIENT_CAS`. The environment must refer
	to a comma-separated list of files containing allowed clients' CA certificates in PEM format.

## v1.3.0

- Support for the `STEADYBIT_LOG_FORMAT` env variable. When set to `json`, extensions will log JSON lines to stderr.

## v1.2.1

- Also observe the events `experiment.execution.failed`, `experiment.execution.canceled`
	and `experiment.execution.errored` to report all relevant event types to Datadog.

## v1.2.0

- Reports events to Datadog to mark the start and end of experiments.

## v1.1.0

- Correctly mark duration parameter for status check action as required.
- Add monitor status widgets to the execution view.

## v1.0.0

- Initial release
