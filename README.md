# Cassandra Operator
[![Build Status](https://travis-ci.org/camilocot/cassandra-operator.svg?branch=master)](https://travis-ci.org/camilocot/cassandra-operator)
### Project Status: pre-alpha
Cassandra Operator - This is a Kubernetes Operator for Cassandra

The goal of this Operator is to support various life-cycle actions
for a Cassandra instance, such as:

- Decommission a C* instance
- Bootstrap a C* instance
- Configuring authentication

This Cassandra operator is implemented  using the [operator-sdk][operator_sdk]. The SDK CLI `operator-sdk` generates the project layout and controls the development life cycle. In addition, this implementation replaces the use of [client-go][client_go] with the SDK APIs to watch, query, and mutate Kubernetes resources.


## Quick Start

The quick start guide walks through the process of building the Cassandra operator image using the SDK CLI, setting up the RBAC, deploying operators, and creating a Cassandra cluster.

### Prerequisites

- [dep][dep_tool] version v0.4.1+.
- [go][go_tool] version v1.10+.
- [docker][docker_tool] version 17.03+.
- [kubectl][kubectl_tool] version v1.9.0+.
- Access to a kubernetes v.1.9.0+ cluster. The cassandra-operator uses `apps/v1` statefulset, the Kubernetes cluster version should be greater than 1.9.

**Note**: This guide uses quay.io for the public registry.

**Prerequisites**
Dynamic volume provisioning and storage class configured in Kubernetes. The StatefulSet controller managed by the operator creates PersistentVolumeClaims that are bound to PersistentVolumes, that the cluster should dynamically provision. A manifest for local storage class is included as example in `deploy/storage/`, `local` volumes is an alpha feature that requires the PersistentLocalVolumes feature gate to be enabled if Kubernetes cluster version is lower than v.1.10.

### Install the Operator SDK CLI

First, checkout and install the operator-sdk CLI:

```sh
$ cd $GOPATH/src/github.com/operator-framework/operator-sdk
$ git checkout tags/v0.0.5
$ dep ensure
$ go install github.com/operator-framework/operator-sdk/commands/operator-sdk
```

### Initial Setup

Checkout this Vault Operator repository:

```sh
$ mkdir $GOPATH/src/github.com/camilocot
$ cd $GOPATH/src/github.com/camilocot
$ git clone https://github.com/camilocot/cassandra-operator.git
$ cd cassandra-operator
```

Vendor the dependencies:

```sh
$ dep ensure
```

### Build and run the operator

Build the Cassandra operator image and push it to a public registry such as quay.io:

```sh
$ export IMAGE=quay.io/example/casandra-operator:v0.0.1
$ operator-sdk build $IMAGE
$ docker push $IMAGE
```

Setup RBAC for the Vault operator and its related resources:

```sh
$ kubectl create -f deploy/rbac.yaml
```

Deploy the Cassandra operator:

```sh
$ kubectl create -f deploy/operator.yaml
```
### Deploying a Cassandra cluster

Create a Cassandra cluster:

```sh
$ kubectl create -f deploy/cr.yaml
```

Verify that the Cassandra cluster is up:

```sh
$ kubectl get pods -l app=cassandra
```

### Other operators used as reference
- [Zalando Postgres][zalando-postgres-operator]
- [Vault][vault-operator]
- [Etcd][etcd-operator]
- [Prometheus][prometheus-operator]
- [Redis][redis-operator]
- [ArangoDB][arangodb-operator]
- [TensorFlow][tensorflow-operator]
- [Prometheus Jmx Exporter][prometheus-jmx-exporter-operator]

[client_go]:https://github.com/kubernetes/client-go
[vault_operator]:https://github.com/coreos/vault-operator
[operator_sdk]:https://github.com/operator-framework/operator-sdk
[dep_tool]:https://golang.github.io/dep/docs/installation.html
[go_tool]:https://golang.org/dl/
[docker_tool]:https://docs.docker.com/install/
[kubectl_tool]:https://kubernetes.io/docs/tasks/tools/install-kubectl/
[zalando-postgres-operator]:https://github.com/zalando-incubator/postgres-operator/
[vault-operator]:https://github.com/operator-framework/operator-sdk-samples/tree/master/vault-operator
[etcd-operator]:https://github.com/coreos/etcd-operator
[prometheus-operator]:https://github.com/coreos/prometheus-operator
[redis-operator]:https://github.com/spotahome/redis-operator
[arangodb-operator]:https://github.com/arangodb/kube-arangodb
[tensorflow-operator]:https://github.com/kubeflow/tf-operator
[prometheus-jmx-exporter-operator]:https://github.com/banzaicloud/prometheus-jmx-exporter-operator
