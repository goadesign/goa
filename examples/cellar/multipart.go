package cellar

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"

	storagec "goa.design/goa/examples/cellar/gen/http/storage/client"
	storages "goa.design/goa/examples/cellar/gen/http/storage/server"
	storage "goa.design/goa/examples/cellar/gen/storage"
)

// StorageMultiAddDecoderFunc implements the multipart decoder for service
// "storage" endpoint "multi_add". The decoder must populate the argument p
// after encoding.
func StorageMultiAddDecoderFunc(mr *multipart.Reader, p *[]*storage.Bottle) error {
	var bottles []*storages.BottleRequestBody
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to load part: %s", err)
		}
		dec := json.NewDecoder(part)
		var bottle storages.BottleRequestBody
		if err := dec.Decode(&bottle); err != nil {
			return fmt.Errorf("failed to decode part: %s", err)
		}
		bottles = append(bottles, &bottle)
	}
	*p = storages.NewMultiAddBottle(bottles)
	return nil
}

// StorageMultiAddEncoderFunc implements the multipart encoder for service
// "storage" endpoint "multi_add".
func StorageMultiAddEncoderFunc(mw *multipart.Writer, p []*storage.Bottle) error {
	bottles := storagec.NewBottleRequestBody(p)
	for _, bottle := range bottles {
		b, err := json.Marshal(bottle)
		if err != nil {
			return err
		}
		if err := mw.WriteField("bottle", string(b)); err != nil {
			return err
		}
	}
	return nil
}

// StorageMultiUpdateDecoderFunc implements the multipart decoder for service
// "storage" endpoint "multi_update". The decoder must populate the argument p
// after encoding.
func StorageMultiUpdateDecoderFunc(mr *multipart.Reader, p **storage.MultiUpdatePayload) error {
	var bottles []*storages.BottleRequestBody
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to load part: %s", err)
		}
		dec := json.NewDecoder(part)
		var bottle storages.BottleRequestBody
		if err := dec.Decode(&bottle); err != nil {
			return fmt.Errorf("failed to decode part: %s", err)
		}
		bottles = append(bottles, &bottle)
	}
	reqBody := storages.MultiUpdateRequestBody{Bottles: bottles}
	*p = storages.NewMultiUpdatePayload(&reqBody, []string{})
	return nil
}

// StorageMultiUpdateEncoderFunc implements the multipart encoder for service
// "storage" endpoint "multi_update".
func StorageMultiUpdateEncoderFunc(mw *multipart.Writer, p *storage.MultiUpdatePayload) error {
	reqBody := storagec.NewMultiUpdateRequestBody(p)
	for _, bottle := range reqBody.Bottles {
		b, err := json.Marshal(bottle)
		if err != nil {
			return err
		}
		if err := mw.WriteField("bottle", string(b)); err != nil {
			return err
		}
	}
	return nil
}
