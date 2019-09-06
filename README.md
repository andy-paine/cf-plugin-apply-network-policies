A Cloud Foundry CLI plugin for managing network policies
========

Installation
------------

```
$ go get github.com/andy-paine/cf-plugin-apply-network-policies
$ cf install-plugin $GOPATH/bin/cf-plugin-apply-network-policies
```

Usage
-----

```
$ cf apply-network-manifest $YAML_FILE
```

Examples for YAML file config can be found in `examples/`. As the plugin just looks for the `network-policies` key in a YAML file, the network-policy config can be included in a normal `manifest.yml` if desired.
