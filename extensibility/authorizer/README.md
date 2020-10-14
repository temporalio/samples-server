### Authorizer

This sample shows how to inject a low-level authorizer component that can control access to all API calls. It includes an implementation of the authorizer `myAuthorizer` which allows all requests to the "temporal-system" namespace and denies `UpdateNameSpace` calls for all other namespaces.

The sample implementation of an authorizer `myAuthorizer` allows all requests to the "temporal-system" namespace and denies `UpdateNameSpace` calls for all other namespaces.

### Steps to run this sample
1. Start up the dependencies by running the `make start-dependencies` command from within the main Temporal repository as described in the [contribution guide](https://github.com/temporalio/temporal/blob/master/CONTRIBUTING.md#runing-server-locally).

2. Create the database schema by running `make install-schema`.

3. Start Temporal by running `go run authorizer/server/main.go`.

4. Use `tctl` to interact with Temporal

- Run `tctl n l` to list available namespaces. You should only see "temporal-system" initially.
- Run `tctl --ns test n register to create a namespace "test"
- Run `tctl n l` To see "test" listed
- Run `tctl --ns test n update` to try to update the "test" namespace. You should see a `PermissionDenied` error because `myAuthorizer` denies `UpdateNameSpace` calls.
