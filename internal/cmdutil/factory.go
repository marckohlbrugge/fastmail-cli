package cmdutil

import (
	"github.com/marckohlbrugge/fastmail-cli/internal/auth"
	"github.com/marckohlbrugge/fastmail-cli/internal/iostreams"
	"github.com/marckohlbrugge/fastmail-cli/internal/jmap"
)

// Factory provides dependencies for commands.
type Factory struct {
	IOStreams   *iostreams.IOStreams
	TokenSource *auth.TokenSource

	// Lazy-initialized JMAP client
	jmapClient *jmap.Client
}

// NewFactory creates a new Factory with default dependencies.
func NewFactory() *Factory {
	return &Factory{
		IOStreams:   iostreams.System(),
		TokenSource: auth.NewTokenSource(),
	}
}

// JMAPClient returns the JMAP client, initializing it if necessary.
func (f *Factory) JMAPClient() (*jmap.Client, error) {
	if f.jmapClient != nil {
		return f.jmapClient, nil
	}

	token, err := f.TokenSource.GetToken()
	if err != nil {
		return nil, err
	}

	f.jmapClient = jmap.NewClient(token)
	return f.jmapClient, nil
}
