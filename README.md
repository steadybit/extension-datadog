<img src="./logo.png" height="130" align="right" alt="Datadog logo depicting a dog with the text 'Datadog'">

# Steadybit extension-datadog

A [Steadybit](https://www.steadybit.com/) check implementation for data exposed through Datadog.

## Capabilities

 - Check monitor status

## Deployment

We recommend that you deploy the extension with our [official Helm chart](https://github.com/steadybit/helm-charts/tree/main/charts/steadybit-extension-datadog).

## Agent Configuration

**Note:** When deployed in Kubernetes using our [official Helm chart](https://github.com/steadybit/helm-charts/tree/main/charts/steadybit-extension-datadog), this is not necessary because the extension can be auto-discovered.

The Steadybit agent needs to be configured to interact with the Datadog extension by adding the following environment variables:

```shell
# Make sure to adapt the URLs and indices in the environment variables names as necessary for your setup

STEADYBIT_AGENT_ACTIONS_EXTENSIONS_0_URL=http://steadybit-extension-datadog.steadybit-extension.svc.cluster.local:8090
STEADYBIT_AGENT_DISCOVERIES_EXTENSIONS_0_URL=http://steadybit-extension-datadog.steadybit-extension.svc.cluster.local:8090
STEADYBIT_AGENT_EVENTS_EXTENSIONS_0_URL=http://steadybit-extension-datadog.steadybit-extension.svc.cluster.local:8090
```

When leveraging our official Helm charts, you can set the configuration through additional environment variables on the agent:

```
--set agent.env[0].name=STEADYBIT_AGENT_ACTIONS_EXTENSIONS_0_URL \
--set agent.env[0].value="http://steadybit-extension-datadog.steadybit-extension.svc.cluster.local:8090" \
--set agent.env[1].name=STEADYBIT_AGENT_DISCOVERIES_EXTENSIONS_0_URL \
--set agent.env[1].value="http://steadybit-extension-datadog.steadybit-extension.svc.cluster.local:8090"
--set agent.env[1].name=STEADYBIT_AGENT_EVENTS_EXTENSIONS_0_URL \
--set agent.env[1].value="http://steadybit-extension-datadog.steadybit-extension.svc.cluster.local:8090"
```
