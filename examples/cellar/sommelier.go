package cellar

import (
	"context"

	sommelier "goa.design/goa/examples/cellar/gen/sommelier"
	goalog "goa.design/goa/logging"
)

// sommelier service example implementation.
// The example methods log the requests and return zero values.
type sommelierSvc struct {
	logger goalog.Logger
}

// Required for compatibility with Service interface
func (s *sommelierSvc) GetLogger() goalog.Logger {
	return s.logger
}

// NewSommelier returns the sommelier service implementation.
func NewSommelier(logger goalog.Logger) sommelier.Service {
	return &sommelierSvc{logger: logger}
}

// Pick implements pick.
func (s *sommelierSvc) Pick(ctx context.Context, p *sommelier.Criteria) (res sommelier.StoredBottleCollection, err error) {
	if p.Name == nil && len(p.Varietal) == 0 && p.Winery == nil {
		return nil, sommelier.NoCriteria("must specify a name or one or more varietals or a winery")
	}
	// TBD: implement lookup return sommeliner.NoMatch if empty
	s.logger.Debug("sommelier.pick")
	return res, nil
}
