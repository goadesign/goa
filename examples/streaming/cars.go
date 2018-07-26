package cars

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	carssvc "goa.design/goa/examples/streaming/gen/cars"
)

// cars service example implementation.

// The example methods log the requests and return zero values.
type carsSvc struct {
	logger *log.Logger
}

// NewCars returns the cars service implementation.
func NewCars(logger *log.Logger) carssvc.Service {
	return &carsSvc{logger}
}

// Login creates a valid JWT given valid credentials. Login returns an error of
// type carsSvc.Unauthorized if the credentials are invalid.
func (s *carsSvc) Login(ctx context.Context, p *carssvc.LoginPayload) (string, error) {
	// create JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"nbf":    time.Date(2015, 10, 10, 12, 0, 0, 0, time.UTC).Unix(),
		"iat":    time.Now().Unix(),
		"scopes": []string{"stream:read"},
	})

	s.logger.Printf("user '%s' logged in", p.User)

	// note that if "SignedString" returns an error then it is returned as
	// an internal error to the client
	return token.SignedString(Key)
}

// Lists car models by body style.
func (s *carsSvc) List(ctx context.Context, p *carssvc.ListPayload, stream carssvc.ListServerStream) (err error) {
	for _, c := range modelsByStyle[p.Style] {
		if err := stream.Send(c); err != nil {
			return err
		}
	}
	return stream.Close()
}

// Add car models.
func (s *carsSvc) Add(ctx context.Context, p *carssvc.AddPayload, stream carssvc.AddServerStream) (err error) {
	var res carssvc.StoredCarCollection
	for {
		sp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if car := sp.Car; car != nil {
			sc := &carssvc.StoredCar{
				Make:      car.Make,
				Model:     car.Model,
				BodyStyle: car.BodyStyle,
			}
			modelsByStyle[car.BodyStyle] = append(modelsByStyle[car.BodyStyle], sc)
			res = append(res, sc)
		}
	}
	return stream.SendAndClose(res)
}

// Update car models.
func (s *carsSvc) Update(ctx context.Context, p *carssvc.UpdatePayload, stream carssvc.UpdateServerStream) (err error) {
	for {
		cars, err := stream.Recv()
		fmt.Println(fmt.Sprintf("%#v", cars))
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		var res carssvc.StoredCarCollection
		for _, car := range cars {
			fmt.Println(fmt.Sprintf("%#v", car))
			sc := &carssvc.StoredCar{
				Make:      car.Make,
				Model:     car.Model,
				BodyStyle: car.BodyStyle,
			}
			modelsByStyle[car.BodyStyle] = append(modelsByStyle[car.BodyStyle], sc)
			fmt.Println(fmt.Sprintf("%#v", sc))
			res = append(res, sc)
		}
		if err := stream.Send(res); err != nil {
			return err
		}
	}
	return stream.Close()
}

var modelsByStyle = map[string][]*carssvc.StoredCar{
	"sedan": []*carssvc.StoredCar{
		&carssvc.StoredCar{Make: "Acura", Model: "TLX", BodyStyle: "sedan"},
		&carssvc.StoredCar{Make: "Audi", Model: "A4", BodyStyle: "sedan"},
		&carssvc.StoredCar{Make: "BMW", Model: "M3", BodyStyle: "sedan"},
		&carssvc.StoredCar{Make: "Chevrolet", Model: "Cruze", BodyStyle: "sedan"},
		&carssvc.StoredCar{Make: "Ford", Model: "Focus", BodyStyle: "sedan"},
		&carssvc.StoredCar{Make: "Honda", Model: "Accord", BodyStyle: "sedan"},
		&carssvc.StoredCar{Make: "Hyundai", Model: "Accent", BodyStyle: "sedan"},
		&carssvc.StoredCar{Make: "Infiniti", Model: "Q50", BodyStyle: "sedan"},
		&carssvc.StoredCar{Make: "Kia", Model: "Rio", BodyStyle: "sedan"},
		&carssvc.StoredCar{Make: "Lexus", Model: "ES", BodyStyle: "sedan"},
		&carssvc.StoredCar{Make: "Mazda", Model: "6", BodyStyle: "sedan"},
		&carssvc.StoredCar{Make: "Mercedes", Model: "C-Class", BodyStyle: "sedan"},
		&carssvc.StoredCar{Make: "Nissan", Model: "Altima", BodyStyle: "sedan"},
		&carssvc.StoredCar{Make: "Porsche", Model: "Panamera", BodyStyle: "sedan"},
		&carssvc.StoredCar{Make: "Subaru", Model: "Impreza", BodyStyle: "sedan"},
		&carssvc.StoredCar{Make: "Volkswagen", Model: "Passat", BodyStyle: "sedan"},
	},
	"hatchback": []*carssvc.StoredCar{
		&carssvc.StoredCar{Make: "Acura", Model: "MDX", BodyStyle: "hatchback"},
		&carssvc.StoredCar{Make: "Audi", Model: "Q3", BodyStyle: "hatchback"},
		&carssvc.StoredCar{Make: "BMW", Model: "X3", BodyStyle: "hatchback"},
		&carssvc.StoredCar{Make: "Chevrolet", Model: "Equinox", BodyStyle: "hatchback"},
		&carssvc.StoredCar{Make: "Ford", Model: "Escape", BodyStyle: "hatchback"},
		&carssvc.StoredCar{Make: "Honda", Model: "CRV", BodyStyle: "hatchback"},
		&carssvc.StoredCar{Make: "Hyundai", Model: "Santa Fe", BodyStyle: "hatchback"},
		&carssvc.StoredCar{Make: "Infiniti", Model: "QX30", BodyStyle: "hatchback"},
		&carssvc.StoredCar{Make: "Kia", Model: "Sorento", BodyStyle: "hatchback"},
		&carssvc.StoredCar{Make: "Lexus", Model: "NX", BodyStyle: "hatchback"},
		&carssvc.StoredCar{Make: "Mazda", Model: "CX5", BodyStyle: "hatchback"},
		&carssvc.StoredCar{Make: "Mercedes", Model: "GLA-Class", BodyStyle: "hatchback"},
		&carssvc.StoredCar{Make: "Nissan", Model: "Rogue", BodyStyle: "hatchback"},
		&carssvc.StoredCar{Make: "Porsche", Model: "Cayenne", BodyStyle: "hatchback"},
		&carssvc.StoredCar{Make: "Subaru", Model: "Outback", BodyStyle: "hatchback"},
		&carssvc.StoredCar{Make: "Volkswagen", Model: "Golf", BodyStyle: "hatchback"},
	},
}
