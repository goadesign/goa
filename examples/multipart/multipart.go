package api

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"

	"goa.design/goa/examples/multipart/gen/http/resume/client"
	"goa.design/goa/examples/multipart/gen/http/resume/server"
	resume "goa.design/goa/examples/multipart/gen/resume"
)

// ResumeAddDecoderFunc implements the multipart decoder for service "resume"
// endpoint "add". The decoder must populate the argument p after encoding.
func ResumeAddDecoderFunc(mr *multipart.Reader, p *[]*resume.Resume) error {
	var resumes []*server.ResumeRequestBody
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to load part: %s", err)
		}
		dec := json.NewDecoder(part)
		var r server.ResumeRequestBody
		if err := dec.Decode(&r); err != nil {
			return fmt.Errorf("failed to decode part: %s", err)
		}
		resumes = append(resumes, &r)
	}
	*p = server.NewAddResume(resumes)
	return nil
}

// ResumeAddEncoderFunc implements the multipart encoder for service "resume"
// endpoint "add".
func ResumeAddEncoderFunc(mw *multipart.Writer, p []*resume.Resume) error {
	resumes := client.NewResumeRequestBody(p)
	for _, resume := range resumes {
		b, err := json.Marshal(resume)
		if err != nil {
			return err
		}
		if err := mw.WriteField("resume", string(b)); err != nil {
			return err
		}
	}
	return nil
}
