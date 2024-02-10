package settings

const (
	CliBinaryName               = "er"
	DebugModeLoggerEnvVar       = "EVENT_REACTOR_DEBUG"
	PubSubEndpoint              = "pubsub"
	GoTemplateDefaultDelimLeft  = "{{"
	GoTemplateDefaultDelimRight = "}}"
	SignatureHeader             = "X-Event-Reactor-Signature"
	WebexApiUrlDefault          = "https://api.ciscospark.com/v1/messages"
)

var (
	RootOptions      = RootFlags{}
	DebugModeEnabled = false
	IsQuiet          = false
)

type RootFlags struct {
	NoColor bool
}
