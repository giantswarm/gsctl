package request

type Request struct {
	Cluster Cluster
}

type Cluster struct {
	Scaling Scaling
}

type Scaling struct {
	Max int64
	Min int64
}
