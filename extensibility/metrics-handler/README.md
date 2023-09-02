### Metrics reporter

This sample shows how to inject a custom metrics handler.

### Steps to run this sample
1. Start up the dependencies by running the `make start-dependencies` command from within the main Temporal repository
as described in the 
[contribution guide](https://github.com/temporalio/temporal/blob/master/CONTRIBUTING.md#runing-server-locally).

2. Create the database schema by running `make install-schema`.

3. Start Temporal by running `go run metrics-reporter/server/main.go`.

4. Metrics will be output to console.
