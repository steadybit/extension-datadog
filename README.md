<img src="./logo.png" height="130" align="right" alt="Datadog logo depicting a dog with the text 'Datadog'">

# Steadybit extension-datadog

A [Steadybit](https://www.steadybit.com/) check implementation for data exposed through Datadog.

Learn about the capabilities of this extension in our [Reliability Hub](https://hub.steadybit.com/extension/com.steadybit.extension_datadog).

## Configuration

| Environment Variable                  | Helm value               | Meaning                                                                                            | Required | Default |
|---------------------------------------|--------------------------|----------------------------------------------------------------------------------------------------|----------|---------|
| `STEADYBIT_EXTENSION_API_KEY`         | `datadog.apiKey`         | [Datadog API Key](https://docs.datadoghq.com/account_management/api-app-keys/)                     | yes      |         |
| `STEADYBIT_EXTENSION_APPLICATION_KEY` | `datadog.applicationKey` | [Datadog Application Key](https://docs.datadoghq.com/account_management/api-app-keys/)             | yes      |         |
| `STEADYBIT_EXTENSION_SITE_PARAMETER`  | `datadog.siteParameter`  | [Datadog Site Parameter](https://docs.datadoghq.com/getting_started/site/#access-the-datadog-site) | yes      |         |
| `STEADYBIT_EXTENSION_SITE_URL`        | `datadog.siteUrl`        | [Datadog Site Url](https://docs.datadoghq.com/getting_started/site/#access-the-datadog-site)       | yes      |         |

The extension supports all environment variables provided by [steadybit/extension-kit](https://github.com/steadybit/extension-kit#environment-variables).

## Datadog Permissions

The extension requires the following application key scopes:
- `monitors_downtime`
- `monitors_read`

## Installation

We recommend that you install the extension with
our [official Helm chart](https://github.com/steadybit/extension-datadog/tree/main/charts/steadybit-extension-datadog).

### Helm

```bash
helm repo add steadybit-extension-datadog https://steadybit.github.io/extension-datadog
helm repo update
```

```bash
helm upgrade steadybit-extension-datadog \
  --install \
  --wait \
  --timeout 5m0s \
  --create-namespace \
  --namespace steadybit-extension \
  --set datadog.apiKey="{{API_KEY}}" \
  --set datadog.applicationKey="{{APPLICATION_KEY}}" \
  --set datadog.siteParameter="{{SITE_PARAMETER}}" \
  --set datadog.siteUrl="{{SITE_URL}}" \
  steadybit-extension-datadog/steadybit-extension-datadog`
```

### Docker

You may alternatively start the Docker container manually.

```bash
docker run \
  --env STEADYBIT_LOG_LEVEL=info \
  --env STEADYBIT_EXTENSION_API_KEY="{{API_KEY}}" \
  --env STEADYBIT_EXTENSION_APPLICATION_KEY="{{APPLICATION_KEY}}" \
  --env STEADYBIT_EXTENSION_SITE_PARAMETER="{{SITE_PARAMETER}}" \
  --env STEADYBIT_EXTENSION_SITE_URL="{{SITE_URL}}" \
  --expose 8090 \
  ghcr.io/steadybit/extension-datadog:latest
```

## Register the extension

Make sure to register the extension at the steadybit platform. Please refer to
the [documentation](https://docs.steadybit.com/integrate-with-steadybit/extensions/extension-installation) for more information.

### Linux Package

Please use our [outpost-linux.sh script](https://docs.steadybit.com/install-and-configure/install-outpost-agent-preview/install-on-linux-hosts) to install the extension on your Linux machine.
The script will download the latest version of the extension and install it using the package manager.

After installing configure the extension by editing `/etc/steadybit/extension-datadog` and then restart the service.
