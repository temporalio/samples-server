PromQL to Datadog
=================

The goal of this example script is to demonstrate the minimum work necessary to read recently generated metrics from a Temporal Cloud account using the Prometheus API and import them into DataDog while handling some common edge and error cases. 
Python was chosen for relative simplity and clarity, with some choices made to get from downloading this script to seeing metrics in DataDog as quickly as possible. Examples in both [Typescript](promql-to-dd-ts) and [Go](promql-to-dd-go) have also been provided.

Destination metrics could be modified to match a DataDog environment's naming conventions and metrics types by modifying the DataDog API calls as needed.

**These examples are provided as-is, without support. They are intended as reference material only.**

Requirements
------------

* Python 3.9+
* install python package dependencies
* have a datadog account and access to an api key
* have created a temporal observability endpoint with a CA
* have client certs signed by the CA uploaded to temporal

To install the required dependencies run:

```bash
python3 -m pip install requests
python3 -m pip install tenacity
python3 -m pip install datadog_api_client
```

Usage
-----

Running the script (either as an executable or with `python3 promql-to-dd.py`) will print out information about dependencies, and the `-h` cli argument can be used to see usage and command line options.

Either ENV vars or cli args can be specified for any of the required inputs, and all inputs are required to be specified in one way or another.

Here is the help output:

```
usage: promql-to-dd.py [-h] --temporal-account TEMPORAL_ACCOUNT --metrics-client-cert METRICS_CLIENT_CERT --metrics-client-key METRICS_CLIENT_KEY --dd-site DD_SITE --dd-api-key DD_API_KEY

Export Prometheus API metrics and import into DataDog

options:
  -h, --help            show this help message and exit
  --temporal-account TEMPORAL_ACCOUNT
  --metrics-client-cert METRICS_CLIENT_CERT
  --metrics-client-key METRICS_CLIENT_KEY
  --dd-site DD_SITE
  --dd-api-key DD_API_KEY
```

The env var names are the same as the `ALL_CAPS` names of the cli arg values, and cli args take precedence.

Example
-------

Here's an example, of setting up an environment and running the script:

```bash
# set up python virtual environment
python3 -m venv venv
. venv/bin/activate

# install deps
python3 -m pip install requests
python3 -m pip install tenacity
python3 -m pip install datadog_api_client

# set up environment
export TEMPORAL_ACCOUNT=acctcode
export METRICS_CLIENT_CERT=/path/to/cert
export METRICS_CLIENT_KEY=/path/to/key
export DD_SITE=datadog.api.endpoint.domain
export DD_API_KEY=123apikey321

# run the script
python3 promql-to-dd.py
```

Temporal Cloud Observability DataDog Import Sequence Diagram
------------------------------------------------------------

This sequence diagram shows the logical progression of metrics information from Temporal Cloud server side metrics collection through to user import into DataDog.

```mermaid
sequenceDiagram
    participant Temporal Server
    participant Gathering
    participant Processing
    participant PromQL
    participant Importer
    participant DataDog
    loop 15s interval
        Gathering ->> Temporal Server: scrape
        Gathering ->> Processing: push
    end 
    Note right of Processing: Gathering is<br>gobally distributed<br>and pushed at<br>various times
    loop 30s interval
        Processing ->> Processing: aggregate
            Note right of Processing: Aggregation is over<br>a look back
        Processing ->> PromQL: publish
    end
    Note right of PromQL: Aggregation and<br>publish loop can run<br>in parallel and out<br>of order
    loop 30s interval
        Importer ->> PromQL: query 10m range
        PromQL -->> Importer: published results
        Importer ->> DataDog: send metrics series
        DataDog -->> Importer: report success
    end
    Note right of DataDog: 10m range covers<br>out of order<br>results and<br>transient errors
    ```