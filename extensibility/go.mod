module github.com/temporalio/service-samples

go 1.14

require (
	github.com/uber-go/tally v3.3.17+incompatible
	// TODO: replace this with latest server release to pick up
	// 	extensibility support change.
	// 	CommitID: 3e11440de58d05583aa04208d0d89b5650ea82e7
	go.temporal.io/server v1.9.2
)
