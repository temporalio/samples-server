# Temporal Server Samples
These samples show how to run and customize Temporal server for local development and production scenarios.

Learn more about Temporal at:
* Documentation: https://docs.temporal.io
* Main repository: https://github.com/temporalio/temporal
* Go SDK: https://github.com/temporalio/sdk-go
* Java SDK: https://github.com/temporalio/sdk-java

## Prerequisites

- docker
- docker-compose
## Steps to run samples
Please follow instructions from README.md file in every sample directory.

## Samples

- **[Docker Compose](./compose/)**: docker-compose files to run a local Temporal Server with various database and dependency configurations (PostgreSQL, MySQL, Cassandra, Elasticsearch, OpenSearch).
- **[TLS](./tls/)**: how to configure Transport Layer Security (TLS) to secure network communication with and within Temporal cluster.
- **[Authorizer](./extensibility/authorizer)**: how to inject a low-level authorizer component that can control access to all API calls.

## Contributing

We'd love your help in making Temporal great. Please review our [contribution guide](https://github.com/temporalio/temporal/blob/master/CONTRIBUTING.md).


[MIT License](LICENSE)