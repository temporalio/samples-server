### Authorizer

This sample shows how to inject a low-level authorizer component that can control access to all API calls. It includes an implementation of the authorizer `myAuthorizer` which allows all requests to the "temporal-system" namespace and denies `UpdateNameSpace` calls for all other namespaces. 
The sample implementation of the authorizer interface `authorization.Authorizer` allows all requests to the "temporal-system" namespace and denies `UpdateNamespace` calls for all other namespaces.

### Steps to run this sample
1. Start up the dependencies by running the `make start-dependencies` command from within the main Temporal repository as described in the [contribution guide](https://github.com/temporalio/temporal/blob/master/CONTRIBUTING.md#run-temporal-server-locally).

2. Create the database schema by running `make install-schema-cass-es`.

3. Start Temporal by running `go run authorizer/server/main.go`.

4. Use `temporal` cli to interact with Temporal

- Run `temporal operator namespace list` to list available namespaces. You should only see "temporal-system" initially.
- Run `temporal operator namespace create -n test` to create a namespace "test"
- Run `temporal operator namespace list` to see "test" listed
- Run `temporal operator namespace update -n test` to try to update the "test" namespace. You should see a `PermissionDenied` error because `myAuthorizer` denies `UpdateNamespace` calls.
