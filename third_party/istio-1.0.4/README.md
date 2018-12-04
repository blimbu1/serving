The istio\*.yaml files are generated by running

```
./download-istio.sh
```

The Helm options we used for generating istio.yaml are:

1. `sidecarInjectorWebhook.enabled=true` & `sidecarInjectorWebhook.enableNamespacesByDefault=true`: We allow sidecar injection on all namespaces.
2. `global.proxy.autoInject=disabled`: However, only apply sidecar injection for Pods annotated with `istio.sidecar.inject=true`, and not as a default.
3. `prometheus.enabled=false`: Disable Prometheus by default.

Our goal here is to allow sidecar injection for Pods created by Knative, and
nothing else. This template is used in integration tests and also released as
an Istio-one-line-installation so that our users don't have to go through a lot
of steps to install Istio.