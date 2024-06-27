# calculator

Calculator is a demo project to show how I might build a service in Go.

The idea is for a client to send calculations to a gRPC server. The server
should distribute the calculations as jobs. We'll just pretend they're
heavy-duty calculations for now.

```plantuml
@startuml
!include <C4/C4_Context>

Person(actor, actor)
System(calculator, calculator, "does mathematical calculations")
Rel(actor, calculator, "submits calculation request", gRPC)
@enduml
```

Because some calculations might take a while, a user will get a job ID back and
be able to retrieve it later.

```plantuml
@startuml
!include <C4/C4_Context>

Person(actor, actor)
System(calculator, calculator, "does mathematical calculations")
Rel(actor, calculator, "retrieves calculation request", gRPC)
@enduml
```

The calculator has two major components: the API (in gRPC) and the calculators,
which are the workers. The API can send the calculation requests to any number
of workers, but only one of them will pick up a job at a time.

```plantuml
@startuml
!include <C4/C4_Container>

Container(api, API, gRPC, "manages calculation requests")
Container(calc, calculator, Go)
Container(mq, Message Queue)
Rel(api, mq, "Submits jobs")
Rel(calc, mq, "Takes jobs")
@enduml
```

When a calculation is complete, the worker will send the result back to the API.

<!-- reminder: I need to scrub expired results -->

```plantuml
@startuml
!include <C4/C4_Container>

Container(calc, calculator, Go)
Container(mq, Message Queue)
Rel(calc, mq, "Submits results")

' What tech should I use? Probably just a key/value store so I don't have to
' sweat floats, ints, etc.
ContainerDb(store, results)
@enduml
```

The API will follow up by storing the result for retrieval.

```plantuml
@startuml
!include <C4/C4_Container>

Container(api, API, gRPC)
Container(mq, Message Queue)
Rel(api, mq, "Receives results")

' What tech should I use? Probably just a key/value store so I don't have to
' sweat floats, ints, etc.
ContainerDb(store, results)

Rel(api, store, "Stores results")
@enduml
```

## Running

## Building

## Deploying to Docker

## Generating Protos

Run `make proto` to regenerate the protos.
