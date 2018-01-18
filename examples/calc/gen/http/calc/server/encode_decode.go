// Code generated with goa v2.0.0-wip, DO NOT EDIT.
//
// calc HTTP server encoders and decoders
//
// Command:
// $ goa gen goa.design/goa/examples/calc/design

package server

import (
	"context"
	"net/http"
	"strconv"

	goa "goa.design/goa"
	goahttp "goa.design/goa/http"
)

// EncodeAddResponse returns an encoder for responses returned by the calc add
// endpoint.
func EncodeAddResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(int)
		enc := encoder(ctx, w)
		body := res
		w.WriteHeader(http.StatusOK)
		return enc.Encode(body)
	}
}

// DecodeAddRequest returns a decoder for requests sent to the calc add
// endpoint.
func DecodeAddRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			a   int
			b   int
			err error

			params = mux.Vars(r)
		)
		{
			aRaw := params["a"]
			v, err2 := strconv.ParseInt(aRaw, 10, strconv.IntSize)
			if err2 != nil {
				err = goa.MergeErrors(err, goa.InvalidFieldTypeError("a", aRaw, "integer"))
			}
			a = int(v)
		}
		{
			bRaw := params["b"]
			v, err2 := strconv.ParseInt(bRaw, 10, strconv.IntSize)
			if err2 != nil {
				err = goa.MergeErrors(err, goa.InvalidFieldTypeError("b", bRaw, "integer"))
			}
			b = int(v)
		}
		if err != nil {
			return nil, err
		}

		return NewAddAddPayload(a, b), nil
	}
}

// EncodeAddedResponse returns an encoder for responses returned by the calc
// added endpoint.
func EncodeAddedResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(int)
		enc := encoder(ctx, w)
		body := res
		w.WriteHeader(http.StatusOK)
		return enc.Encode(body)
	}
}

// DecodeAddedRequest returns a decoder for requests sent to the calc added
// endpoint.
func DecodeAddedRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			query map[string][]int
			err   error
		)
		{
			queryRaw := r.URL.Query()
			if len(queryRaw) != 0 {
				query = make(map[string][]int, len(queryRaw))
				for key, valRaw := range queryRaw {
					var val []int
					{
						val = make([]int, len(valRaw))
						for i, rv := range valRaw {
							v, err2 := strconv.ParseInt(rv, 10, strconv.IntSize)
							if err2 != nil {
								err = goa.MergeErrors(err, goa.InvalidFieldTypeError("val", valRaw, "array of integers"))
							}
							val[i] = int(v)
						}
					}
					query[key] = val
				}
			}
		}
		if err != nil {
			return nil, err
		}

		return query, nil
	}
}
