
# Machineconfig Validation Mutating Webhook

## DON'T USE THIS FOR ANYTHING IMPORTANT, IT'S A SCIENCE EXPERIMENT

This is just a proof-of-concept mutating webhook to validate and annotate incoming MachineConfigs. It tries to validate the ignition section of incoming MachineConfigs and errors/rejects the request if it is unable to.

It also lets you bypass admission ignition validation by specifying a magic annotation `machineconfiguration.openshift.io/skip-validation` on the MachineConfig (just so there's an immediate way out if it's broken)

## How to use it

1. Have an openshift cluster (the manifests use openshift annotations to inject certs and create TLS secrets, it won't work on non-openshift without some adjustment)
2. Apply deployment.yaml (it uses the image from my quay.io/jkyros/ by default) and it will dump the admission controller stuff in the `openshift-machine-config-operator` namespace

```console
oc apply -f deployment.yaml
```

3. Once the pod is running (you should have a `machine-config-admission-5d6d465744-kmpmv` or somesuch in the MCO namespace)

```console
[jkyros@jkyros-t590 machineconfig-validation-webhook]$ oc get pods -n openshift-machine-config-operator
NAME                                         READY   STATUS    RESTARTS   AGE
machine-config-admission-5d6d465744-9rzsp    1/1     Running   0          7m4s
machine-config-controller-84c5864697-qsvf4   2/2     Running   0          128m
machine-config-daemon-28zsj                  2/2     Running   4          122m
machine-config-daemon-7hpgh                  2/2     Running   4          123m
```

4. Apply a MachineConfig with bad ignition, e.g. `oc apply -f examples/badmachineconfig.yaml` and observe that you receive the validation error instead of the MCO quietly degrading :smile:

```console
[jkyros@jkyros-t590 machineconfig-validation-webhook]$ oc apply  -f examples/badmachineconfig.yaml 
Error from server: error when creating "examples/badmachineconfig.yaml": admission webhook "machine-config-admission.openshift-machine-config-operator.svc" denied the request: MachineConfig '99-test-file' contains invalid ignition: parsing Ignition config failed: invalid version. Supported spec versions: 2.2, 3.0, 3.1, 3.2, 3.3, 3.4
```
