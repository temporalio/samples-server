# Temporal Server docker-compose files

These docker-compose files enable you to run a local instance of the Temporal Server.
There are a variety of docker-compose files, each utilizing a different set of dependencies.

## Prerequisites

To use these files, you must first have the following installed:

- [Docker](https://docs.docker.com/engine/installation/)
- [docker-compose](https://docs.docker.com/compose/install/)

## How to use

The following steps will run a local instance of the Temporal Server using the default configuration file (`docker-compose.yml`):

1. Clone this repository.
2. Change directory into the root of the project.
3. Run the `docker-compose up` command.

```bash
git clone https://github.com/temporalio/samples-server.git
cd samples-server/compose
docker-compose up
```

> ⚠️ If you are on an M1 Mac, note that Temporal v1.12 to v1.14 had fatal issues with ARM builds. v1.14.2 onwards should be fine for M1 Macs.

After the Server has started, you can open the Temporal Web UI in your browser: [http://localhost:8080](http://localhost:8080).

You can also interact with the Server using a preconfigured CLI (tctl).
First create an alias for `tctl`:

```bash
alias tctl="docker exec temporal-admin-tools tctl"
```

The following is an example of how to register a new namespace `test-namespace` with 1 day of retention:

```bash
tctl --ns test-namespace namespace register -rd 1
```

You can find our `tctl` docs on [docs.temporal.io](https://docs.temporal.io/docs/system-tools/tctl/).

Get started building Workflows with a [Go sample](https://github.com/temporalio/samples-go), [Java sample](https://github.com/temporalio/samples-java), or write your own using one of the [SDKs](https://docs.temporal.io/docs/sdks-introduction).

### Other configuration files

The default configuration file (`docker-compose.yml`) uses a PostgreSQL database, an Elasticsearch instance, and exposes the Temporal gRPC Frontend on port 7233.
The other configuration files in the repo spin up instances of the Temporal Server using different databases and dependencies.
For example you can run the Temporal Server with MySQL and Elastic Search with this command:

```bash
docker-compose -f docker-compose-mysql-es.yml up
```

Here is a list of available files and the dependencies they use.

| File                                   | Description                                                   |
|----------------------------------------|---------------------------------------------------------------|
| docker-compose.yml                     | PostgreSQL and Elasticsearch (default)                        |
| docker-compose-tls.yml                 | PostgreSQL and Elasticsearch with TLS                         |
| docker-compose-postgres.yml            | PostgreSQL                                                    |
| docker-compose-cass-es.yml             | Cassandra and Elasticsearch                                   |
| docker-compose-mysql.yml               | MySQL                                                         |
| docker-compose-mysql-es.yml            | MySQL and Elasticsearch                                       |
| docker-compose-postgres-opensearch.yml | PostgreSQL and OpenSearch                                     |
| docker-compose-multirole.yml           | PostgreSQL and Elasticsearch with mult-role Server containers |

### Using multi-role configuration

First install the loki plugin (this is one time operation)
```bash
docker plugin install grafana/loki-docker-driver:latest --alias loki --grant-all-permissions
```

Start multi-role Server configuration:
```
docker compose -f docker-compose-multirole.yaml up
```

Some exposed endpoints:
- http://localhost:8080 - Temporal Web UI
- http://localhost:8085 - Grafana dashboards
- http://localhost:9090 - Prometheus UI
- http://localhost:9090/targets - Prometheus targets
- http://localhost:8000/metrics - Server metrics

### Using the web interface

`docker-compose.yml` includes the Temporal Web UI.

If you run command:

```bash
docker-compose up
```

You access the Temporal Web UI at http://localhost:8080.

### Enabling metrics (with Grafana and Prometheus)

We maintain two example docker-compose setups with server metrics enabled, and Prometheus and Grafana with [our Server and SDK dashboards](https://github.com/temporalio/dashboards):

- https://github.com/tsurdilo/my-temporal-dockercompose
- https://github.com/temporalio/background-checks

### Use a custom image configuration

If you want, you can even use a custom Docker image of the Temporal Server.

Clone the main Temporal Server repo: [https://github.com/temporalio/temporal](https://github.com/temporalio/temporal):

```bash
git clone https://github.com/temporalio/temporal.git
```

In the following command, replace **<YOUR_TAG>** and **<YOUR_COMMIT>** to build the custom Docker image:

```bash
git checkout <YOUR_COMMIT>
docker build . -t temporalio/server:<YOUR_TAG> --build-arg TARGET=server
```

Next, in the `docker-compose.yml` file, replace the `TEMPORAL_VERSION` value in the `.env` file with **<YOUR_TAG>**.

Then run the `docker-compose up` command:

```bash
docker-compose up
```

## Using Temporal docker images in production

These docker-compose setups use the `temporalio/server` image with a separate initialization step to set up database schemas. The `temporal-admin-tools` service runs once to create and initialize the database schema, then the `temporal` service starts using `temporalio/server`.

In a typical production setting, dependencies such as `cassandra` or `elasticsearch` are managed/started independently of the Temporal server, and schemas are set up as part of your deployment process rather than at startup.

To use the `temporalio/server` container in a production setting, use the following command:

```plain
docker run -e CASSANDRA_SEEDS=10.x.x.x                  -- csv of Cassandra server ipaddrs
    -e KEYSPACE=<keyspace>                              -- Cassandra keyspace
    -e VISIBILITY_KEYSPACE=<visibility_keyspace>        -- Cassandra visibility keyspace
    -e SKIP_SCHEMA_SETUP=true                           -- do not setup Cassandra schema during startup
    -e NUM_HISTORY_SHARDS=1024  \                       -- Number of history shards
    -e SERVICES=history,matching \                      -- Spin-up only the provided services
    -e LOG_LEVEL=debug,info \                           -- Logging level
    -e DYNAMIC_CONFIG_FILE_PATH=config/foo.yaml         -- Dynamic config file to be watched
    temporalio/server:<tag>
```

## Server Configuration Templates

The Temporal Server uses a base configuration template that defines the structure for persistence, visibility, and other settings.

### Configuration template location by version

**Pre-v1.30 (external template):**
- Configuration template: [`docker/config_template.yaml`](https://github.com/temporalio/temporal/blob/main/docker/config_template.yaml)
- The template is stored as a separate file in the Docker image
- Environment variables are substituted into this template at runtime

**v1.30 and later (embedded template):**
- Configuration template: [`common/config/config_template_embedded.yaml`](https://github.com/temporalio/temporal/blob/main/common/config/config_template_embedded.yaml)
- The template is embedded directly in the server binary
- More efficient and reduces dependencies on external files
- Environment variable substitution works the same way

### Impact on docker-compose configurations

The docker-compose files in this repository work with both pre-v1.30 and v1.30+ server versions. The main differences are:

1. **Admin tools**: v1.30+ includes improved tooling like `temporal-elasticsearch-tool` and enhanced `temporal-cassandra-tool` commands
2. **Configuration**: v1.30+ uses the embedded template, but accepts the same environment variables
3. **Schema management**: Setup scripts detect and use new tools when available, with fallback to legacy methods

For customizing server configuration beyond environment variables, refer to the appropriate template file for your server version.

