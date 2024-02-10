package adapter

import (
	"github.com/kcloutie/event-reactor/pkg/config"
	"github.com/kcloutie/event-reactor/pkg/reactor/email"
	"github.com/kcloutie/event-reactor/pkg/reactor/powershell"
	"github.com/kcloutie/event-reactor/pkg/reactor/webex"
	"github.com/kcloutie/event-reactor/pkg/reactor/webhook"

	"github.com/kcloutie/event-reactor/pkg/reactor"
	"github.com/kcloutie/event-reactor/pkg/reactor/githubcomment"
	"go.uber.org/zap"
)

func GetReactorNewFunctions(loadTestReactor bool) map[string]func(log *zap.Logger, reactorConfig config.ReactorConfig) reactor.ReactorInterface {
	results := map[string]func(log *zap.Logger, reactorConfig config.ReactorConfig) reactor.ReactorInterface{}

	githubReactor := githubcomment.New()
	results[githubReactor.GetName()] = func(log *zap.Logger, reactorConfig config.ReactorConfig) reactor.ReactorInterface {
		reactor := githubcomment.New()
		reactor.SetLogger(log)
		reactor.SetReactor(reactorConfig)
		return reactor
	}

	emailReactor := email.New()
	results[emailReactor.GetName()] = func(log *zap.Logger, reactorConfig config.ReactorConfig) reactor.ReactorInterface {
		reactor := email.New()
		reactor.SetLogger(log)
		reactor.SetReactor(reactorConfig)
		return reactor
	}

	pwshReactor := powershell.New()
	results[pwshReactor.GetName()] = func(log *zap.Logger, reactorConfig config.ReactorConfig) reactor.ReactorInterface {
		reactor := powershell.New()
		reactor.SetLogger(log)
		reactor.SetReactor(reactorConfig)
		return reactor
	}

	webhookReactor := webhook.New()
	results[webhookReactor.GetName()] = func(log *zap.Logger, reactorConfig config.ReactorConfig) reactor.ReactorInterface {
		reactor := webhook.New()
		reactor.SetLogger(log)
		reactor.SetReactor(reactorConfig)
		return reactor
	}

	webexReactor := webex.New()
	results[webexReactor.GetName()] = func(log *zap.Logger, reactorConfig config.ReactorConfig) reactor.ReactorInterface {
		reactor := webex.New()
		reactor.SetLogger(log)
		reactor.SetReactor(reactorConfig)
		return reactor
	}

	if loadTestReactor {
		// Add test reactor
		testReactor := reactor.NewTestReactor()
		results[testReactor.GetName()] = func(log *zap.Logger, reactorConfig config.ReactorConfig) reactor.ReactorInterface {
			reactor := reactor.NewTestReactor()
			reactor.SetLogger(log)
			reactor.SetReactor(reactorConfig)
			return reactor
		}
	}

	return results
}
