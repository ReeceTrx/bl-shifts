package notifiers

type Notifier interface {
	Send(codes []string) error
}
