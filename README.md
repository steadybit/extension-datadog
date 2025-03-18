<img src="./logo.png" height="130" align="right" alt="Datadog logo depicting a dog with the text 'Datadog'">

# Steadybit extension-datadog

A [Steadybit](https://www.steadybit.com/) check implementation for data exposed through Datadog.

Learn about the capabilities of this extension in our [Reliability Hub](https://hub.steadybit.com/extension/com.steadybit.extension_datadog).

## Configuration

| Environment Variable                                        | Helm value                              | Meaning                                                                                                                | Required | Default |
|-------------------------------------------------------------|-----------------------------------------|------------------------------------------------------------------------------------------------------------------------|----------|---------|
| `STEADYBIT_EXTENSION_API_KEY`                               | `datadog.apiKey`                        | [Datadog API Key](https://docs.datadoghq.com/account_management/api-app-keys/)                                         | yes      |         |
| `STEADYBIT_EXTENSION_APPLICATION_KEY`                       | `datadog.applicationKey`                | [Datadog Application Key](https://docs.datadoghq.com/account_management/api-app-keys/)                                 | yes      |         |
| `STEADYBIT_EXTENSION_SITE_PARAMETER`                        | `datadog.siteParameter`                 | [Datadog Site Parameter](https://docs.datadoghq.com/getting_started/site/#access-the-datadog-site)                     | yes      |         |
| `STEADYBIT_EXTENSION_SITE_URL`                              | `datadog.siteUrl`                       | [Datadog Site Url](https://docs.datadoghq.com/getting_started/site/#access-the-datadog-site)                           | yes      |         |
| `HTTPS_PROXY`                                               | via extraEnv variables                  | Configure the proxy to be used for Datadog communication.                                                              | no       |         |
| `STEADYBIT_EXTENSION_DISCOVERY_ATTRIBUTES_EXCLUDES_MONITOR` | `discovery.attributes.excludes.monitor` | List of Target Attributes which will be excluded during discovery. Checked by key equality and supporting trailing "*" | false    |         |

The extension supports all environment variables provided by [steadybit/extension-kit](https://github.com/steadybit/extension-kit#environment-variables).

When installed as linux package this configuration is in`/etc/steadybit/extension-datadog`.

## Datadog Permissions

The extension requires the following application key scopes:
- `monitors_downtime`
- `monitors_read`

## Installation

### Kubernetes

Detailed information about agent and extension installation in kubernetes can also be found in
our [documentation](https://docs.steadybit.com/install-and-configure/install-agent/install-on-kubernetes).

#### Recommended (via agent helm chart)

All extensions provide a helm chart that is also integrated in the
[helm-chart](https://github.com/steadybit/helm-charts/tree/main/charts/steadybit-agent) of the agent.

You must provide additional values to activate this extension.

```
--set extension-datadog.enabled=true \
--set extension-datadog.datadog.apiKey="{{API_KEY}}" \
--set extension-datadog.datadog.applicationKey="{{APPLICATION_KEY}}" \
--set extension-datadog.datadog.siteParameter="{{SITE_PARAMETER}}" \
--set extension-datadog.datadog.siteUrl="{{SITE_URL}}" \
```

Additional configuration options can be found in
the [helm-chart](https://github.com/steadybit/extension-datadog/blob/main/charts/steadybit-extension-datadog/values.yaml) of the
extension.

#### Alternative (via own helm chart)

If you need more control, you can install the extension via its
dedicated [helm-chart](https://github.com/steadybit/extension-datadog/blob/main/charts/steadybit-extension-datadog).

```bash
helm repo add steadybit-extension-datadog https://steadybit.github.io/extension-datadog
helm repo update
helm upgrade steadybit-extension-datadog \
  --install \
  --wait \
  --timeout 5m0s \
  --create-namespace \
  --namespace steadybit-agent \
  --set datadog.apiKey="{{API_KEY}}" \
  --set datadog.applicationKey="{{APPLICATION_KEY}}" \
  --set datadog.siteParameter="{{SITE_PARAMETER}}" \
  --set datadog.siteUrl="{{SITE_URL}}" \
  steadybit-extension-datadog/steadybit-extension-datadog
```

### Linux Package

Please use
our [agent-linux.sh script](https://docs.steadybit.com/install-and-configure/install-agent/install-on-linux-hosts)
to install the extension on your Linux machine. The script will download the latest version of the extension and install
it using the package manager.

After installing, configure the extension by editing `/etc/steadybit/extension-datadog` and then restart the service.

## Extension registration

Make sure that the extension is registered with the agent. In most cases this is done automatically. Please refer to
the [documentation](https://docs.steadybit.com/install-and-configure/install-agent/extension-registration) for more
information about extension registration and how to verify.


## Proxy
To communicate to Datadog via a proxy, we need the environment variable `https_proxy` to be set.
This can be set via helm using the extraEnv variable

```bash
--set "extraEnv[0].name=HTTPS_PROXY" \
--set "extraEnv[0].value=https:\\user:pwd@CompanyProxy.com:8888"
```

## Version and Revision

The version and revision of the extension:
- are printed during the startup of the extension
- are added as a Docker label to the image
- are available via the `version.txt`/`revision.txt` files in the root of the image
