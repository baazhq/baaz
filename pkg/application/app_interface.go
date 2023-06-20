package application

type App interface {
	ReconcileApplicationDeployer() error
}
