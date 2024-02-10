package listener

import (
	"github.com/kcloutie/event-reactor/pkg/listener/generic"
	"github.com/kcloutie/event-reactor/pkg/listener/pubsub"
)

func GetListeners() []ListenerInterface {
	listeners := []ListenerInterface{}
	listeners = append(listeners, generic.New())
	listeners = append(listeners, pubsub.New())

	return listeners
}
