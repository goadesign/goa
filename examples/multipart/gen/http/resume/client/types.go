// Code generated by goa v2.0.0-wip, DO NOT EDIT.
//
// resume HTTP client types
//
// Command:
// $ goa gen goa.design/goa/examples/multipart/design -o
// $(GOPATH)/src/goa.design/goa/examples/multipart

package client

import (
	goa "goa.design/goa"
	resume "goa.design/goa/examples/multipart/gen/resume"
	resumeviews "goa.design/goa/examples/multipart/gen/resume/views"
)

// ListResponseBody is the type of the "resume" service "list" endpoint HTTP
// response body.
type ListResponseBody []*StoredResumeResponse

// StoredResumeResponse is used to define fields on response body types.
type StoredResumeResponse struct {
	// ID of the resume
	ID *int `form:"id,omitempty" json:"id,omitempty" xml:"id,omitempty"`
	// Time when resume was created
	CreatedAt *string `form:"created_at,omitempty" json:"created_at,omitempty" xml:"created_at,omitempty"`
	// Name in the resume
	Name *string `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`
	// Experience section in the resume
	Experience []*ExperienceResponse `form:"experience,omitempty" json:"experience,omitempty" xml:"experience,omitempty"`
	// Education section in the resume
	Education []*EducationResponse `form:"education,omitempty" json:"education,omitempty" xml:"education,omitempty"`
}

// ExperienceResponse is used to define fields on response body types.
type ExperienceResponse struct {
	// Name of the company
	Company *string `form:"company,omitempty" json:"company,omitempty" xml:"company,omitempty"`
	// Name of the role in the company
	Role *string `form:"role,omitempty" json:"role,omitempty" xml:"role,omitempty"`
	// Duration (in years) in the company
	Duration *int `form:"duration,omitempty" json:"duration,omitempty" xml:"duration,omitempty"`
}

// EducationResponse is used to define fields on response body types.
type EducationResponse struct {
	// Name of the institution
	Institution *string `form:"institution,omitempty" json:"institution,omitempty" xml:"institution,omitempty"`
	// Major name
	Major *string `form:"major,omitempty" json:"major,omitempty" xml:"major,omitempty"`
}

// ResumeRequestBody is used to define fields on request body types.
type ResumeRequestBody struct {
	// Name in the resume
	Name string `form:"name" json:"name" xml:"name"`
	// Experience section in the resume
	Experience []*ExperienceRequestBody `form:"experience,omitempty" json:"experience,omitempty" xml:"experience,omitempty"`
	// Education section in the resume
	Education []*EducationRequestBody `form:"education,omitempty" json:"education,omitempty" xml:"education,omitempty"`
}

// ExperienceRequestBody is used to define fields on request body types.
type ExperienceRequestBody struct {
	// Name of the company
	Company string `form:"company" json:"company" xml:"company"`
	// Name of the role in the company
	Role string `form:"role" json:"role" xml:"role"`
	// Duration (in years) in the company
	Duration int `form:"duration" json:"duration" xml:"duration"`
}

// EducationRequestBody is used to define fields on request body types.
type EducationRequestBody struct {
	// Name of the institution
	Institution string `form:"institution" json:"institution" xml:"institution"`
	// Major name
	Major string `form:"major" json:"major" xml:"major"`
}

// NewResumeRequestBody builds the HTTP request body from the payload of the
// "add" endpoint of the "resume" service.
func NewResumeRequestBody(p []*resume.Resume) []*ResumeRequestBody {
	body := make([]*ResumeRequestBody, len(p))
	for i, val := range p {
		body[i] = &ResumeRequestBody{
			Name: val.Name,
		}
		if val.Experience != nil {
			body[i].Experience = make([]*ExperienceRequestBody, len(val.Experience))
			for j, val := range val.Experience {
				body[i].Experience[j] = &ExperienceRequestBody{
					Company:  val.Company,
					Role:     val.Role,
					Duration: val.Duration,
				}
			}
		}
		if val.Education != nil {
			body[i].Education = make([]*EducationRequestBody, len(val.Education))
			for j, val := range val.Education {
				body[i].Education[j] = &EducationRequestBody{
					Institution: val.Institution,
					Major:       val.Major,
				}
			}
		}
	}
	return body
}

// NewListStoredResumeCollectionOK builds a "resume" service "list" endpoint
// result from a HTTP "OK" response.
func NewListStoredResumeCollectionOK(body ListResponseBody) resumeviews.StoredResumeCollectionView {
	v := make([]*resumeviews.StoredResumeView, len(body))
	for i, val := range body {
		v[i] = &resumeviews.StoredResumeView{
			ID:        val.ID,
			CreatedAt: val.CreatedAt,
			Name:      val.Name,
		}
		v[i].Experience = make([]*resumeviews.ExperienceView, len(val.Experience))
		for j, val := range val.Experience {
			v[i].Experience[j] = &resumeviews.ExperienceView{
				Company:  val.Company,
				Role:     val.Role,
				Duration: val.Duration,
			}
		}
		v[i].Education = make([]*resumeviews.EducationView, len(val.Education))
		for j, val := range val.Education {
			v[i].Education[j] = &resumeviews.EducationView{
				Institution: val.Institution,
				Major:       val.Major,
			}
		}
	}
	return v
}

// ValidateStoredResumeResponse runs the validations defined on
// StoredResumeResponse
func ValidateStoredResumeResponse(body *StoredResumeResponse) (err error) {
	if body.ID == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("id", "body"))
	}
	if body.Name == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("name", "body"))
	}
	if body.Experience == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("experience", "body"))
	}
	if body.Education == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("education", "body"))
	}
	if body.CreatedAt == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("created_at", "body"))
	}
	return
}

// ValidateExperienceResponse runs the validations defined on ExperienceResponse
func ValidateExperienceResponse(body *ExperienceResponse) (err error) {
	if body.Company == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("company", "body"))
	}
	if body.Role == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("role", "body"))
	}
	if body.Duration == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("duration", "body"))
	}
	return
}

// ValidateEducationResponse runs the validations defined on EducationResponse
func ValidateEducationResponse(body *EducationResponse) (err error) {
	if body.Institution == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("institution", "body"))
	}
	if body.Major == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("major", "body"))
	}
	return
}
