module github.com/giantswarm/gsctl

go 1.13

require (
	github.com/Jeffail/gabs v1.4.0
	github.com/Masterminds/semver v1.5.0
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/coreos/bbolt v1.3.3 // indirect
	github.com/coreos/etcd v3.3.17+incompatible // indirect
	github.com/fatih/color v1.9.0
	github.com/giantswarm/apiextensions v0.0.0-20191213075442-71155aa0f5b7
	github.com/giantswarm/columnize v2.0.3-0.20190718092621-cc99d98ffb29+incompatible
	github.com/giantswarm/gscliauth v0.1.1-0.20200312170820-9ee36484efa2
	github.com/giantswarm/gsclientgen v2.0.3+incompatible
	github.com/giantswarm/k8sclient v0.0.0-20191213144452-f75fead2ae06
	github.com/giantswarm/kubeconfig v0.0.0-20191209121754-c5784ae65a49
	github.com/giantswarm/microerror v0.0.0-20191011121515-e0ebc4ecf5a5
	github.com/giantswarm/micrologger v0.0.0-20190118112544-0926d9b7c541
	github.com/go-openapi/errors v0.19.3-0.20190617201723-9b273e805998 // indirect
	github.com/go-openapi/runtime v0.19.4
	github.com/go-openapi/strfmt v0.19.5
	github.com/gobuffalo/envy v1.8.1 // indirect
	github.com/gobuffalo/packr v1.30.1
	github.com/golang/protobuf v1.3.5 // indirect
	github.com/google/go-cmp v0.4.0
	github.com/googleapis/gnostic v0.4.1 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2
	github.com/howeyc/gopass v0.0.0-20170109162249-bf9dde6d0d2c
	github.com/juju/errgo v0.0.0-20140925100237-08cceb5d0b53
	github.com/prometheus/client_golang v1.5.1 // indirect
	github.com/prometheus/procfs v0.0.10 // indirect
	github.com/rogpeppe/go-internal v1.5.0 // indirect
	github.com/skratchdot/open-golang v0.0.0-20190402232053-79abb63cd66e // indirect
	github.com/spf13/afero v1.2.2
	github.com/spf13/cobra v0.0.6
	github.com/spf13/pflag v1.0.5
	golang.org/x/crypto v0.0.0-20200311171314-f7b00557c8c4 // indirect
	golang.org/x/sys v0.0.0-20200302150141-5c8b2ff67527 // indirect
	gomodules.xyz/jsonpatch/v2 v2.1.0 // indirect
	google.golang.org/appengine v1.6.5 // indirect
	gopkg.in/yaml.v2 v2.2.8
	honnef.co/go/tools v0.0.0-20190523083050-ea95bdfd59fc
	k8s.io/apiextensions-apiserver v0.17.4 // indirect
	k8s.io/apimachinery v0.17.4
	k8s.io/client-go v0.17.4
	k8s.io/cluster-bootstrap v0.17.4 // indirect
	k8s.io/kube-openapi v0.0.0-20200204173128-addea2498afe // indirect
	sigs.k8s.io/controller-runtime v0.5.1 // indirect
	sigs.k8s.io/structured-merge-diff v1.0.1 // indirect
)
