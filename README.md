# calculator

Calculator is a demo project to show how I might build a service in Go.

The idea is for a client to send calculations to a gRPC server. The server
should distribute the calculations as jobs. We'll just pretend they're
heavy-duty calculations for now.

```mermaid
C4Context
title User submission

Person(actor, "user")
System(calculator, "does mathematical calculations")
Rel(actor, calculator, "FibOf", "gRPC")
```

Because some calculations might take a while, a user will get a job ID back and
be able to retrieve it later. This is implemented as long-running operations
where a user can get the results of the submitted calculation.

```mermaid
C4Context
title Retrieve solution

Person(actor, "user")
System(calculator, "does mathematical calculations")
Rel(actor, calculator, "GetOperation", "gRPC")
```

The calculator has two major components: the API (in gRPC) and the calculators,
which are the workers. The API can send the calculation requests to any number
of workers, but only one of them will pick up a job at a time.

```mermaid
C4Container
title Distriuton of calculations

System_Boundary(b, "calculator") {
    Container(api, API, gRPC, "manages calculation requests")
    ContainerQueue(mq, "Message Queue")
    ContainerDb(store, "Data Store", "etcd")

    Container(calc, calculator, "Go")


    UpdateLayoutConfig($c4ShapeInRow="2", $c4BoundaryInRow="2")

    Rel(api, mq, "Submits jobs")
    Rel(api, store, "Creates operations")
    Rel(calc, mq, "Takes jobs")
    Rel(calc, store, "Updates operations with results")
}
```

## Running

To run the daemon locally:

```shell
CALCULATORD_RABBIT_USER=guest CALCULATORD_RABBIT_PASS=guest ./calculatord
```

To run the worker locally:

```shell
CALCULATORW_RABBIT_USER=guest CALCULATORW_RABBIT_PASS=guest ./calculatorw
```

To use grpcurl to submit a calculation:

```shell
grpcurl -d '{"first": 0, "second": 1, "nth_position": 10}' \
    -plaintext \
    -proto proto/calculator.proto \
    -import-path $(pwd)/proto \
    -import-path $(pwd)/proto/third_party/googleapis \
    localhost:8080 calculator.Calculations/FibonacciOf
```

It returns an operation with a name, such as:

```json
{
  "name": "4ea7e923-8ec0-42ff-b974-97b9869f8ab4",
  "metadata": {
    "@type": "type.googleapis.com/calculator.CalculationMetadata",
    "created": "2024-07-15T20:33:08.268024Z"
  }
}
```

To use grpcurl to check the status of a calculation:

```shell
grpcurl -d '{"name": "4ea7e923-8ec0-42ff-b974-97b9869f8ab4"}' \
    -plaintext \
    -proto proto/calculator.proto \
    -import-path $(pwd)/proto \
    -import-path $(pwd)/proto/third_party/googleapis \
    localhost:8080 calculator.Calculations/GetOperation
```

## Building

To build the daemon:

```shell
make calculatord
```

## Deploying to Docker

## Generating Protos

Run `make proto` to regenerate the protos.
