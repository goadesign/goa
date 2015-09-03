package app

import "github.com/raphael/goa/codegen/code"

// ValidationWriter generate code that implements data validation.
type ValidationWriter struct {
	*code.Writer
	//HandlerTmpl *template.Template
}

// NewValidationWriter returns a validation code writer.
func NewValidationWriter(filename string) (*ValidationWriter, error) {
	cw, err := code.NewWriter(filename)
	if err != nil {
		return nil, err
	}
	//handlerTmpl, err := template.New("resource").Funcs(cw.FuncMap).Parse(handlerT)
	//if err != nil {
	//return nil, err
	//}
	w := ValidationWriter{
		Writer: cw,
		//HandlerTmpl: handlerTmpl,
	}
	return &w, nil
}

// Write writes the code for the validation.
func (w *ValidationWriter) Write(targetPack string) error {
	if err := w.Write(targetPack); err != nil {
		return err
	}
	return nil
}

/*// Regular expression used to validate RFC1035 hostnames*/
//var hostnameRegex = regexp.MustCompile(`^[[:alnum:]][[:alnum:]\-]{0,61}[[:alnum:]]|[[:alpha:]]$`)

//// Simple regular expression for IPv4 values, more rigorous checking is done via net.ParseIP
//var ipv4Regex = regexp.MustCompile(`^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$`)

//// validateFormat returns a validation function that validates the format of the given string
//// The format specification follows the json schema draft 4 validation extension.
//// see http://json-schema.org/latest/json-schema-validation.html#anchor105
//// Supported formats are:
//// - "date-time": RFC3339 date time value
//// - "email": RFC5322 email address
//// - "hostname": RFC1035 Internet host name
//// - "ipv4" and "ipv6": RFC2673 and RFC2373 IP address values
//// - "uri": RFC3986 URI value
//// - "mac": IEEE 802 MAC-48, EUI-48 or EUI-64 MAC address value
//// - "cidr": RFC4632 and RFC4291 CIDR notation IP address value
//// - "regexp": Regular expression syntax accepted by RE2
//func validateFormat(f string) func(name string, val interface{}) error {
//return func(name string, val interface{}) error {
//if val == nil {
//return nil
//}
//if sval, ok := val.(string); !ok {
//return fmt.Errorf("type of %s is invalid, got '%v' (%s), need string",
//name, val, reflect.TypeOf(val))
//} else {
//var err error
//switch strings.ToLower(f) {
//case "date-time":
//_, err = time.Parse(time.RFC3339, sval)
//case "email":
//_, err = mail.ParseAddress(sval)
//case "hostname":
//if !hostnameRegex.MatchString(sval) {
//err = fmt.Errorf("hostname value '%s' does not match %s",
//sval, hostnameRegex.String())
//}
//case "ipv4", "ipv6":
//ip := net.ParseIP(sval)
//if ip == nil {
//err = fmt.Errorf("\"%s\" is an invalid %s value", sval, f)
//}
//if f == "ipv4" {
//if !ipv4Regex.MatchString(sval) {
//err = fmt.Errorf("\"%s\" is an invalid ipv4 value", sval)
//}
//}
//case "uri":
//_, err = url.ParseRequestURI(sval)
//case "mac":
//_, err = net.ParseMAC(sval)
//case "cidr":
//_, _, err = net.ParseCIDR(sval)
//case "regexp":
//_, err = regexp.Compile(sval)
//default:
//err = fmt.Errorf("unknown validation format '%s'", f)
//}
//if err == nil {
//return nil
//}
//return fmt.Errorf("invalid %s value, %s", name, err)
//}
//}
//}

//// validateIntMaxValue returns a validation function that checks whether given value is a int that
//// is lesser than max.
//func validateIntMaximum(max int) Validation {
//return func(name string, val interface{}) error {
//if val == nil {
//return nil
//}
//if ival, ok := val.(int); !ok {
//return fmt.Errorf("type of %s is invalid, got '%v', need integer",
//name, val)
//} else if ival > max {
//return fmt.Errorf("%v is an invalid %s value: maximum allowed is %v",
//ival, name, max)
//}
//return nil
//}
//}

//// validateIntMinValue returns a validation function that checks whether given value is a int that
//// is greater than min.
//func validateIntMinimum(min int) Validation {
//return func(name string, val interface{}) error {
//if val == nil {
//return nil
//}
//if ival, ok := val.(int); !ok {
//return fmt.Errorf("type of %s is invalid, got '%v', need integer",
//name, val)
//} else if ival < min {
//return fmt.Errorf("%v is an invalid %s value: minimum allowed is %v",
//ival, name, min)
//}
//return nil
//}
//}

//// validateMinLength returns a validation function that checks whether given string or array has
//// at least the number of given characters or elements.
//func validateMinLength(min int) Validation {
//return func(name string, val interface{}) error {
//if val == nil {
//return nil
//}
//if sval, ok := val.(string); ok {
//if len(sval) < min {
//return fmt.Errorf("%v (%d characters) is an invalid %s value: minimum allowed length is %v",
//sval, len(sval), name, min)
//}
//} else {
//k := reflect.TypeOf(val).Kind()
//if k == reflect.Slice || k == reflect.Array {
//v := reflect.ValueOf(val)
//if v.Len() < min {
//return fmt.Errorf("%v (%d items) is an invalid %s value: minimum allowed length is %v",
//v.Interface(), v.Len(), name, min)
//}
//} else {
//return fmt.Errorf("'%v' is an invalid %s value, need string or array",
//val, name)
//}
//}
//return nil
//}
//}

//// validateMaxLength returns a validation function that checks whether given string or array has
//// at most the number of given characters or elements.
//func validateMaxLength(max int) Validation {
//return func(name string, val interface{}) error {
//if val == nil {
//return nil
//}
//if sval, ok := val.(string); ok {
//if len(sval) > max {
//return fmt.Errorf("%v (%d characters) is an invalid %s value: maximum allowed length is %v",
//sval, len(sval), name, max)
//}
//} else {
//k := reflect.TypeOf(val).Kind()
//if k == reflect.Slice || k == reflect.Array {
//v := reflect.ValueOf(val)
//if v.Len() > max {
//return fmt.Errorf("%v (%d items) is an invalid %s value: maximum allowed length is %v",
//v.Interface(), v.Len(), name, max)
//}
//} else {
//return fmt.Errorf("'%v' is an invalid %s value, need string or array",
//val, name)
//}
//}
//return nil
//}
//}

//// validateEnum returns a validation function that checks whether given value is one of the
//// valid values.
//func validateEnum(valid []interface{}) Validation {
//return func(name string, val interface{}) error {
//ok := false
//for _, v := range valid {
//if v == val {
//ok = true
//break
//}
//}
//if !ok {
//sValid := make([]string, len(valid))
//for i, v := range valid {
//sValid[i] = fmt.Sprintf("%v", v)
//}
//return fmt.Errorf("\"%v\" is an invalid %s value: allowed values are %s",
//val, name, strings.Join(sValid, ", "))
//}
//return nil
//}
/*}*/
