package cellar

import (
	"context"
	"log"

	sommelier "goa.design/goa/examples/cellar/gen/sommelier"
)

// sommelier service example implementation.
// The example methods log the requests and return zero values.
type sommelierSvc struct {
	logger *log.Logger
}

// NewSommelier returns the sommelier service implementation.
func NewSommelier(logger *log.Logger) sommelier.Service {
	return &sommelierSvc{logger}
}

// Pick implements pick.
func (s *sommelierSvc) Pick(ctx context.Context, p *sommelier.Criteria) (res sommelier.StoredBottleCollection, err error) {
	if p.Name == nil && len(p.Varietal) == 0 && p.Winery == nil {
		return nil, sommelier.NoCriteria("must specify a name or one or more varietals or a winery")
	}
	// TBD: implement lookup return sommeliner.NoMatch if empty
	s.logger.Print("sommelier.pick")
	return res, nil
}
