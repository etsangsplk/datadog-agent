# Configuration Options

This page details the configuration options supported in Agent version 6 or later
compared to version 5. If you don't see a configuration option listed here, this
might mean:

 * The option is supported as it is, in this case you can find it in the [example file][datadog-yaml]
 * The option refers to a feature that's currently under development
 * The option refers to a feature that's scheduled but will come later

## Orchestration + Agent Management

Orchestration has now been deferred to OS facilities wherever possible. To this purpose
we now rely on upstart/systemd on linux environments and windows services on Windows.
Enabling the APM and Process agents bundled with the agent can now be achieved via
configuration flags defined in the main configuration file: `datadog.yaml`.

### Process Agent
To enable the process agent add the following to `datadog.yaml`:
```
...
process_config:
  enabled: true
...
```

### Trace Agent
To enable the trace agent add the following to `datadog.yaml`:
```
...
apm_config:
  enabled: true
...
```

The OS-level services will be enabled by default for all agents. The agents will process
the configuration and decide whether to stay up or gracefully shut down. You may decide
to disable the OS-level service units, but that will require your manual intervention if
you ever wish to re-enable any of the agents.


## New options in version 6

This is the list of configuration options that are either new, renamed or changed
in any way.

| Old Name | New Name | Notes |
| --- | --- | --- |
| `proxy_host`  | `proxy`  | Proxy settings are now expressed as a list of URIs like `http://user:password@proxyurl:port`, one per transport type (see the `proxy` section of [datadog.yaml][datadog-yaml] for more details). |
| `collect_instance_metadata` | `enable_metadata_collection` | This now enabled the new metadata collection mechanism |
| `collector_log_file` | `log_file` ||
| `syslog_host`  | `syslog_uri`  | The Syslog configuration is now expressed as an URI |
|| `syslog_pem`  | Syslog configuration root CA for TLS client server validation |


## Removed options

This is the list of configuration options that were removed in the new Agent
beacause either superseded by new options or related to features that works
differently from Agent version 5.

| Name | Notes |
| --- | --- |
| `proxy_port` | superseded by `proxy` |
| `proxy_user` | superseded by `proxy` |
| `proxy_password` | superseded by `proxy` |
| `proxy_forbid_method_switch` | obsolete |
| `use_mount` | deprecated in v5 |
| `use_curl_http_client` | obsolete |
| `exclude_process_args` | deprecated feature |
| `check_timings` | superseded by internal stats |
| `dogstatsd_target` | |
| `dogstreams` | |
| `custom_emitters` | |
| `forwarder_log_file` | superseded by `log_file` |
| `dogstatsd_log_file` | superseded by `log_file` |
| `jmxfetch_log_file` | superseded by `log_file` |
| `syslog_port` | superseded by `syslog_uri` |
| `check_freq` | |
| `collect_orchestrator_tags` | feature now implemented in metadata collectors |
| `utf8_decoding` | |
| `developer_mode` | |
| `use_forwarder` | |
| `autorestart` | |
| `dogstream_log` | |
| `use_curl_http_client` | |
| `gce_updated_hostname` | v6 behaves like v5 with `gce_updated_hostname` set to true. May affect reported hostname, see [doc][gce-hostname] |
| `collect_security_groups` | feature still available with the aws integration  |

## [BETA] Encrypted secrets

**This feature is in beta and its options or behaviour might break between
minor or bugfix releases of the Agent.**

This feature is only available on Linux at the moment.

Starting with the agent 6.3.0 the agent is able to collect encrypted secrets in
the configurations files from an external source. This allows users to use the
agent alongside a secret management services (like `Vault` or other).

This section will cover how to set up this feature.

### Encrypting password in check configuration

To fetch a secret from a check configuration simply use the `ENC[]` notation.
This handle can be used as a *value* of any YAML fields in your configuration
(not the key) in any section (`init_config`, `instances`, `logs`, ...).

Encrypted secrets are supported in every configuration backend: file, etcd,
consul ...

Example:

```yaml
instances:
  - server: db_prod
    user: ENC[db_prod_user]
    password: ENC[db_prod_password]
```

In the above example we have two encrypted  secrets : `db_prod_user` and
`db_prod_password`. Those are the secrets **handle** and must uniquely identify
a secret within your secrets management tool.

Between the brackets every character is allowed as long as the configuration
YAML is valid. This means you could use any format you want.

Example:

```
ENC[{"env": "prod", "check": "postgres", "id": "user_password", "az": "us-east-1a"}]
```

In this example the secret handle is the string `{"env": "prod", "check":
"postgres", "id": "user_password", "az": "us-east-1a"}`.

**Autodiscovery**:

Secrets are resolved **after** Autodiscovery template variables. This means you
can use them in a secret handle to fetch secrets specific to a container.

Example:

```
instances:
  - server: %%host%%
    user: ENC[db_prod_user_%%host%%]
    password: ENC[db_prod_password_%%host%%]
```

### Fetching secrets

To fetch the secrets from you configurations you have to provide an executable
capable of fetching passwords from your secrets management tool.

An external executable solved multiple issues:

- Guarantees that the agent will never have more access/information than what
  you transfer to it.
- Maximum freedom and flexibility: allowing users to use any secrets management
  tool (including closed sources ones) without having to rebuild the agent.
- Solving the **initial trust** is a very complex problem as it usually changes
  from users to users. Each setup requires more or less control and might rely on
  private tools. This shifts the problem to where it is easier to solve, without
  having to rebuild the agent.

This executable will be executed by the agent once per check's instance. The
agent will cache secrets internally to reduce the number of calls (useful in
a containerized environment for example).

The executable **MUST** (the agent will refuse to use it otherwise):

- Belong to the same user running the agent (usually `dd-agent`).
- Have **no** rights for `group` or `other`.
- Have at least `exec` right for the owner.
- The executable will not share any environment variables with the agent.

#### Configuration

In `datadog.yaml` you must set the following variables:

```yaml
secret_backend_command: /path/to/you/executable
```

More settings are available: see `datadog.yaml`.

#### The executable API

The executable has to respect a very simple API: it reads a JSON on the
Standard input and output a JSON containing the decrypted secrets on the
Standard output.

**Input:**

The executable will receive a JSON payload on the `Standard input` containing
the list of secret to fetch:

```json
{
  "version": "1.0",
  "secrets": ["secret1", "secret2"]
}
```

- `version`: is a string containing the format version (currently "1.0").
- `secrets`: is a list of strings, each string is a **handle** from a
  configuration corresponding to a secret to fetch.

**Output:**

The executable is expected to output on the `Standard output` a JSON containing
the fetched secrets:

```json
{
  "secret1": {
    "value": "secret_value",
    "error": null
  },
  "secret2": {
    "value": null,
    "error": "could not fetch the secret"
  }
}
```

The expected payload is a JSON object, each key is one of the **handle** requested in the input payload.
The value for each **handle** is a JSON object with 2 fields:

- `value`: a string: the actual secret value to be used in the check
  configurations (can be `null` in case of error).
- `error`: a string: the error message if needed. If `error` is different that
  `null` the configuration will be considered erroneous and dropped.

Example:

Here is a dummy script prefixing every secret by `decrypted_`:

```py
#!/usr/bin/python

import json
import sys

# Reading the input payload from STDIN
payload = sys.stdin.read()
# parsing the payload
requested_secrets = json.loads(payload)

secrets = {}
for secret_handle in requested_secrets["secrets"]:
    secrets[secret_handle] = {"value": "decrypted_"+secret_handle, "error": None}

print json.dumps(secrets)
```

This will update this configuration:

```yaml
instances:
  - server: db_prod
    user: ENC[db_prod_user]
    password: ENC[db_prod_password]
```

to this:

```yaml
instances:
  - server: db_prod
    user: decrypted_db_prod_user
    password: decrypted_db_prod_password
```

[datadog-yaml]: https://raw.githubusercontent.com/DataDog/datadog-agent/master/pkg/config/config_template.yaml
[gce-hostname]: changes.md#gce-hostname
