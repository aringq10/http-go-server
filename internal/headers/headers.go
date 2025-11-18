package headers

import (
    "fmt"
    "regexp"
    "strings"
)

const crlf = "\r\n"
const fieldNameRegEx = "^[A-Za-z0-9,#$%&'*+.^_`|~-]+$"

type Headers map[string]string

func NewHeaders() Headers {
    return make(Headers)
}

func (h Headers) Get(key string) string {
    key = strings.ToLower(key)
    return h[key]
}

func (h Headers) Set(key string, value string) {
    key = strings.ToLower(key)
    if h[key] != "" {
        value = h[key] + ", " + value
    }
    h[key] = value
}

func (h Headers) Replace(key string, value string) {
    key = strings.ToLower(key)
    h[key] = value
}

func (h Headers) Remove(key string) {
    key = strings.ToLower(key)
    delete(h, key)
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
    tmpSplice := strings.Split(string(data), crlf)

    if len(tmpSplice) == 1 {
        return 0, false, nil
    }
    headerString := tmpSplice[0]

    n = len(headerString) + 2
    err = nil

    if headerString == "" {
        done = true
        return
    }

    tmpSplice = strings.SplitN(headerString, ": ", 2)
    if len(tmpSplice) == 1 {
        return 0, false, fmt.Errorf("field line is missing ':' or a space after ':' - \"%v\"", headerString)
    }
    fieldName := tmpSplice[0]
    fieldValue := tmpSplice[1]

    tmpSplice = strings.Fields(fieldName)
    if len(tmpSplice) != 1 || fieldName[len(fieldName) - 1] == ' '{
        return 0, false, fmt.Errorf("invalid field name - \"%v\"", headerString)
    }
    fieldName = tmpSplice[0]

    re := regexp.MustCompile(fieldNameRegEx)
    if re.MatchString(fieldName) == false {
        return 0, false, fmt.Errorf("field name contains invalid characters - \"%v\"", headerString)
    }

    fieldValue = strings.TrimSpace(fieldValue)

    h.Set(fieldName, fieldValue)

    return
}
