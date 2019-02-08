package api

import (
	"context"
	"log"
	"time"

	resume "goa.design/goa/examples/multipart/gen/resume"
)

// stored is a simple in-memory storage for the resumes.
var stored resume.StoredResumeCollection

// resume service example implementation.
// The example methods log the requests and return zero values.
type resumesrvc struct {
	logger *log.Logger
}

// NewResume returns the resume service implementation.
func NewResume(logger *log.Logger) resume.Service {
	return &resumesrvc{logger}
}

// List all stored resumes
func (s *resumesrvc) List(ctx context.Context) (res resume.StoredResumeCollection, err error) {
	res = stored
	return
}

// Add n number of resumes and return their IDs. This is a multipart request
// and each part has field name 'resume' and contains the encoded resume to be
// added.
func (s *resumesrvc) Add(ctx context.Context, p []*resume.Resume) (res []int, err error) {
	for _, r := range p {
		if resumeExists(r.Name) {
			continue
		}
		sr := resume.StoredResume{
			ID:         len(stored) + 1,
			Name:       r.Name,
			Experience: r.Experience,
			Education:  r.Education,
			CreatedAt:  time.Now().Format(time.RFC3339),
		}
		stored = append(stored, &sr)
		res = append(res, sr.ID)
	}
	return
}

func resumeExists(name string) bool {
	for _, r := range stored {
		if r.Name == name {
			return true
		}
	}
	return false
}
