# cdd-go Architecture

<!-- BADGES_START -->
[![License](https://img.shields.io/badge/license-Apache--2.0%20OR%20MIT-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Test Coverage](https://img.shields.io/badge/Test_Coverage-91.2%25-green)](https://github.com/samuel/cdd-go)
[![Doc Coverage](https://img.shields.io/badge/Doc_Coverage-100.0%25-green)](https://github.com/samuel/cdd-go)
<!-- BADGES_END -->

The **cdd-go** tool acts as a dedicated compiler and transpiler. Its fundamental architecture follows standard compiler design principles, divided into three distinct phases: **Frontend (Parsing)**, **Intermediate Representation (IR)**, and **Backend (Emitting)**.

This decoupled design ensures that any format capable of being parsed into the IR can subsequently be emitted into any supported output format, whether that is a server-side route, a client-side SDK, a database ORM, or an OpenAPI specification.

## 🏗 High-Level Overview

```mermaid
graph TD
    %% Styling Definitions
    classDef frontend fill:#57caff,stroke:#4285f4,stroke-width:2px,color:#20344b,font-family:Roboto Mono
    classDef core fill:#ffd427,stroke:#f9ab00,stroke-width:3px,color:#20344b,font-family:Google Sans,font-weight:bold
    classDef backend fill:#5cdb6d,stroke:#34a853,stroke-width:2px,color:#20344b,font-family:Roboto Mono
    classDef endpoint fill:#ffffff,stroke:#20344b,stroke-width:1px,color:#20344b,font-family:Google Sans

    subgraph Frontend [Parsers]
        A[OpenAPI .yaml/.json]:::endpoint --> P1(OpenAPI Parser):::frontend
        B[Go Models / Source]:::endpoint --> P2(Go Parser):::frontend
        C[Server Routes / Frameworks]:::endpoint --> P3(Framework Parser):::frontend
        D[Client SDKs / ORMs]:::endpoint --> P4(Ext Parser):::frontend
    end

    subgraph Core [Intermediate Representation]
        IR((CDD IR)):::core
    end

    subgraph Backend [Emitters]
        E1(OpenAPI Emitter):::backend --> X[OpenAPI .yaml/.json]:::endpoint
        E2(Go Emitter):::backend --> Y[Go Models / Structs]:::endpoint
        E3(Server Emitter):::backend --> Z[Server Routes / Controllers]:::endpoint
        E4(Client Emitter):::backend --> W[Client SDKs / API Calls]:::endpoint
        E5(Data Emitter):::backend --> V[ORM Models / CLI Parsers]:::endpoint
    end

    P1 --> IR
    P2 --> IR
    P3 --> IR
    P4 --> IR

    IR --> E1
    IR --> E2
    IR --> E3
    IR --> E4
    IR --> E5
```

## 🧩 Core Components

### 1. The Frontend (Parsers)

The Frontend's responsibility is to read an input source and translate it into the universal CDD Intermediate Representation (IR).

* **Static Analysis (AST-Driven)**: For `Go` source code, the tool **does not** use dynamic reflection or execute the code. Instead, it reads the source files, generates an Abstract Syntax Tree (using `github.com/dave/dst` to preserve formatting), and navigates the tree to extract classes, structs, functions, type signatures, API client definitions, server routes, and docstrings.
* **OpenAPI Parsing**: For OpenAPI inputs, the parser normalizes the structure using `encoding/json` and the standard `openapi` library into the IR.

### 2. Intermediate Representation (IR)

The Intermediate Representation is the crucial "glue" of the architecture. It is a normalized, language-agnostic data structure that represents concepts like:
* **Models**: Entities containing typed properties, required fields, defaults, and descriptions.
* **Endpoints / Operations**: HTTP verbs, paths, path/query/body parameters, and responses. In the IR, an operation is an abstract concept that can represent *either* a Server Route receiving a request *or* an API Client dispatching a request.
* **Metadata**: Tooling hints, docstrings, and validations.

By standardizing on a single IR, the system guarantees that parsing logic and emitting logic remain completely decoupled.

### 3. The Backend (Emitters)

The Backend's responsibility is to take the universal IR and generate valid target output.

* **Code Generation**: Emitters iterate over the IR and generate idiomatic `Go` source code. 
  * A **Server Emitter** creates routing controllers (`net/http`) and request-validation logic.
* **Specification Generation**: Emitters translating back to OpenAPI serialize the IR into standard OpenAPI 3.2.0 JSON, rigorously formatting descriptions, type constraints, and endpoint schemas based on what was parsed from the source code.

## 🔄 Extensibility

Because of the IR-centric design, adding support for a new `Go` framework requires minimal effort:
1. **To support parsing a new framework**: Write a parser that converts the framework's AST/DSL into the CDD IR. Once written, the framework can automatically be exported to OpenAPI, Client SDKs, CLI parsers, or any other existing output target.
2. **To support emitting a new framework**: Write an emitter that converts the CDD IR into the framework's DSL/AST. Once written, the framework can automatically be generated from OpenAPI or any other supported input.

## 🛡 Design Principles

1. **A Single Source of Truth**: Developers should be able to maintain their definitions in whichever format is most ergonomic for their team (OpenAPI files, Native Code) and generate the rest.
2. **Zero-Execution Parsing**: Ensure security and resilience by strictly statically analyzing inputs. The compiler must never need to run the target code to understand its structure.
3. **Lossless Conversion**: Maximize the retention of metadata (e.g., type annotations, docstrings, default values, validators) during the transition `Source -> IR -> Target`. This is largely achieved by employing the `dst` library to maintain comments and whitespace integrity throughout transformations.
4. **Symmetric Operations**: An Endpoint in the IR holds all the information necessary to generate both the Server-side controller that fulfills it, and the Client-side SDK method that calls it.