# Steadybit Datadog Extension 

This Helm chart adds the Steadybit Datadog extension to your Kubernetes cluster via a deployment.

## Quick Start

### Add Steadybit Helm repository

```
helm repo add steadybit https://steadybit.github.io/helm-charts
helm repo update
```

### Installing the Chart

To install the chart with the name `steadybit-extension-datadog`. To learn more about supported the supported values for `datadog.siteParameter` and `datadog.siteUrl`, please see [Datadog's site documentation page](https://docs.datadoghq.com/getting_started/site/#access-the-datadog-site). You may alternatively decide to configure the `datadog.*` values through a pre-existing secret. See the documentation for [`datadog.existingSecret`](https://github.com/steadybit/helm-charts/blob/main/charts/steadybit-extension-datadog/values.yaml#L15) to learn more.

```bash
$ helm upgrade steadybit-extension-datadog \
    --install \
    --wait \
    --timeout 5m0s \
    --create-namespace \
    --namespace steadybit-extension \
    --set datadog.apiKey="{{API_KEY}}" \
    --set datadog.applicationKey="{{APPLICATION_KEY}}" \
    --set datadog.siteParameter="datadoghq.eu" \
    --set datadog.siteUrl="https://app.datadoghq.eu" \
    steadybit/steadybit-extension-datadog
```
