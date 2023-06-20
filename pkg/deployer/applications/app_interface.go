package applications

type App interface {
	ReconcileApplication() error
}
