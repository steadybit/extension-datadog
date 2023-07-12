# Changelog

## v1.7.0

 - Links to Datadogs monitors are now using the timeframe of the experiment execution.
 - "Monitor Status Check" has a new parameter `Status Check Mode`. Supported values are `All the time` (default) and `At least once`.
 - New Action to create a Downtime for a monitor during an experiment execution.
 - Details about step executions are sent to Datadog as events.

## v1.6.0

 - Monitor shouldn't have a blast radius
 - Run as non-root user
 - Update dependencies
 
## v1.5.0

 - Print build information on extension startup.

## v1.4.0

 - Support creation of a TLS server through the environment variables `STEADYBIT_EXTENSION_TLS_SERVER_CERT` and `STEADYBIT_EXTENSION_TLS_SERVER_KEY`. Both environment variables must refer to files containing the certificate and key in PEM format.
 - Support mutual TLS through the environment variable `STEADYBIT_EXTENSION_TLS_CLIENT_CAS`. The environment must refer to a comma-separated list of files containing allowed clients' CA certificates in PEM format.

## v1.3.0

 - Support for the `STEADYBIT_LOG_FORMAT` env variable. When set to `json`, extensions will log JSON lines to stderr.

## v1.2.1

 - Also observe the events `experiment.execution.failed`, `experiment.execution.canceled` and `experiment.execution.errored` to report all relevant event types to Datadog.

## v1.2.0

 - Reports events to Datadog to mark the start and end of experiments.

## v1.1.0

 - Correctly mark duration parameter for status check action as required.
 - Add monitor status widgets to the execution view. 

## v1.0.0

 - Initial release